// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php


// Package man implements the Matrix protocol.
package man

import (
	"errors"
	"fmt"
	"math/big"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/matrix/go-matrix/random"

	"github.com/matrix/go-matrix/ca"

	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/reelection"

	"github.com/matrix/go-matrix/accounts"
	"github.com/matrix/go-matrix/accounts/signhelper"
	"github.com/matrix/go-matrix/blkgenor"
	"github.com/matrix/go-matrix/blkverify/blkverify"
	"github.com/matrix/go-matrix/broadcastTx"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/common/hexutil"
	"github.com/matrix/go-matrix/consensus"
	"github.com/matrix/go-matrix/consensus/clique"
	"github.com/matrix/go-matrix/consensus/manash"
	"github.com/matrix/go-matrix/core"
	"github.com/matrix/go-matrix/core/bloombits"
	"github.com/matrix/go-matrix/core/rawdb"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/core/vm"
	"github.com/matrix/go-matrix/depoistInfo"
	"github.com/matrix/go-matrix/man/downloader"
	"github.com/matrix/go-matrix/man/filters"
	"github.com/matrix/go-matrix/man/gasprice"
	"github.com/matrix/go-matrix/mandb"
	"github.com/matrix/go-matrix/event"
	"github.com/matrix/go-matrix/internal/manapi"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/miner"
	"github.com/matrix/go-matrix/msgsend"
	"github.com/matrix/go-matrix/pod"
	"github.com/matrix/go-matrix/p2p"
	"github.com/matrix/go-matrix/params"
	"github.com/matrix/go-matrix/rlp"
	"github.com/matrix/go-matrix/rpc"

	"sync"

	"github.com/matrix/go-matrix/leaderelect"
	"github.com/matrix/go-matrix/olconsensus"
)

var MsgCenter *mc.Center

type LesServer interface {
	Start(srvr *p2p.Server)
	Stop()
	Protocols() []p2p.Protocol
	SetBloomBitsIndexer(bbIndexer *core.ChainIndexer)
}

// Matrix implements the Matrix full node service.
type Matrix struct {
	config      *Config
	chainConfig *params.ChainConfig

	// Channel for shutting down the service
	shutdownChan chan bool // Channel for shutting down the Matrix

	// Handlers
	txPool          *core.TxPool
	blockchain      *core.BlockChain
	protocolManager *ProtocolManager
	lesServer       LesServer

	// DB interfaces
	chainDb mandb.Database // Block chain database

	eventMux       *event.TypeMux
	engine         consensus.Engine
	accountManager *accounts.Manager

	bloomRequests chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	bloomIndexer  *core.ChainIndexer             // Bloom indexer operating during block imports

	APIBackend *ManAPIBackend

	miner     *miner.Miner
	gasPrice  *big.Int
	manbase common.Address

	networkId     uint64
	netRPCService *manapi.PublicNetAPI

	broadTx *broadcastTx.BroadCast //YY

	//algorithm
	ca         *ca.Identity //node传进来的
	msgcenter  *mc.Center   //node传进来的
	hd         *msgsend.HD  //node传进来的
	signHelper *signhelper.SignHelper

	reelection   *reelection.ReElection //换届服务
	random       *random.Random
	topNode      *olconsensus.TopNodeService
	blockgen     *blkgenor.BlockGenor
	blockVerify  *blkverify.BlockVerify
	leaderServer *leaderelect.LeaderIdentity

	lock sync.RWMutex // Protects the variadic fields (e.g. gas price and manbase)
}

func (s *Matrix) AddLesServer(ls LesServer) {
	s.lesServer = ls
	ls.SetBloomBitsIndexer(s.bloomIndexer)
}

// New creates a new Matrix object (including the
// initialisation of the common Matrix object)
func New(ctx *pod.ServiceContext, config *Config) (*Matrix, error) {
	if config.SyncMode == downloader.LightSync {
		return nil, errors.New("can't run man.Matrix in light sync mode, use les.LightMatrix")
	}
	if !config.SyncMode.IsValid() {
		return nil, fmt.Errorf("invalid sync mode %d", config.SyncMode)
	}
	chainDb, err := CreateDB(ctx, config, "chaindata")
	if err != nil {
		return nil, err
	}
	chainConfig, genesisHash, genesisErr := core.SetupGenesisBlock(chainDb, config.Genesis)
	if _, ok := genesisErr.(*params.ConfigCompatError); genesisErr != nil && !ok {
		return nil, genesisErr
	}
	log.Info("Initialised chain configuration", "config", chainConfig)

	man := &Matrix{
		config:         config,
		chainDb:        chainDb,
		chainConfig:    chainConfig,
		eventMux:       ctx.EventMux,
		accountManager: ctx.AccountManager,
		ca:             ctx.Ca,
		msgcenter:      ctx.MsgCenter,
		hd:             ctx.HD,
		signHelper:     ctx.SignHelper,

		engine:        CreateConsensusEngine(ctx, &config.Manash, chainConfig, chainDb),
		shutdownChan:  make(chan bool),
		networkId:     config.NetworkId,
		gasPrice:      config.GasPrice,
		manbase:     config.Manerbase,
		bloomRequests: make(chan chan *bloombits.Retrieval),
		bloomIndexer:  NewBloomIndexer(chainDb, params.BloomBitsBlocks),
	}
	log.Info("Initialising Matrix protocol", "versions", ProtocolVersions, "network", config.NetworkId)

	if !config.SkipBcVersionCheck {
		bcVersion := rawdb.ReadDatabaseVersion(chainDb)
		if bcVersion != core.BlockChainVersion && bcVersion != 0 {
			return nil, fmt.Errorf("Blockchain DB version mismatch (%d / %d). Run gman upgradedb.\n", bcVersion, core.BlockChainVersion)
		}
		rawdb.WriteDatabaseVersion(chainDb, core.BlockChainVersion)
	}
	var (
		vmConfig    = vm.Config{EnablePreimageRecording: config.EnablePreimageRecording}
		cacheConfig = &core.CacheConfig{Disabled: config.NoPruning, TrieNodeLimit: config.TrieCache, TrieTimeLimit: config.TrieTimeout}
	)
	man.blockchain, err = core.NewBlockChain(chainDb, cacheConfig, man.chainConfig, man.engine, vmConfig)
	if err != nil {
		return nil, err
	}
	// Rewind the chain in case of an incompatible config upgrade.
	if compat, ok := genesisErr.(*params.ConfigCompatError); ok {
		log.Warn("Rewinding chain to upgrade configuration", "err", compat)
		man.blockchain.SetHead(compat.RewindTo)
		rawdb.WriteChainConfig(chainDb, genesisHash, chainConfig)
	}
	man.bloomIndexer.Start(man.blockchain)

	if config.TxPool.Journal != "" {
		config.TxPool.Journal = ctx.ResolvePath(config.TxPool.Journal)
	}
	man.txPool = core.NewTxPool(config.TxPool, man.chainConfig, man.blockchain, ctx.GetConfig().DataDir)

	if man.protocolManager, err = NewProtocolManager(man.chainConfig, config.SyncMode, config.NetworkId, man.eventMux, man.txPool, man.engine, man.blockchain, chainDb, ctx.MsgCenter); err != nil {
		return nil, err
	}
	//man.protocolManager.Msgcenter = ctx.MsgCenter
	MsgCenter = ctx.MsgCenter
	man.miner, err = miner.New(man.blockchain, man.chainConfig, man.EventMux(), man.engine, man.blockchain.DPOSEngine(), man.hd, man.CA())
	if err != nil {
		return nil, err
	}
	man.miner.SetExtra(makeExtraData(config.ExtraData))

	//algorithm
	dbDir := ctx.GetConfig().DataDir
	man.reelection, err = reelection.New(man.blockchain, dbDir)
	if err != nil {
		return nil, err
	}
	man.random, err = random.New(man.msgcenter)
	if err != nil {
		return nil, err
	}

	man.APIBackend = &ManAPIBackend{man, nil}
	gpoParams := config.GPO
	if gpoParams.Default == nil {
		gpoParams.Default = config.GasPrice
	}
	man.APIBackend.gpo = gasprice.NewOracle(man.APIBackend, gpoParams)
	depoistInfo.NewDepositInfo(man.APIBackend)
	man.broadTx = broadcastTx.NewBroadCast(man.APIBackend) //YY

	man.leaderServer, err = leaderelect.NewLeaderIdentityService(man, "leader服务")

	man.topNode = olconsensus.NewTopNodeService(man.blockchain.DPOSEngine())
	topNodeInstance := olconsensus.NewTopNodeInstance(man.signHelper, man.hd)
	man.topNode.SetTopNodeStateInterface(topNodeInstance)
	man.topNode.SetValidatorAccountInterface(topNodeInstance)
	man.topNode.SetMessageSendInterface(topNodeInstance)
	man.topNode.SetMessageCenterInterface(topNodeInstance)

	if err = man.topNode.Start(); err != nil {
		return nil, err
	}

	man.blockgen, err = blkgenor.New(man)
	if err != nil {
		return nil, err
	}

	man.blockVerify, err = blkverify.NewBlockVerify(man)
	if err != nil {
		return nil, err
	}

	return man, nil
}

func makeExtraData(extra []byte) []byte {
	if len(extra) == 0 {
		// create default extradata
		extra, _ = rlp.EncodeToBytes([]interface{}{
			uint(params.VersionMajor<<16 | params.VersionMinor<<8 | params.VersionPatch),
			"gman",
			runtime.Version(),
			runtime.GOOS,
		})
	}
	if uint64(len(extra)) > params.MaximumExtraDataSize {
		log.Warn("Miner extra data exceed limit", "extra", hexutil.Bytes(extra), "limit", params.MaximumExtraDataSize)
		extra = nil
	}
	return extra
}

// CreateDB creates the chain database.
func CreateDB(ctx *pod.ServiceContext, config *Config, name string) (mandb.Database, error) {
	db, err := ctx.OpenDatabase(name, config.DatabaseCache, config.DatabaseHandles)
	if err != nil {
		return nil, err
	}
	if db, ok := db.(*mandb.LDBDatabase); ok {
		db.Meter("man/db/chaindata/")
	}
	return db, nil
}

// CreateConsensusEngine creates the required type of consensus engine instance for an Matrix service
func CreateConsensusEngine(ctx *pod.ServiceContext, config *manash.Config, chainConfig *params.ChainConfig, db mandb.Database) consensus.Engine {
	// If proof-of-authority is requested, set it up
	if chainConfig.Clique != nil {
		return clique.New(chainConfig.Clique, db)
	}
	// Otherwise assume proof-of-work
	switch config.PowMode {
	case manash.ModeFake:
		log.Warn("Manash used in fake mode")
		return manash.NewFaker()
	case manash.ModeTest:
		log.Warn("Manash used in test mode")
		return manash.NewTester()
	case manash.ModeShared:
		log.Warn("Manash used in shared mode")
		return manash.NewShared()
	default:
		engine := manash.New(manash.Config{
			CacheDir:       ctx.ResolvePath(config.CacheDir),
			CachesInMem:    config.CachesInMem,
			CachesOnDisk:   config.CachesOnDisk,
			DatasetDir:     config.DatasetDir,
			DatasetsInMem:  config.DatasetsInMem,
			DatasetsOnDisk: config.DatasetsOnDisk,
		})
		engine.SetThreads(-1) // Disable CPU mining
		return engine
	}
}

// APIs return the collection of RPC services the matrix package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (s *Matrix) APIs() []rpc.API {
	apis := manapi.GetAPIs(s.APIBackend)

	// Append any APIs exposed explicitly by the consensus engine
	apis = append(apis, s.engine.APIs(s.BlockChain())...)

	// Append all the local APIs and return
	return append(apis, []rpc.API{
		{
			Namespace: "eth",
			Version:   "1.0",
			Service:   NewPublicMatrixAPI(s),
			Public:    true,
		}, {
			Namespace: "eth",
			Version:   "1.0",
			Service:   NewPublicMinerAPI(s),
			Public:    true,
		}, {
			Namespace: "eth",
			Version:   "1.0",
			Service:   downloader.NewPublicDownloaderAPI(s.protocolManager.downloader, s.eventMux),
			Public:    true,
		}, {
			Namespace: "miner",
			Version:   "1.0",
			Service:   NewPrivateMinerAPI(s),
			Public:    false,
		}, {
			Namespace: "eth",
			Version:   "1.0",
			Service:   filters.NewPublicFilterAPI(s.APIBackend, false),
			Public:    true,
		}, {
			Namespace: "admin",
			Version:   "1.0",
			Service:   NewPrivateAdminAPI(s),
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPublicDebugAPI(s),
			Public:    true,
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPrivateDebugAPI(s.chainConfig, s),
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   s.netRPCService,
			Public:    true,
		},
	}...)
}

func (s *Matrix) ResetWithGenesisBlock(gb *types.Block) {
	s.blockchain.ResetWithGenesisBlock(gb)
}

func (s *Matrix) Manerbase() (eb common.Address, err error) {
	s.lock.RLock()
	manbase := s.manbase
	s.lock.RUnlock()

	if manbase != (common.Address{}) {
		return manbase, nil
	}
	if wallets := s.AccountManager().Wallets(); len(wallets) > 0 {
		if accounts := wallets[0].Accounts(); len(accounts) > 0 {
			manbase := accounts[0].Address

			s.lock.Lock()
			s.manbase = manbase
			s.lock.Unlock()

			log.Info("Manerbase automatically configured", "address", manbase)
			return manbase, nil
		}
	}
	return common.Address{}, fmt.Errorf("manbase must be explicitly specified")
}

// SetManerbase sets the mining reward address.
func (s *Matrix) SetManerbase(manbase common.Address) {
	s.lock.Lock()
	s.manbase = manbase
	s.lock.Unlock()

	s.miner.SetManerbase(manbase)
}

func (s *Matrix) StartMining(local bool) error {
	eb, err := s.Manerbase()
	if err != nil {
		log.Error("Cannot start mining without manbase", "err", err)
		return fmt.Errorf("manbase missing: %v", err)
	}
	if clique, ok := s.engine.(*clique.Clique); ok {
		wallet, err := s.accountManager.Find(accounts.Account{Address: eb})
		if wallet == nil || err != nil {
			log.Error("Manerbase account unavailable locally", "err", err)
			return fmt.Errorf("signer missing: %v", err)
		}
		clique.Authorize(eb, wallet.SignHash)
	}
	if local {
		// If local (CPU) mining is started, we can disable the transaction rejection
		// mechanism introduced to speed sync times. CPU mining on mainnet is ludicrous
		// so none will ever hit this path, whereas marking sync done on CPU mining
		// will ensure that private networks work in single miner mode too.
		atomic.StoreUint32(&s.protocolManager.acceptTxs, 1)
	}
	go s.miner.Start(eb)
	return nil
}

func (s *Matrix) StopMining()         { s.miner.Stop() }
func (s *Matrix) IsMining() bool      { return s.miner.Mining() }
func (s *Matrix) Miner() *miner.Miner { return s.miner }

func (s *Matrix) AccountManager() *accounts.Manager    { return s.accountManager }
func (s *Matrix) BlockChain() *core.BlockChain         { return s.blockchain }
func (s *Matrix) TxPool() *core.TxPool                 { return s.txPool }
func (s *Matrix) EventMux() *event.TypeMux             { return s.eventMux }
func (s *Matrix) Engine() consensus.Engine             { return s.engine }
func (s *Matrix) DPOSEngine() consensus.DPOSEngine     { return s.blockchain.DPOSEngine() }
func (s *Matrix) ChainDb() mandb.Database              { return s.chainDb }
func (s *Matrix) IsListening() bool                    { return true } // Always listening
func (s *Matrix) ManVersion() int                      { return int(s.protocolManager.SubProtocols[0].Version) }
func (s *Matrix) NetVersion() uint64                   { return s.networkId }
func (s *Matrix) Downloader() *downloader.Downloader   { return s.protocolManager.downloader }
func (s *Matrix) CA() *ca.Identity                     { return s.ca }
func (s *Matrix) MsgCenter() *mc.Center                { return s.msgcenter }
func (s *Matrix) SignHelper() *signhelper.SignHelper   { return s.signHelper }
func (s *Matrix) ReElection() *reelection.ReElection   { return s.reelection }
func (s *Matrix) HD() *msgsend.HD                      { return s.hd }
func (s *Matrix) TopNode() *olconsensus.TopNodeService { return s.topNode }

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (s *Matrix) Protocols() []p2p.Protocol {
	if s.lesServer == nil {
		return s.protocolManager.SubProtocols
	}
	return append(s.protocolManager.SubProtocols, s.lesServer.Protocols()...)
}

// Start implements node.Service, starting all internal goroutines needed by the
// Matrix protocol implementation.
func (s *Matrix) Start(srvr *p2p.Server) error {
	// Start the bloom bits servicing goroutines
	s.startBloomHandlers()

	// Start the RPC service
	s.netRPCService = manapi.NewPublicNetAPI(srvr, s.NetVersion())

	// Figure out a max peers count based on the server limits
	maxPeers := srvr.MaxPeers
	if s.config.LightServ > 0 {
		if s.config.LightPeers >= srvr.MaxPeers {
			return fmt.Errorf("invalid peer config: light peer count (%d) >= total peer count (%d)", s.config.LightPeers, srvr.MaxPeers)
		}
		maxPeers -= s.config.LightPeers
	}
	// Start the networking layer and the light server if requested
	s.protocolManager.Start(maxPeers)
	if s.lesServer != nil {
		s.lesServer.Start(srvr)
	}
	//s.broadTx.Start()//YY
	return nil
}
func (s *Matrix) FetcherNotify(hash common.Hash, number uint64) {
	ids := ca.GetRolesByGroup(common.RoleValidator | common.RoleBroadcast)
	for _, id := range ids {
		peer := s.protocolManager.Peers.Peer(id.String()[:16])
		if peer == nil {
			log.Info("==========YY===========", "get PeerID is nil by Validator ID:id", id.String(), "Peers:", s.protocolManager.Peers.peers)
			continue
		}
		s.protocolManager.fetcher.Notify(id.String()[:16], hash, number, time.Now(), peer.RequestOneHeader, peer.RequestBodies)
	}
}

// Stop implements node.Service, terminating all internal goroutines used by the
// Matrix protocol.
func (s *Matrix) Stop() error {
	s.bloomIndexer.Close()
	s.blockchain.Stop()
	s.protocolManager.Stop()
	if s.lesServer != nil {
		s.lesServer.Stop()
	}
	s.txPool.Stop()
	s.miner.Stop()
	s.eventMux.Stop()

	s.chainDb.Close()
	s.broadTx.Stop() //YY
	close(s.shutdownChan)

	return nil
}
