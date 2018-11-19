package lottery

import (
	"math/big"
	"sort"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/params"
)

const (
	N      = 10
	FIRST  = 1 //一等奖数目
	SECOND = 2 //二等奖数目
	THIRD  = 3 //三等奖数目
)

var (
	FIRSTPRIZE   *big.Int = big.NewInt(5e+18) //一等奖金额  5man
	SENCONDPRIZE *big.Int = big.NewInt(3e+18) //二等奖金额 2man
	THIRDPRIZE   *big.Int = big.NewInt(1e+18) //三等奖金额 1man
)

type TxCmpResult struct {
	Tx        *types.SelfTransaction
	CmpResult int64
}

// A slice of Pairs that implements sort.Interface to sort by Value.
type TxCmpResultList []TxCmpResult

func (p TxCmpResultList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p TxCmpResultList) Len() int           { return len(p) }
func (p TxCmpResultList) Less(i, j int) bool { return p[i].CmpResult < p[j].CmpResult }

type ChainReader interface {
	// Config retrieves the blockchain's chain configuration.
	Config() *params.ChainConfig

	// CurrentHeader retrieves the current header from the local chain.
	CurrentHeader() *types.Header

	// GetHeader retrieves a block header from the database by hash and number.
	GetHeader(hash common.Hash, number uint64) *types.Header

	// GetHeaderByNumber retrieves a block header from the database by number.
	GetHeaderByNumber(number uint64) *types.Header

	// GetHeaderByHash retrieves a block header from the database by its hash.
	GetHeaderByHash(hash common.Hash) *types.Header

	// GetBlock retrieves a block sfrom the database by hash and number.
	GetBlock(hash common.Hash, number uint64) *types.Block

	GetBlockByNumber(number uint64) *types.Block
}

type TxsLottery struct {
	chain ChainReader
	seed  LotterySeed
}

type LotterySeed interface {
	GetSeed(num uint64) *big.Int
}

func New(chain ChainReader, seed LotterySeed) *TxsLottery {
	tlr := &TxsLottery{
		chain: chain,
		seed:  seed,
	}

	return tlr
}
func abs(n int64) int64 {
	y := n >> 63
	return (n ^ y) - y
}

func (tlr *TxsLottery) LotteryCalc(num uint64) map[string]map[common.Address]*big.Int {
	//选举周期的最后时刻分配
	if !common.IsReElectionNumber(num + 1) {
		return nil
	}
	LotteryAccount := make(map[string]map[common.Address]*big.Int, 0)
	txsCmpResultList := tlr.getLotteryList(num, 10)
	tlr.lotteryChoose(txsCmpResultList, LotteryAccount)

	return LotteryAccount
}

func (tlr *TxsLottery) getLotteryList(num uint64, lotteryNum int) TxCmpResultList {
	originBlockNum := common.GetLastBroadcastNumber(num) + 1

	randSeed := tlr.seed.GetSeed(num)
	expHash := common.BigToHash(randSeed)
	txsCmpResultList := make(TxCmpResultList, 0)
	for originBlockNum < num {

		if !common.IsBroadcastNumber(originBlockNum) {
			continue
		}
		txs := tlr.chain.GetBlockByNumber(originBlockNum).Transactions()
		for _, tx := range txs {
			extx := tx.GetMatrix_EX()
			if (extx != nil) && len(extx) > 0 && extx[0].TxType == common.ExtraNormalTxType||extx == nil {
				txCmpResult := TxCmpResult{&tx, abs(tx.Hash().Big().Int64() - expHash.Big().Int64())}
				txsCmpResultList = append(txsCmpResultList, txCmpResult)
			}

		}

		originBlockNum++

	}
	sort.Sort(txsCmpResultList)
	return txsCmpResultList[0:lotteryNum]
}

func (tlr *TxsLottery) lotteryChoose(txsCmpResultList TxCmpResultList, LotteryAccountMap map[string]map[common.Address]*big.Int) {
	firstLottery := make(map[common.Address]*big.Int, FIRST)
	secondLottery := make(map[common.Address]*big.Int, SECOND)
	thirdLottery := make(map[common.Address]*big.Int, THIRD)
	for _, v := range txsCmpResultList {
		from, err := types.Sender(types.NewEIP155Signer(tlr.chain.Config().ChainId), *v.Tx)
		if nil != err {
			continue
		}
		//抽取一等奖
		LotteryAccount, _ := LotteryAccountMap["First"]
		if len(LotteryAccount) < FIRST {

			firstLottery[from] = FIRSTPRIZE

			continue
		}
		//抽取过的账户跳过
		//if nil != tlr.chooseIn(LotteryAccount, from) {
		//	continue
		//}
		//抽取二等奖
		LotteryAccount, _ = LotteryAccountMap["Second"]
		if len(LotteryAccount) < SECOND {

			secondLottery[from] = SENCONDPRIZE

			continue
		}

		//抽取过的账户跳过
		//if nil != tlr.chooseIn(LotteryAccount, from) {
		//	continue
		//}
		//抽取三等奖
		LotteryAccount, _ = LotteryAccountMap["Third"]
		if len(LotteryAccount) < THIRD {
			thirdLottery[from] = THIRDPRIZE
			continue
		}
		break

	}
	LotteryAccountMap["First"] = firstLottery
	LotteryAccountMap["Second"] = secondLottery
	LotteryAccountMap["third"] = thirdLottery
}
