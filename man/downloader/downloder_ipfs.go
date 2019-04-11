// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package downloader

import (
	"bytes"
	"compress/gzip"
	"container/list"
	"context"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/rlp"
	"gopkg.in/karalabe/cookiejar.v2/collections/prque"
)

const (
	IpfsHashLen                 = 46
	LastestBlockStroeNum        = 100    //100
	Cache2StoreHashMaxNum       = 216000 //每月产生的区块//测试 2000  //要改成月216000 //2628000 //24 hour* (3600 second /12 second）*365
	Cache1StoreCache2Num        = 6000   //500年的 //测试10000  //改成12000 1000年
	Cache2StoreBatchBlockMaxNum = 8800   //每年产生的//测试20     //要改为8800(24*365 )//约年 产生 的 300单位的区块
	Cache1StoreBatchCache2Num   = 500    //500年的//测试100    // 要改为1000  年
	BATCH_NUM                   = 300
	SingleBlockStore            = false
)

//var logPrint bool = false
//var logOriginMode bool = false
//var gBlockFile *File

type Hash []byte //[IpfsHashLen]byte
type BlockStore struct {
	Numberstore map[uint64]NumberMapingCoupledHash
}
type NumberMapingCoupledHash struct {
	//Blockhash map[common.Hash]string
	Blockhash map[string]string
}

// lastest 100
type LastestBlcokCfg struct {
	CurrentNum uint64
	HashNum    int
	MapList    BlockStore
}

//new cache2
type Caches2CfgMap struct {
	CurCacheBeignNum uint64
	CurCacheBlockNum uint64
	NumHashStore     uint32
	//CurOfCache1Pos	 uint32
	MapList BlockStore
}

//cache1
type Cache1StoreCfg struct {
	OriginBlockNum    uint64
	CurrentBlockNum   uint64
	Cache2FileNum     uint32
	StCahce2Hash      [Cache1StoreCache2Num]string
	StBatchCahce2Hash [Cache1StoreBatchCache2Num]string
}
type DownloadRetry struct {
	header       *types.Header
	realBeginNum uint64
	flag         uint64
	ReqPendflg   int
	coinstr      string
	downNum      int
}
type IPfsDownloader struct {
	BIpfsIsRunning      bool
	IpfsDealBlocking    int32
	StrIpfspeerID       string
	StrIpfsSecondpeerID string
	StrIPFSLocationPath string
	StrIPFSServerInfo   string
	StrIPFSExecName     string
	HeaderIpfsCh        chan []BlockIpfsReq //[]*types.Header
	//	BlockRcvCh          chan *types.Block
	//runQuit      chan struct{}
	//timeOutCh    chan struct{}
	DownMutex       *sync.Mutex
	DownRetrans     *list.List //*prque.Prque // []DownloadRetry prque.New()
	BatchStBlock    *BatchBlockSt
	SnapshootInfoCh chan SnapshootReq
}

type listBlockInfo struct {
	blockNum      uint64
	blockHeadHash string   //common.Hash
	coinKind      []string //多币种
	blockIpfshash []string //Hash
}

//var HeaderIpfsCh chan []*types.Header

var (
	strLastestBlockFile  = "lastestblockInfo.gb" //存放最近100块缓存文件
	strCache1BlockFile   = "firstCacheInfo.jn"   //存放一级缓存文件
	strCacheDirectory    = "ipfsCachecommon"     //发布的目录
	strTmpCache2File     = "secondCacheInfo.gb"  //暂时存放查询的二级缓存文件
	strNewBlockStoreFile = "NewTmpBlcok.rp"      //新来的block 暂时保存文件
	//	StrIpfspeerID        = "QmPXtaMvY6ZB67Xgeb8M2D8KuyPBXbyVEyTzaxs5TpjuNi" //peer id
	strBatchHeaderFile    = "batchheader.rp"
	strBatchBodyFile      = "batchbody.rp"
	strBatchReceiptFile   = "batchreceipt.rp"
	strTmpBatchCache2File = "secondBatchCacheInfo.gb"
)
var HeadBatchFlag uint64 = 0x12345678
var BodyBatchFlag uint64 = 0x23456781
var ReceiptBatchFlag uint64 = 0x34567812

type GetIpfsCache struct {
	bassign            bool
	lastestCache       *LastestBlcokCfg
	lastestNum         uint64
	getipfsCache1      *Cache1StoreCfg
	getipfsCache2      *Caches2CfgMap
	getipfsBatchCache2 *Caches2CfgMap
}
type StoreIpfsCache struct {
	curBlockPos      uint32
	curBatchBlockPos uint32
	storeipfsCache1  *Cache1StoreCfg
}
type DownloadFileInfo struct {
	Downloadflg          bool
	IpfsPath             string
	StrIPFSServerInfo    string
	StrIPFSServer2Info   string
	StrIPFSServer3Info   string
	StrIPFSServer4Info   string
	StrIPFSServer5Info   string
	StrIPFSServer6Info   string
	StrIPFSServer7Info   string
	StrIPFSServer8Info   string
	StrIPFSServer9Info   string
	StrIPFSServer10Info  string
	PrimaryDescription   string
	SecondaryDescription string
}

type BatchBlockSt struct {
	curBlockNum        uint64
	ExpectBeginNum     uint64
	ExpectBeginNumhash string //common.Hash
	minBlockNum        uint64
	minBlockNumHash    string //common.Hash
	fileflag           int
	headerStoreFile    *os.File
	bodyStoreFile      *os.File
	receiptStoreFile   *os.File
}
type StopIpfs struct {
	Stop func()
}
type IPFSBlockStat struct {
	gIPFSerrorNum          int
	curBlockNum            uint64
	totalBlockSize         int64
	totalZipBlockSize      int64
	totalBatchBlockSize    int64
	totalZipBatchBlockSize int64
	totalSnapDataSize      int64
	totalZipSnapDataSize   int64
}

var StopIpfsHandler = StopIpfs{}

var gIpfsCache GetIpfsCache
var gIpfsStoreCache StoreIpfsCache
var IpfsInfo DownloadFileInfo
var logMap bool
var listPeerId [2]string
var testShowlog int
var gIpfsPath string
var runQuit chan int         //struct{}
var timeOutFlg int           //chan int
var gtimeOutSign chan string //struct{}
var gIpfsStat IPFSBlockStat
var gIpfsProcessBlockNumber uint64

func init() {
	/*creatInfo := DownloadFileInfo{
		Downloadflg:          true,
		IpfsPath:             "D:\\lb\\go-ipfs",
		StrIPFSServerInfo:    "/ip4/192.168.3.30/tcp/4001/ipfs/QmQSazdGapokSejxeTTQc4tCRcHgqRPtoMeW3trRk4zA1S",
		PrimaryDescription:   "QmPXtaMvY6ZB67Xgeb8M2D8KuyPBXbyVEyTzaxs5TpjuNi",
		SecondaryDescription: "",
	}
	WriteJsFile("ipfsinfo.json", creatInfo)*/
	//len := GetFileSize("D:\\send1.txt")
	//err :=
	ReadJsFile("ipfsinfo.json", &IpfsInfo)
	//fmt.Println("read ipfs ", err, IpfsInfo.Downloadflg, IpfsInfo.StrIPFSServerInfo)
	if /*IpfsInfo.IpfsPath == "" ||*/ IpfsInfo.StrIPFSServerInfo == "" || IpfsInfo.PrimaryDescription == "" {
		IpfsInfo.Downloadflg = false
	} else {
		runQuit = make(chan int)         //struct{})
		gtimeOutSign = make(chan string) //struct{})
		//timeOutCh = make(chan int)
	}
}
func GetIpfsMode() bool {
	return IpfsInfo.Downloadflg
}
func newIpfsDownload() *IPfsDownloader {

	return &IPfsDownloader{
		BIpfsIsRunning: false,
		HeaderIpfsCh:   make(chan []BlockIpfsReq, 1), //*types.Header, 1),
		//	BlockRcvCh:     make(chan *types.Block, 1),
		//runQuit:      make(chan struct{}),
		DownRetrans:     list.New(), //prque.New(), //make([]DownloadRetry, 6),
		DownMutex:       new(sync.Mutex),
		BatchStBlock:    new(BatchBlockSt),
		SnapshootInfoCh: make(chan SnapshootReq),
		//timeOutCh:    make(chan struct{}),
	}

}

//var gStrIpfsName string

func CheckDirAndCreate(dir string) error {
	_, err := os.Stat(dir)
	if err == nil {
		return nil
	}
	if os.IsNotExist(err) {
		return os.Mkdir(dir, os.ModePerm)
	}
	return fmt.Errorf("know")

}
func GetFileSize(filename string) int64 {
	var result int64
	filepath.Walk(filename, func(path string, f os.FileInfo, err error) error {
		result = f.Size()
		return nil
	})
	return result
}
func (d *Downloader) IpfsDownloadTestInit() error {

	CheckDirAndCreate(strCacheDirectory)
	//
	d.dpIpfs.StrIPFSLocationPath = "D:\\lb\\go-ipfs"
	d.dpIpfs.StrIPFSExecName = d.dpIpfs.StrIPFSLocationPath + "\\ipfs.exe"
	//gStrIpfsPath = d.dpIpfs.StrIPFSLocationPath
	gIpfsPath = d.dpIpfs.StrIPFSExecName
	d.dpIpfs.StrIPFSServerInfo = "/ip4/192.168.3.30/tcp/4001/ipfs/QmQSazdGapokSejxeTTQc4tCRcHgqRPtoMeW3trRk4zA1S"
	return nil // foe test
}

func StopIpfsProcess() {
	log.Warn("ipfs Downloader StopIpfsProcess  manual exit")
	StopIpfsHandler.Stop()
	fmt.Println("ipfs manual exit")
}

var strAnother = "The process cannot access the file because it is being used by another process"
var strIPFSstdErr = "api not running"
var strIPFSstd2Err = "routing: not found"
var strIPFSpatherr = "file does not exist"

func (d *Downloader) IpfsDownloadInit() error {

	var out bytes.Buffer
	var outerr bytes.Buffer
	// Directory
	CheckDirAndCreate(strCacheDirectory)
	//fmt.Println("IpfsDownloadInit enter")
	//
	d.dpIpfs.StrIPFSLocationPath = IpfsInfo.IpfsPath      //"D:\\lb\\go-ipfs"
	d.dpIpfs.StrIPFSExecName = IpfsInfo.IpfsPath + "ipfs" //d.dpIpfs.StrIPFSLocationPath + "\\ipfs.exe"
	gIpfsPath = d.dpIpfs.StrIPFSExecName                  //gStrIpfsPath = d.dpIpfs.StrIPFSLocationPath
	d.dpIpfs.StrIpfspeerID = IpfsInfo.PrimaryDescription
	d.dpIpfs.StrIpfsSecondpeerID = IpfsInfo.SecondaryDescription
	d.dpIpfs.StrIPFSServerInfo = IpfsInfo.StrIPFSServerInfo //"/ip4/192.168.3.30/tcp/4001/ipfs/QmQSazdGapokSejxeTTQc4tCRcHgqRPtoMeW3trRk4zA1S"
	//	return nil // foe test
	listPeerId[0] = d.dpIpfs.StrIpfspeerID
	listPeerId[1] = d.dpIpfs.StrIpfsSecondpeerID
	if d.dpIpfs.StrIpfsSecondpeerID != "" {
		listPeerId[1] = d.dpIpfs.StrIpfsSecondpeerID
	}
	fmt.Println("peer ID ", listPeerId[0], listPeerId[1])
	log.Warn("ipfs Downloader init", "peerid0", listPeerId[0], "peerid1", listPeerId[1])
	//d.dpIpfs.BatchStBlock = new(BatchBlockSt)
	out.Reset()
	outerr.Reset()
	c := exec.Command(d.dpIpfs.StrIPFSExecName, "init")
	c.Stdout = &out
	c.Stderr = &outerr
	err := c.Run()

	strErrInfo := outerr.String()
	strttt := err.Error()
	if err != nil {
		log.Warn("ipfs IpfsDownloadInit init error", "error", err, "ipfs err", strErrInfo) //, "fff", strttt)
		//return err
		if strings.Index(strttt, strIPFSpatherr) > 0 {
			d.IpfsMode = false //启动失败时 置为false
			IpfsInfo.Downloadflg = false
			d.bIpfsDownload = 0
			log.Warn("ipfs IpfsDownloadInit init error", strttt)
			return nil
		}
	}
	out.Reset()
	outerr.Reset()
	c = exec.Command(d.dpIpfs.StrIPFSExecName, "bootstrap", "rm", "all")
	err = c.Run()
	strErrInfo = outerr.String()

	if err != nil {
		log.Error("ipfs IpfsDownloadInit bootstrap rm error", "error", err, "ipfs err", strErrInfo)
		//return err
	}
	out.Reset()
	outerr.Reset()
	c = exec.Command(d.dpIpfs.StrIPFSExecName, "bootstrap", "add", d.dpIpfs.StrIPFSServerInfo) //"/ip4/192.168.3.30/tcp/4001/ipfs/QmQSazdGapokSejxeTTQc4tCRcHgqRPtoMeW3trRk4zA1S")
	err = c.Run()
	strErrInfo = outerr.String()

	if err != nil {
		log.Error("ipfs IpfsDownloadInit bootstrap add error", "error", err, "ipfs err", strErrInfo)
		//return err
	}
	if IpfsInfo.StrIPFSServer2Info != "" {
		c = exec.Command(d.dpIpfs.StrIPFSExecName, "bootstrap", "add", IpfsInfo.StrIPFSServer2Info)
		err = c.Run()
	}
	if IpfsInfo.StrIPFSServer3Info != "" {
		c = exec.Command(d.dpIpfs.StrIPFSExecName, "bootstrap", "add", IpfsInfo.StrIPFSServer3Info)
		err = c.Run()
	}
	if IpfsInfo.StrIPFSServer4Info != "" {
		c = exec.Command(d.dpIpfs.StrIPFSExecName, "bootstrap", "add", IpfsInfo.StrIPFSServer4Info)
		err = c.Run()
	}
	if IpfsInfo.StrIPFSServer5Info != "" {
		c = exec.Command(d.dpIpfs.StrIPFSExecName, "bootstrap", "add", IpfsInfo.StrIPFSServer5Info)
		err = c.Run()
	}
	if IpfsInfo.StrIPFSServer6Info != "" {
		c = exec.Command(d.dpIpfs.StrIPFSExecName, "bootstrap", "add", IpfsInfo.StrIPFSServer6Info)
		err = c.Run()
	}
	if IpfsInfo.StrIPFSServer7Info != "" {
		c = exec.Command(d.dpIpfs.StrIPFSExecName, "bootstrap", "add", IpfsInfo.StrIPFSServer7Info)
		err = c.Run()
	}
	if IpfsInfo.StrIPFSServer8Info != "" {
		c = exec.Command(d.dpIpfs.StrIPFSExecName, "bootstrap", "add", IpfsInfo.StrIPFSServer8Info)
		err = c.Run()
	}
	if IpfsInfo.StrIPFSServer9Info != "" {
		c = exec.Command(d.dpIpfs.StrIPFSExecName, "bootstrap", "add", IpfsInfo.StrIPFSServer9Info)
		err = c.Run()
	}
	if IpfsInfo.StrIPFSServer10Info != "" {
		c = exec.Command(d.dpIpfs.StrIPFSExecName, "bootstrap", "add", IpfsInfo.StrIPFSServer10Info)
		err = c.Run()
	}

	out.Reset()
	outerr.Reset()

	fmt.Println("ipfs daemon run", d.dpIpfs.StrIPFSExecName)
	/*
		c = exec.Command(d.dpIpfs.StrIPFSExecName, "daemon") //"add", "D:\\melog3332.txt")
		d.dpIpfs.BIpfsIsRunning = true
		err = c.Run()
	*/

	ctx, cancel := context.WithCancel(context.Background())
	StopIpfsHandler.Stop = cancel
	//err = exec.CommandContext(ctx, d.dpIpfs.StrIPFSExecName, "daemon").Run()
	cm := exec.CommandContext(ctx, d.dpIpfs.StrIPFSExecName, "daemon")
	cm.Stdout = &out
	cm.Stderr = &outerr
	err = cm.Run()
	strErrInfo = outerr.String()

	d.dpIpfs.BIpfsIsRunning = false
	if err != nil {
		log.Error("ipfs IpfsDownloadInit daemon error,exit init", "error", err, "out", out.String(), "ipfs err", outerr.String())
		//return err
	} //"The process cannot access the file because it is being used by another process"
	if strings.Index(strErrInfo, strAnother) >= 0 {

	} else {
		/*d.IpfsMode = false //启动失败时 置为false
		IpfsInfo.Downloadflg = false
		d.bIpfsDownload = 0*/
	}
	fmt.Println("ipfsDownloadInit error", err)

	return nil
}
func RestartIpfsDaemon() {
	var outerr bytes.Buffer
	var out bytes.Buffer

	StopIpfsProcess()
	//time.Sleep(1 * time.Second)
	log.Warn("ipfs RestartIpfsDaemon daemon")
	ctx, cancel := context.WithCancel(context.Background())
	StopIpfsHandler.Stop = cancel
	flgCh := make(chan int)
	go func() {
		fmt.Println("ipfs restart again", gIpfsPath)
		cd := exec.CommandContext(ctx, gIpfsPath, "daemon")
		cd.Stdout = &out
		cd.Stderr = &outerr
		log.Warn("ipfs RestartIpfsDaemon daemon again")
		flgCh <- 0
		err := cd.Run()
		//strErrInfo := outerr.String()
		//d.dpIpfs.BIpfsIsRunning = true
		if err != nil {
			log.Error("ipfs RestartIpfsDaemon daemon error, will exit", "error", err, "out", out.String(), "outerr", outerr.String())
			//return err
		}
		//d.IpfsMode = false //启动失败时 置为false
		fmt.Println("****RestartIpfs Daemon will exit", err, out.String(), outerr.String())
	}()

	<-flgCh
	log.Warn("ipfs---- RestartIpfsDaemon daemon over")
	fmt.Println("ipfs ---restart over")
	time.Sleep(20 * time.Second)
	log.Warn("ipfs---- RestartIpfsDaemon daemon sleep over")
}
func CheckIpfsStatus(err error) {
	//log.Warn("ipfs---- CheckIpfsStatus--------")
	if err.Error() == "exit status 1" {
		RestartIpfsDaemon()
	}

}

func ipfsGetTimeout() {
	//return
	var flg int
	recvSync := time.NewTicker(8 * time.Minute)
	defer func() {
		log.Warn("ipfs---- ipfsGetTimeout out--------", "flg", flg)
		recvSync.Stop()
	}()
	//log.Warn("ipfs---- ipfsGetTimeout in")
	//quit    chan struct{}
	for {
		select {
		case <-recvSync.C:
			log.Warn("ipfs---- ipfsGetTimeout")
			flg = 1
			RestartIpfsDaemon()
			timeOutFlg = 1
		case <-runQuit:
			flg = 0
			return
		}
	}
}
func TimeoutExec(str string) {
	var flg int
	timeOut := time.NewTicker(16 * time.Minute)
	defer func() {
		log.Warn("ipfs---- ipfsTimeoutExec out--------", "flg", flg, "hash", str)
		timeOut.Stop()
	}()
	for {
		select {
		case <-timeOut.C:
			log.Warn("ipfs---- TimeoutExec time out")
			flg = 1
			RestartIpfsDaemon()
			timeOutFlg = 1
		case <-runQuit:
			flg = 0
			return
		}
	}
}
func IpfsStartTimer(hash string) {
	timeOutFlg = 0
	gtimeOutSign <- hash //- struct{}{}
}
func IpfsStopTimer() {
	//超时的不需要再停止
	if timeOutFlg == 0 {
		runQuit <- 1
	}
}
func (d *Downloader) IpfsTimeoutTask() {
	log.Warn("IpfsTimeoutTask enter in")
	for {
		select {
		case str := <-gtimeOutSign:
			timeOutFlg = 0
			TimeoutExec(str)
		}
	}

}
func (d *Downloader) dealIPFSerrorProc() {
	if gIpfsStat.gIPFSerrorNum > 10 {
		log.Warn("ipfs error too much ,disable ipfs download")
		d.bIpfsDownload = 0
	}
}
func (d *Downloader) ClearDirectoryContent() {
	tmpFile, err1 := os.OpenFile(path.Join(strCacheDirectory, strLastestBlockFile), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	fmt.Println("file1", err1)
	tmpFile2, err2 := os.OpenFile(path.Join(strCacheDirectory, strCache1BlockFile), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	fmt.Println("file1", err2)

	tmpFile.Close()
	tmpFile2.Close()
	if err1 == nil || err2 == nil {
		fmt.Println("broadcast clear directory")
		log.Warn("broadcast clear directory")
		d.IPfsDirectoryUpdate()
	}

}

//WriteJsFile serialize
func WriteJsFile(filename string, v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		log.Error("ipfs json.Marshal  error", "error", err)
		return err
	}
	err = ioutil.WriteFile(filename, b, os.ModeAppend)
	if err != nil {
		log.Error("ipfs json.Marshal WriteFile error", "error", err)
		return err
	}
	return nil
}

//ReadJsFile serialize
func ReadJsFile(filename string, v interface{}) error {

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Error("ipfs json.Unmarshal ReadFile error", "error", err)
		return err
	}
	//decode
	err = json.Unmarshal(data, v)
	if err != nil {
		len := GetFileSize(filename)
		log.Error("ipfs json.Marshal error", "error", err, "fileSize", len)
		return err
	}
	return nil
}

//  storeCache serialize gob
func storeCache(data interface{}, file *os.File) error {
	file.Truncate(0)
	buffer := new(bytes.Buffer)
	encoder := gob.NewEncoder(buffer)
	err := encoder.Encode(data)
	if err != nil {
		log.Error("error store gob Encode error", "error", err)
		return fmt.Errorf("ipfs error store gob Encode error", err)
	}
	_, err = file.Write(buffer.Bytes())
	return err
}

//loadCache serialize gob
func loadCache(data interface{}, len int32, file *os.File) error {
	//rawBuf := make([]byte, len)
	//ioutil.ReadFile(name)
	//readLen, err := file.Read(rawBuf)
	rawBuf, err := ioutil.ReadAll(file)
	if err != nil {
		log.Error("ipfs loadCache load file read error", "error", err)
		return fmt.Errorf("ipfs load file read error", err)
	}
	buffer := bytes.NewBuffer(rawBuf)
	dec := gob.NewDecoder(buffer)
	err = dec.Decode(data)
	if err != nil {
		len := GetFileSize(file.Name())
		log.Error("ipfs error store gob decode error", "error", err, "fileSize", len)
		return fmt.Errorf("ipfs error store gob decode error", err)
	}
	return nil
}

//解压 压缩

func CompressFile(Dst string, Src string) error {
	newfile, err := os.Create(Dst)
	if err != nil {
		return err
	}
	defer newfile.Close()

	file, err := os.Open(Src)
	if err != nil {
		return err
	}

	zw := gzip.NewWriter(newfile)

	filestat, err := file.Stat()
	if err != nil {
		return nil
	}

	zw.Name = filestat.Name()
	zw.ModTime = filestat.ModTime()
	_, err = io.Copy(zw, file)
	if err != nil {
		return nil
	}

	zw.Flush()
	if err := zw.Close(); err != nil {
		return nil
	}
	return nil
}

//解压文件Src到Dst
func DeCompressFile(Dst string, Src string) error {
	file, err := os.Open(Src)
	if err != nil {
		return err
	}
	defer file.Close()

	/*buf := make([]byte, 3)
	if n, err := file.Read(buf); err != nil || n < 4 {
		return fmt.Errorf("zip file error")
	}

	if bytes.Equal(buf, []byte("\x1f\x8b\x08\x08")) == false {
		return fmt.Errorf("not zip file")
	}*/

	newfile, err := os.Create(Dst)
	if err != nil {
		return err
	}
	defer newfile.Close()

	zr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}

	filestat, err := file.Stat()
	if err != nil {
		return err
	}

	zr.Name = filestat.Name()
	zr.ModTime = filestat.ModTime()
	_, err = io.Copy(newfile, zr)
	if err != nil {
		return err
	}

	if err := zr.Close(); err != nil {
		return err
	}
	return nil
}

// IpfsGetBlockByHash get block
func IpfsGetBlockByHash(strHash string, compress bool) (*os.File, error) {
	var out bytes.Buffer
	var outerr bytes.Buffer
	var fileName string
	//log.Debug("ipfs IpfsGetBlockByHash info before", "strHash", strHash)
	if strHash == "" {
		log.Error("ipfs IpfsGetBlockByHash strHash error", "strHash", strHash)
		gIpfsStat.gIPFSerrorNum++
		return nil, fmt.Errorf("IpfsGetBlockByHash strHash error")
	}
	c := exec.Command(gIpfsPath, "get", strHash)
	c.Stdout = &out
	c.Stderr = &outerr
	timeOutFlg = 0
	//go ipfsGetTimeout()
	IpfsStartTimer(strHash)
	//	log.Trace("ipfs IpfsGetBlockByHash run in")
	err := c.Run()
	log.Trace("ipfs IpfsGetBlockByHash run out", "strHash", strHash)
	//runQuit <- 1 //struct{}{}
	IpfsStopTimer()
	//c.StdinPipe()
	//strErrInfo := outerr.String()

	log.Debug("ipfs IpfsGetBlockByHash info", "error", err, "strHash", strHash)

	if err != nil {
		log.Error("ipfs IpfsGetBlockByHash error", "error", err, "ipfs err", outerr.String())
		gIpfsStat.gIPFSerrorNum++
		if timeOutFlg == 0 {
			CheckIpfsStatus(err)
		}
		return nil, err
	}

	gIpfsStat.gIPFSerrorNum = 0

	fileName = strHash
	if compress == true {
		fileName = strHash + ".unzip"
		DeCompressFile(fileName, strHash)
		os.Remove(strHash)
	}
	return os.OpenFile(fileName, os.O_RDONLY /*|os.O_APPEND*/, 0644)

}

//IpfsAddNewFile
func IpfsAddNewFile(filePath string, compress bool) (Hash, int64, error) {
	var out bytes.Buffer
	var outerr bytes.Buffer
	var addfilePath string = filePath
	var zipfilesize int64
	if compress == true {
		addfilePath = filePath + ".zip"
		err := CompressFile(addfilePath, filePath)
		if err != nil {
			log.Error("ipfs IpfsAddNewFile CompressFile  error ", "error", err, "filePath", filePath)
			return nil, 0, err
		}
		defer os.Remove(addfilePath)

		fhandler, _ := os.Stat(addfilePath)
		zipfilesize = fhandler.Size()

	}
	c := exec.Command(gIpfsPath, "add", "-q", "-s", "size-1048576", addfilePath) //1M
	c.Stdout = &out
	c.Stderr = &outerr
	err := c.Run()
	//strErrInfo := outerr.String()

	log.Trace("ipfs IpfsAddNewFile to ipfs network", "filePath", addfilePath)

	if err != nil {
		log.Error("ipfs IpfsAddNewFile to  ipfs network", "error", err, "ipfs err", outerr.String())

		RestartIpfsDaemon()
		c = exec.Command(gIpfsPath, "add", "-q", "-s", "size-1048576", addfilePath)
		err = c.Run()
		log.Error("ipfs IpfsAddNewFile to  ipfs network error again")
		if err != nil {
			log.Error("ipfs IpfsAddNewFile to  ipfs network error again", "error", err, "ipfs err", outerr.String())
			return nil, zipfilesize, err
		}
		//return nil, err
	}
	return out.Bytes(), zipfilesize, nil
}

//IpfsGetFileCache2ByHash

func IpfsGetFileCache2ByHash(strhash, objfileName string) (*os.File, bool, error) {
	var out bytes.Buffer
	var outerr bytes.Buffer
	if strhash == "" {
		//var errf error = nil
		tmpBlockFile, errf := os.OpenFile(objfileName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644) //"secondCacheInfo.gb"创建新文件
		if errf != nil {
			return nil, false, fmt.Errorf("ipfs error IpfsGetFileCache2ByHash OpenFile error")
		} else {
			return tmpBlockFile, true, nil
		}
	}

	//tmp := []byte(strhash)
	//strhash2 := string(tmp[0:IpfsHashLen])
	strFile := "-o=" + objfileName
	c := exec.Command(gIpfsPath, "get", strFile, strhash)
	//c := exec.Command(gStrIpfsName, "get", "-o=secondCacheInfo.gb", strhash) //strhash)
	//go ipfsGetTimeout()
	IpfsStartTimer(strhash)
	c.Stdout = &out
	c.Stderr = &outerr
	err := c.Run()
	//runQuit <- 1
	IpfsStopTimer()
	//strErrInfo := outerr.String()
	stdErr := outerr.String()
	if err != nil {
		log.Error("ipfs IpfsGetFileCache2ByHash get error", "error", err, "ipfs err", outerr.String())
		if strings.Index(stdErr, strIPFSstdErr) > 0 {
			CheckIpfsStatus(err)
		}
		gIpfsStat.gIPFSerrorNum++
		//CheckIpfsStatus(err)
		return nil, false, err
	}
	gIpfsStat.gIPFSerrorNum = 0
	tmpBlockFile, errf := os.OpenFile(objfileName, os.O_RDWR /*os.O_APPEND*/, 0644)

	return tmpBlockFile, false, errf
}
func (d *Downloader) IpsfAddNewBlockBatchToCache(stCfg *Cache1StoreCfg, blockList []listBlockInfo) error {
	var calArrayPos uint32 = 0
	var lastArrayPos uint32 = 65535000
	var tmpCache2File *os.File = nil
	var newFileFlg = false
	var err error
	var cache2st *Caches2CfgMap
	//var bhasStored = false
	//var bhasStored = false   "secondCacheInfo.gb"
	for _, tmpBlock := range blockList {
		err = nil
		newFileFlg = false
		if tmpBlock.blockNum < stCfg.OriginBlockNum {
			calArrayPos = 0
		} else {
			calArrayPos = uint32((tmpBlock.blockNum - stCfg.OriginBlockNum) / Cache2StoreHashMaxNum)
		}
		log.Trace("ipfs IpsfAddNewBlockBatchToCache calArrayPos", "calArrayPos", calArrayPos, "lastArrayPos", lastArrayPos)
		if calArrayPos >= Cache1StoreCache2Num {
			log.Error("ipfs error IpsfAddNewBlockBatchToCache calc error,calArrayPos exeeed capacity", "calArrayPos", calArrayPos)
			//return fmt.Errorf("         lArrayPos > Cache1StoreCache2Num")
			continue
		}
		if calArrayPos != lastArrayPos {
			if tmpCache2File != nil {

				storeCache(cache2st, tmpCache2File)
				tmpCache2File.Close()
				newHash, _, err1 := IpfsAddNewFile(strTmpCache2File, false) //"secondCacheInfo.gb"

				if err1 != nil {
					log.Error("ipfs error IpfsAddNewFile", "pos", lastArrayPos)
					continue
				}
				//cache1
				stCfg.CurrentBlockNum = tmpBlock.blockNum
				stCfg.Cache2FileNum++
				stCfg.StCahce2Hash[lastArrayPos] = string(newHash[0:IpfsHashLen])

				log.Debug("ipfs Debug IpsfAddNewBlockBatchToCache IpfsAddNewFile", "calArrayPos", calArrayPos, "lastArrayPos", lastArrayPos, "ipfs hash", stCfg.StCahce2Hash[lastArrayPos])
			}

			tmpCache2File, newFileFlg, err = IpfsGetFileCache2ByHash(stCfg.StCahce2Hash[calArrayPos], strTmpCache2File) //"secondCacheInfo.gb"
			if err != nil {
				tmpCache2File.Close()
				log.Error("ipfs IpsfAddNewBlockToCache use IpfsGetFileCache2ByHash error", "error", err)
				//return err
				continue
			}
			cache2st = new(Caches2CfgMap)
			if newFileFlg { //
				cache2st.CurCacheBeignNum = tmpBlock.blockNum
				cache2st.CurCacheBlockNum = tmpBlock.blockNum
				cache2st.NumHashStore = 0
				cache2st.MapList.Numberstore = make(map[uint64]NumberMapingCoupledHash)
				log.Trace("ipfs Debug IpsfAddNewBlockBatchToCache new(Caches2CfgMap)")

			} else {

				err := loadCache(cache2st, 0, tmpCache2File)
				if err != nil {
					log.Error("ipfs IpsfAddNewBlockBatchToCache loadCache error", "error", err)
					//tmpCache2File.Close()
					//return err
					cache2st.CurCacheBeignNum = tmpBlock.blockNum
					cache2st.CurCacheBlockNum = tmpBlock.blockNum
					cache2st.NumHashStore = 0
					cache2st.MapList.Numberstore = make(map[uint64]NumberMapingCoupledHash)
					log.Trace("ipfs Debug IpsfAddNewBlockBatchToCache loadCache error new(Caches2CfgMap)")
				}
				tmpCache2File.Close()
				tmpCache2File, _ = os.OpenFile(strTmpCache2File, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644) //"secondCacheInfo.gb"

				log.Trace("ipfs Debug IpsfAddNewBlockBatchToCache loadCache")
				//file.Close()
				//tgh = true
				//file, _ = os.OpenFile(strTmpCache2File, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644) //O_TRUNC

			}

		}
		//insert
		log.Trace("ipfs Debug IpsfAddNewBlockBatchToCache insertNewValue", "blockNum", tmpBlock.blockNum)

		//err1, newNumber
		for i, tmp := range tmpBlock.coinKind {
			newheadhash := tmp + ":" + tmpBlock.blockHeadHash //全block 为 0:headhash
			err1, _ := insertNewValue(tmpBlock.blockNum, newheadhash, tmpBlock.blockIpfshash[i], false, &cache2st.MapList)
			if err1 != nil {
				continue
			}
		}
		//if newNumber {
		cache2st.NumHashStore++
		cache2st.CurCacheBlockNum = tmpBlock.blockNum
		//}
		lastArrayPos = calArrayPos
		stCfg.CurrentBlockNum = tmpBlock.blockNum
	}
	gIpfsStoreCache.curBlockPos = calArrayPos
	//gIpfsStoreCache.storeipfsCache1 = stCfg

	storeCache(cache2st, tmpCache2File)
	tmpCache2File.Close()
	newHash, _, err := IpfsAddNewFile(strTmpCache2File, false)
	if err != nil {
		log.Error("ipfs IpsfAddNewBlockBatchToCache error IpfsAddNewFile", "err", err)
		return err
	}
	stCfg.Cache2FileNum++
	stCfg.StCahce2Hash[calArrayPos] = string(newHash[0:IpfsHashLen])
	log.Debug("ipfs IpsfAddNewBlockBatchToCache for over IpfsAddNewFile", "calArrayPos", calArrayPos, "ipfshash", string(newHash[0:IpfsHashLen]))
	err = WriteJsFile(path.Join(strCacheDirectory, strCache1BlockFile), stCfg) //firstCacheInfo.jn
	if err != nil {
		log.Error("ipfs IpsfAddNewBlockBatchToCache WriteJsFile error", "error", err)
		return err
	}

	//d.IPfsDirectoryUpdate()
	return nil
}

//IpsfAddNewBlockToCache
func (d *Downloader) IpsfAddNewBlockToCache(stCfg *Cache1StoreCfg, blockNum uint64, headHash string, fhash string) error {

	var calArrayPos uint32 = 0
	if blockNum < stCfg.OriginBlockNum {
		calArrayPos = 0
	} else {
		calArrayPos = uint32((blockNum - stCfg.OriginBlockNum) / Cache2StoreHashMaxNum)
	}
	log.Warn("ipfs IpsfAddNewBlockToCache1&2 ", "calArrayPos", calArrayPos, "Cache2FileNum ", stCfg.Cache2FileNum, "blockNum", blockNum)
	/*if calArrayPos-stCfg.Cache2FileNum > 1 {
		//return
		log.Error("ipfs error IpsfAddNewBlockToCache calc error", "calArrayPos", calArrayPos, "Cache2FileNum ", stCfg.Cache2FileNum)
	}*/
	gIpfsStoreCache.storeipfsCache1 = stCfg

	if calArrayPos >= Cache1StoreCache2Num {
		log.Error("ipfs error IpsfAddNewBlockToCache calc error,calArrayPos exeeed capacity", "calArrayPos", calArrayPos)
		return fmt.Errorf("calArrayPos > Cache1StoreCache2Num")
	}

	//var tmpBlockFile *os.File

	tmpBlockFile, newFileFlg, err := IpfsGetFileCache2ByHash(stCfg.StCahce2Hash[calArrayPos], strTmpCache2File) //stCfg.Cache2FileNum]) //.CurCachehash)
	if err != nil {
		tmpBlockFile.Close()
		log.Error("ipfs IpsfAddNewBlockToCache use IpfsGetFileCache2ByHash error", "error", err)
		return err
	}

	err = d.IpfsSyncSaveSecondCache(newFileFlg, blockNum, headHash, fhash, tmpBlockFile)
	if err != nil {
		tmpBlockFile.Close()
		log.Error("ipfs IpsfAddNewBlockToCache use IpfsSyncSaveSecondCache error", "error", err)
		return err
	}
	tmpBlockFile.Close()

	if logMap {
		log.Trace("`^^^^^^IpsfAddNewBlockToCache cache begin^^^^^^^```")
		ffile, _ := os.OpenFile(strTmpCache2File, os.O_RDWR, 0644)
		cache2st := new(Caches2CfgMap)
		cache2st.MapList.Numberstore = make(map[uint64]NumberMapingCoupledHash)
		loadCache(cache2st, 0, ffile)
		for key, value := range cache2st.MapList.Numberstore {

			log.Trace("SaveSecondCache info key", "key", key)
			for key2, value2 := range value.Blockhash {
				log.Trace("key-value", "key2", key2, "value:", value2, "k-v", cache2st.MapList.Numberstore[key].Blockhash[key2])
			}

		}
		ffile.Close()
		log.Trace("`^^^^^^cache end^^^^^^^```\n")

	}

	// add file
	newHash, _, err := IpfsAddNewFile(strTmpCache2File, false)
	if err != nil {
		log.Error("ipfs IpsfAddNewBlockToCache error IpfsAddNewFile", "err", err)
		return err
	}

	//cache1
	stCfg.CurrentBlockNum = blockNum
	stCfg.Cache2FileNum++
	stCfg.StCahce2Hash[calArrayPos] = string(newHash[0:IpfsHashLen])
	err = WriteJsFile(path.Join(strCacheDirectory, strCache1BlockFile), stCfg)
	if err != nil {
		log.Error("ipfs IpsfAddNewBlockToCache WriteJsFile error", "error", err)
		return err
	}

	return d.IPfsDirectoryUpdate()
	//return nil

}

//IpfsSyncSaveSecondCache
func (d *Downloader) IpfsSyncSaveSecondCache(newFlg bool, blockNum uint64, strheadHash string, fhash string, file *os.File) error {
	var cache2st = new(Caches2CfgMap)
	//var tgh bool = false

	if newFlg { //
		cache2st.CurCacheBeignNum = blockNum
		cache2st.CurCacheBlockNum = blockNum
		cache2st.NumHashStore = 0
		cache2st.MapList.Numberstore = make(map[uint64]NumberMapingCoupledHash)

	} else {

		err := loadCache(cache2st, 0, file)
		if err != nil {
			log.Error("ipfs IpfsSyncSaveSecondCache loadCache error", "error", err)
			file.Close()
			return err
		}

		file.Close()
		//tgh = true
		file, _ = os.OpenFile(strTmpCache2File, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644) //O_TRUNC

	}

	//insert
	err, newNumber := insertNewValue(blockNum, strheadHash, fhash, false, &cache2st.MapList)

	/*if cache2st.NumHashStore == 0 {
		cache2st.CurCacheBeignNum = blockNum
	}*/
	if newNumber {
		cache2st.NumHashStore++
		cache2st.CurCacheBlockNum = blockNum
	}
	if err == nil {
		//err2 := storeCache(cache2st, file)
		return storeCache(cache2st, file)
		//return file.Close()
	} else {
		//file.Close()
		return nil
	}

}

//IpfsSyncSaveLatestBlock
func (d *Downloader) IpfsSyncSaveLatestBlock() {
	//d.IPfsDirectoryUpdate()
}

//IPfsDirectoryUpdate
func (d *Downloader) IPfsDirectoryUpdate() error {
	var out bytes.Buffer
	var outerr bytes.Buffer

	c := exec.Command(d.dpIpfs.StrIPFSExecName, "add", "-Q", "-r", strCacheDirectory)
	c.Stdout = &out
	c.Stderr = &outerr
	//outbuf, err := c.Output()
	err := c.Run()
	//strErrInfo := outerr.String()
	if err != nil {
		log.Error("ipfs IPfsDirectoryUpdate add dictory error", "error", err, "ipfs err", outerr.String())
		RestartIpfsDaemon()
		c := exec.Command(d.dpIpfs.StrIPFSExecName, "add", "-Q", "-r", strCacheDirectory)
		err := c.Run()
		if err != nil {
			log.Error("ipfs IPfsDirectoryUpdate add dictory error again", "error", err)
			return err
		}
	}

	if out.Len() < IpfsHashLen {
		log.Error("ipfs IPfsDirectoryUpdate add dictory error again", "error", err, "out.Len() ", out.Len(), "ipfs err", outerr.String())
		return err
	}
	publishHash := string(out.Bytes()[0 : out.Len()-1]) //0：IpfsHashLen

	out.Reset()
	outerr.Reset()
	//run
	c = exec.Command(d.dpIpfs.StrIPFSExecName, "name", "publish", publishHash) //string(outbuf[:]))
	IpfsStartTimer(publishHash)
	err = c.Run()
	IpfsStopTimer()
	strErrInfo := outerr.String()

	if err != nil {
		log.Error("ipfs IPfsDirectoryUpdate name publish error", "error", err, "publish", publishHash, "ipfs err", strErrInfo)
		return err
	}
	return nil
	//IPFS  name  publish 	// ipfs name resolve  + peerID
}

//IpfsSyncGetFirstCache
func (d *Downloader) IpfsSyncGetFirstCache(index int) (*Cache1StoreCfg, error) {

	//var out bytes.Buffer
	var outerr bytes.Buffer
	//ipnsPath := "/ipns/" + d.dpIpfs.StrIpfspeerID + "/" + strCache1BlockFile //"firstCacheSync.ha"
	ipnsPath := "/ipns/" + listPeerId[index] + "/" + strCache1BlockFile
	c := exec.Command(d.dpIpfs.StrIPFSExecName, "cat", ipnsPath) //或cat
	//c.Stdout = &out
	//go ipfsGetTimeout()
	IpfsStartTimer("first")
	c.Stderr = &outerr
	//err := c.Run()
	outbuf, err := c.Output()
	//runQuit <- 1
	IpfsStopTimer()
	//strErrInfo := outerr.String()

	//new
	curCache1Info := new(Cache1StoreCfg) // Cache1StoreCfg{}
	stdErr := outerr.String()
	if err != nil {
		log.Error("ipfs error IpfsSyncGetFirstCache error", "error", err, "ipfs err", outerr.String())
		if strings.Index(stdErr, strIPFSstdErr) > 0 || strings.Index(stdErr, strIPFSstd2Err) > 0 {
			CheckIpfsStatus(err)
		}
		gIpfsStat.gIPFSerrorNum++
		//CheckIpfsStatus(err)
		d.dealIPFSerrorProc()
		return curCache1Info, err
	}
	//ReadJsFile()
	gIpfsStat.gIPFSerrorNum = 0

	err = json.Unmarshal(outbuf, curCache1Info)
	if err != nil {
		log.Error("ipfs IpfsSyncGetFirstCache json.Unmarshal error", "error", err)
		return curCache1Info, err
	}
	return curCache1Info, nil
}

//IpfsSyncGetSecondCache
func (d *Downloader) IpfsSyncGetSecondCache() {

}

//IpfsSyncGetLatestBlock
func (d *Downloader) IpfsSyncGetLatestBlock(index int) (*LastestBlcokCfg, uint64, error) {
	//var out bytes.Buffer
	var outerr bytes.Buffer
	//ipnsPath := "/ipns/" + d.dpIpfs.StrIpfspeerID + "/" +  //"lastestblock.ha"
	ipnsPath := "/ipns/" + listPeerId[index] + "/" + strLastestBlockFile
	//out.Reset()

	//log.Error("ipfs IpfsSyncGetLatestBlock run cmd before", "listPeerId", listPeerId[index])

	c := exec.Command(gIpfsPath, "get", "-o=ipfsCachecommon/", ipnsPath) //
	//c.Stdout = &out
	c.Stderr = &outerr
	err := c.Run()
	//outbuf, err := c.Output()
	//log.Error("ipfs IpfsSyncGetLatestBlock run cmd after")

	//strErrInfo := outerr.String()
	stdErr := outerr.String()
	curLastestInfo := new(LastestBlcokCfg) //LastestBlcokCfg{}
	if err != nil {

		log.Error("ipfs IpfsSyncGetLatestBlock run cmd error", "error", err, "ipfs err", outerr.String())
		if strings.Index(stdErr, strIPFSstdErr) > 0 {
			CheckIpfsStatus(err)
		}
		//CheckIpfsStatus(err)
		return curLastestInfo, 0, err
	}

	tmpFile, errf := os.Open(path.Join(strCacheDirectory, strLastestBlockFile)) //os.Open(strLastestBlockFile)
	defer tmpFile.Close()
	if errf != nil {
		log.Error("ipfs IpfsSyncGetLatestBlock OpenFile error", "error", errf)
		return curLastestInfo, 0, err
	}

	//curLastestInfo := LastestBlcokCfg{}
	err = loadCache(curLastestInfo, 0, tmpFile)
	if err != nil {
		log.Error("ipfs IpfsSyncGetLatestBlock loadCache json.Unmarshal error", "error", err)
		return curLastestInfo, 0, err
	}

	return curLastestInfo, curLastestInfo.CurrentNum, nil

}

//insertNewValue
func insertNewValue(blockNum uint64, headHash string /*common.Hash*/, strblockhash string, coverflg bool, newBlock *BlockStore) (error, bool) {
	var BnumberNoExist bool = false

	if newBlock.Numberstore == nil {
		newBlock.Numberstore = make(map[uint64]NumberMapingCoupledHash) //
	}
	_, ok := (newBlock.Numberstore)[blockNum]
	if ok == false {
		//newBlock.Numberstore = make(map[uint64]NumberMapingCoupledHash)
		var tmpH NumberMapingCoupledHash
		tmpH.Blockhash = make(map[string]string)
		(newBlock.Numberstore)[blockNum] = tmpH
		//log.Trace("insertNewValue  not exist head hash", "blocknum", blockNum)
		BnumberNoExist = true
	} else {
		strHash, ok := (newBlock.Numberstore)[blockNum].Blockhash[headHash]
		if ok {

			log.Warn("insertNewValue  Blockhash[headHash] value exist ", "blocknum", blockNum)
			//return fmt.Errorf("block hash already exist")
			if coverflg == true { //链接原来hash
				strblockhash = strHash + strblockhash
			}
		}
	}
	newBlock.Numberstore[blockNum].Blockhash[headHash] = strblockhash //string(blockhash[0:IpfsHashLen])
	//log.Trace("ipfs insertNewValue  head hash", "blocknum", blockNum, "headHash", headHash, "blockhash", strblockhash) //string(blockhash[0:IpfsHashLen]))

	return nil, BnumberNoExist
}

func (d *Downloader) addNewBlockBatchToLastest(curLastestInfo *LastestBlcokCfg, blockList []listBlockInfo) error {
	for _, tmpBlock := range blockList {
		curLastestInfo.CurrentNum = tmpBlock.blockNum
		//最近的不考虑多币种
		_, newNumber := insertNewValue(tmpBlock.blockNum, tmpBlock.blockHeadHash, tmpBlock.blockIpfshash[0], false, &curLastestInfo.MapList) //
		if newNumber {
			curLastestInfo.HashNum++
		}
		if curLastestInfo.HashNum > LastestBlockStroeNum {
			curLastestInfo.HashNum = LastestBlockStroeNum
			//delete
			delete(curLastestInfo.MapList.Numberstore, uint64(curLastestInfo.CurrentNum-LastestBlockStroeNum))
		}
		log.Trace("ipfs addNewBlockBatchToLastest insertNewValue  head hash", "blocknum", tmpBlock.blockNum)
	}
	//
	tmpFile, errf := os.OpenFile(path.Join(strCacheDirectory, strLastestBlockFile), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644) //"lastestblockInfo.gb"
	if errf != nil {
		return errf
	}
	err := storeCache(curLastestInfo, tmpFile)
	tmpFile.Close()
	/*
		if err == nil {
			return d.IPfsDirectoryUpdate()
		}*/
	return err

}

//IpfsSynInsertNewBlockHashToLastest
func (d *Downloader) IpfsSynInsertNewBlockHashToLastest(curLastestInfo *LastestBlcokCfg, blockNum uint64, strheadHash string, blockhash string) error {
	//
	//curLastestInfo := LastestBlcokCfg{}
	//loadCache(&curLastestInfo, 0)
	/*err := ReadJsFile(strLastestBlockFile, &curLastestInfo)
	if err != nil {

		return err
	}*/
	//
	/*if blockNum <= curLastestInfo.CurrentNum {
		log.Error("ipfs error IpfsSynInsertNewBlockToLastestHash blockNum already exist ", "blockNum=", blockNum)
		return err
	}*/

	//curLastestInfo.ListHash[curLastestInfo.HashNum].blockNum = blockNum
	//curLastestInfo.ListHash[curLastestInfo.HashNum].blochHash = string(blockhash[:])
	//
	/* 1
	if curLastestInfo.HashNum < LastestBlockStroeNum {
		curLastestInfo.HashNum++
		curLastestInfo.CurrentNum = blockNum

		WriteJsFile(strLastestBlockFile, curLastestInfo)
	} else {
		//
		tmpStroe := LastestBlcokCfg{
			CurrentNum: blockNum,
			HashNum:    LastestBlockStroeNum, //100
			ListHash:   curLastestInfo.ListHash[1:],
		}
		err = WriteJsFile(strLastestBlockFile, tmpStroe)
	}*/

	// 2
	/*curLastestInfo.HashNum++
	curLastestInfo.StrHash[(curLastestInfo.HashNum)%(LastestBlockStroeNum)] = string(blockhash[:])
	err = WriteJsFile(strLastestBlockFile, curLastestInfo)*/
	//3

	curLastestInfo.CurrentNum = blockNum
	err, newNumber := insertNewValue(blockNum, strheadHash, blockhash, false, &curLastestInfo.MapList) //
	if newNumber {
		curLastestInfo.HashNum++
	}
	if curLastestInfo.HashNum > LastestBlockStroeNum {
		curLastestInfo.HashNum = LastestBlockStroeNum
		//delete
		delete(curLastestInfo.MapList.Numberstore, uint64(curLastestInfo.CurrentNum-LastestBlockStroeNum))
	}
	//
	tmpFile, errf := os.OpenFile(path.Join(strCacheDirectory, strLastestBlockFile), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if errf != nil {
		return errf
	}
	err = storeCache(curLastestInfo, tmpFile)

	tmpFile.Close() //关闭文件 再发布

	if logMap {
		log.Trace("********IpfsSynInsertNewBlockHashToLastest lastest begin *********```\n")
		ffile, _ := os.OpenFile(path.Join(strCacheDirectory, strLastestBlockFile), os.O_RDWR, 0644)
		cache2st := new(LastestBlcokCfg)
		cache2st.MapList.Numberstore = make(map[uint64]NumberMapingCoupledHash)
		loadCache(cache2st, 0, ffile)
		for key, value := range cache2st.MapList.Numberstore {
			//fmt.Println(i.CurrentNum, i.HashNum)
			log.Trace("Lastest curLastestInfo key", "key", key)
			for key2, value2 := range value.Blockhash {
				log.Trace("key-value", "key2", key2, "value:", value2, "value2", cache2st.MapList.Numberstore[key].Blockhash[key2])
			}

		}
		ffile.Close()
		log.Trace("```lastest end`````\n")
	}

	if err == nil {
		return d.IPfsDirectoryUpdate()
	}
	return err

}

//RecvBlockToDeal
func (d *Downloader) RecvBlockSaveToipfs(blockqueue *prque.Prque) error {
	// block number
	//curBlockNum := newBlock.NumberU64()
	var err error
	var stCache1Cfg *LastestBlcokCfg
	log.Warn("~~~~~ ipfs RecvBlockToDeal recv new block ~~~~~~", "blockNumLenSzie=", blockqueue.Size())
	if SingleBlockStore == true {
		stCache1Cfg, _, err = d.IpfsSyncGetLatestBlock(0) //"lastestblockInfo.gb"

		if err != nil {

			//if curBlockNum < hasBlocbNum {
			//	log.Warn("ipfs warn RecvBlockToDeal blockNum already exist ", "blockNum=", curBlockNum)
			//return err
			//}
		}
	}
	bNeedBatch := false
	bNeedSanp := false
	var tmpsnap types.SnapSaveInfo
	curlistBlockInfo := make([]listBlockInfo, 0)
	d.blockchain.GetIpfsQMux()
	for {
		var tmplistBlockInfo listBlockInfo //:= new(listBlockInfo)
		if blockqueue.Empty() {
			break
		}

		//stBlock := blockqueue.PopItem().(*types.BlockAllSt) //(*types.Block)
		revInfo := blockqueue.PopItem()
		switch revInfo.(type) {
		case *types.BlockAllSt:

		case types.SnapSaveInfo:
			bNeedSanp = true
			tmpsnap = revInfo.(types.SnapSaveInfo)
			//tmpsnap := revInfo.(types.SnapSaveInfo)
			///d.AddStateRootInfoToIpfs(tmpsnap.BlockNum, tmpsnap.BlockHash, tmpsnap.SnapPath)
			continue
		default:
			continue
		}
		stBlock := revInfo.(*types.BlockAllSt)
		newBlock := stBlock.Sblock
		if newBlock == nil {
			continue
		}
		tmplistBlockInfo.coinKind = make([]string, 0, 8)
		tmplistBlockInfo.blockIpfshash = make([]string, 0, 8)
		tmplistBlockInfo.blockNum = newBlock.NumberU64()
		if tmplistBlockInfo.blockNum == 1 { //说明要重新开始
			d.ClearDirectoryContent()
		}

		//可注掉，目前为统计-begin
		{
			tmpBlockFile, errf := os.OpenFile(strNewBlockStoreFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644) //"NewTmpBlcok.rp"
			if errf != nil {
				log.Error("ipfs RecvBlockToDeal error in open file ", "error=", errf)
				return errf
			}

			//errd := rlp.Encode(tmpBlockFile, newBlock)
			errd := rlp.Encode(tmpBlockFile, stBlock)
			tmpBlockFile.Close()
			{
				fhandler, _ := os.Stat(strNewBlockStoreFile)
				gIpfsStat.totalBlockSize += fhandler.Size()
				log.Trace("ipfs block encode info", "error", errd, "blockNum", tmplistBlockInfo.blockNum, "blockSize", fhandler.Size(), "totalSize", gIpfsStat.totalBlockSize)
			}
		}
		//可注掉，目前为统计-end
		err = nil

		if SingleBlockStore == true {
			////增加压缩区块
			hashs, zipSize, err := IpfsAddNewFile(strNewBlockStoreFile, true)
			if err != nil {
				log.Error("ipfs RecvBlockToDeal error IpfsAddNewFile  ", "error=", err)

				// 先不返回，便于后面写进batch  return err
			} else {
				//待增加多币种
				tmplistBlockInfo.coinKind = append(tmplistBlockInfo.coinKind, "0") //0：默认去不区块类型   map变为 coinkind：+headhash
				tmplistBlockInfo.blockIpfshash = append(tmplistBlockInfo.blockIpfshash, string(hashs[0:IpfsHashLen]))
				//tmpHash := newBlock.Hash()
				tmplistBlockInfo.blockHeadHash = newBlock.Hash().String()     //string(tmpHash[:])
				curlistBlockInfo = append(curlistBlockInfo, tmplistBlockInfo) //将当前block 信息增加到待处理批量结构中
				gIpfsStat.totalZipBlockSize += zipSize
				log.Debug("ipfs block encode IpfsAddNewFile later", "error", "blockZipSize", zipSize, "totalZipSize", gIpfsStat.totalZipBlockSize)
			}
		}

		{ //增加到批量区块存储文件中
			d.BatchStoreAllBlock(stBlock)
		}
		if tmplistBlockInfo.blockNum%BATCH_NUM == 0 {
			log.Debug("ipfs RecvBlockToDeal get mod 300 =0")
			bNeedBatch = true //批量区块存储文件 上传标记
			break
		}
	}
	d.blockchain.GetIpfsQUnMux()
	if bNeedSanp == true {
		d.AddStateRootInfoToIpfs(tmpsnap.BlockNum, tmpsnap.BlockHash, tmpsnap.SnapPath)
	}

	if SingleBlockStore == true {
		if len(curlistBlockInfo) == 0 {
			return fmt.Errorf("curlistBlockInfo len = 0")
		}

		d.addNewBlockBatchToLastest(stCache1Cfg, curlistBlockInfo)
		//d.IPfsDirectoryUpdate()
		//go func() error {
		dealcacheFunc := func() error {
			//readCacheCg := Cache1StoreCfg{}	//readCache =(*Cache1StoreCfg)readCacheCfg
			if gIpfsStoreCache.storeipfsCache1 == nil {
				readCacheCfg, err := d.IpfsSyncGetFirstCache(0) //"firstCacheInfo.jn"
				if err != nil {
					fmt.Println("cache1 is nil, create it", err)
					log.Debug("ipfs RecvBlockToDeal IpfsSyncGetFirstCache  cache1 is nil, create it  ", "error=", err)
					//readCacheCfg2 := Cache1StoreCfg{}
					readCacheCfg.OriginBlockNum = curlistBlockInfo[0].blockNum
					readCacheCfg.CurrentBlockNum = curlistBlockInfo[0].blockNum
					readCacheCfg.Cache2FileNum = 0
					//获取失败从本地文件载入
					ReadJsFile(path.Join(strCacheDirectory, strCache1BlockFile), readCacheCfg)
					log.Debug("ipfs RecvBlockToDeal IpfsSyncGetFirstCache  load localfile", "OriginBlockNum=", readCacheCfg.OriginBlockNum, "CurrentBlockNum=", readCacheCfg.CurrentBlockNum)
				}
				gIpfsStoreCache.storeipfsCache1 = readCacheCfg
			}

			return d.IpsfAddNewBlockBatchToCache(gIpfsStoreCache.storeipfsCache1, curlistBlockInfo)

		} //()
		err = dealcacheFunc()
		//err = d.IPfsDirectoryUpdate()
		if err == nil {
			log.Trace("ipfs RecvBlockToDeal add ipfs sucess")
		} else {
			log.Error("ipfs RecvBlockToDeal add ipfs error ")
		}
	}
	if bNeedBatch == true {
		d.AddNewBatchBlockToIpfs()
		if SingleBlockStore == false {
			err = d.IPfsDirectoryUpdate()
		}
	}
	if SingleBlockStore == true {
		err = d.IPfsDirectoryUpdate()
	}
	testShowlog++

	//if testShowlog == 6 || testShowlog == 18 || testShowlog == 30 || testShowlog == 50 || testShowlog == 70 {
	if testShowlog == 0 {
		log.Trace("********lastest begin *********```\n")
		ffile, _ := os.OpenFile(path.Join(strCacheDirectory, strLastestBlockFile), os.O_RDWR, 0644)
		lastest := new(LastestBlcokCfg)
		lastest.MapList.Numberstore = make(map[uint64]NumberMapingCoupledHash)
		loadCache(lastest, 0, ffile)
		for key, value := range lastest.MapList.Numberstore {
			//fmt.Println(i.CurrentNum, i.HashNum)
			log.Trace("Lastest curLastestInfo key", "key", key)
			for key2, value2 := range value.Blockhash {
				log.Trace("key-value", "key2", key2, "value:", value2, "value2", lastest.MapList.Numberstore[key].Blockhash[key2])
			}
		}
		ffile.Close()
		log.Trace("```lastest end`````\n")

		//ffile, _ = os.OpenFile(path.Join(strCacheDirectory, strCache1BlockFile), os.O_RDWR, 0644)
		log.Trace("```cache 1 & 2 begin`````\n")
		tmpCache1 := new(Cache1StoreCfg)
		ReadJsFile(path.Join(strCacheDirectory, strCache1BlockFile), tmpCache1)

		log.Trace("tmpCache1", "OriginBlockNum", tmpCache1.OriginBlockNum, "CurrentBlockNum", tmpCache1.CurrentBlockNum, "Cache2FileNum", tmpCache1.Cache2FileNum)
		for idx := 0; idx < int(tmpCache1.Cache2FileNum); idx++ {
			log.Trace("tmpCache1 ", "index", idx, "hash", tmpCache1.StCahce2Hash[idx])
			tmpCache2File, _, _ := IpfsGetFileCache2ByHash(tmpCache1.StCahce2Hash[idx], strTmpCache2File) //secondCacheInfo.gb
			cache2st := new(Caches2CfgMap)
			loadCache(cache2st, 0, tmpCache2File)
			for key, value := range cache2st.MapList.Numberstore {
				//fmt.Println(i.CurrentNum, i.HashNum)
				log.Trace("cache2 secondCacheInfo.gb key", "key", key)
				for key2, value2 := range value.Blockhash {
					log.Trace("key-value", "key2", key2, "value:", value2, "value2", cache2st.MapList.Numberstore[key].Blockhash[key2])
				}
			}
			tmpCache2File.Close()

		}
		log.Trace("```cache 1 & 2 end`````\n")
	}

	return nil
}

//RecvBlockToDeal
func (d *Downloader) RecvBlockToDeal(newBlock *types.Block) error {
	// block number
	curBlockNum := newBlock.NumberU64()

	log.Warn("~~~~~ ipfs RecvBlockToDeal recv new block ~~~~~~", "blockNum=", curBlockNum)

	stCache1Cfg, hasBlocbNum, err := d.IpfsSyncGetLatestBlock(0)

	if err != nil {

		if curBlockNum < hasBlocbNum {
			log.Warn("ipfs warn RecvBlockToDeal blockNum already exist ", "blockNum=", curBlockNum)
			//return err
		}
	}
	tmpBlockFile, errf := os.OpenFile(strNewBlockStoreFile, os.O_WRONLY|os.O_CREATE, 0644)
	if errf != nil {
		log.Error("ipfs RecvBlockToDeal error in open file ", "error=", errf)
		return errf
	}

	errd := rlp.Encode(tmpBlockFile, newBlock)
	log.Debug("ipfs block encode info", "error", errd)
	tmpBlockFile.Close()

	bHash, _, err := IpfsAddNewFile(strNewBlockStoreFile, true)
	if err != nil {
		log.Error("ipfs RecvBlockToDeal error IpfsAddNewFile  ", "error=", errf)
		return err
	}

	headHash := newBlock.Hash().String()

	d.IpfsSynInsertNewBlockHashToLastest(stCache1Cfg, curBlockNum, headHash, string(bHash[0:IpfsHashLen]))

	//go func() error {
	dealcacheFunc := func() error {
		//readCacheCg := Cache1StoreCfg{}	//readCache =(*Cache1StoreCfg)readCacheCfg
		readCacheCfg, err := d.IpfsSyncGetFirstCache(0)
		if err != nil {
			fmt.Println("cache1 is nil, create it", err)
			log.Debug("ipfs RecvBlockToDeal IpfsSyncGetFirstCache  cache1 is nil, create it  ", "error=", err)
			//readCacheCfg2 := Cache1StoreCfg{}
			readCacheCfg.OriginBlockNum = curBlockNum
			readCacheCfg.CurrentBlockNum = curBlockNum
			readCacheCfg.Cache2FileNum = 0
			//获取失败从本地文件载入
			ReadJsFile(path.Join(strCacheDirectory, strCache1BlockFile), readCacheCfg)
			log.Debug("ipfs RecvBlockToDeal IpfsSyncGetFirstCache  load localfile", "OriginBlockNum=", readCacheCfg.OriginBlockNum, "CurrentBlockNum=", readCacheCfg.CurrentBlockNum)

		}
		gIpfsStoreCache.storeipfsCache1 = readCacheCfg
		return d.IpsfAddNewBlockToCache(readCacheCfg, curBlockNum, headHash, string(bHash[0:IpfsHashLen])) /* &readCacheCfg,*/

	} //()
	err = dealcacheFunc()
	if err == nil {
		log.Debug("ipfs RecvBlockToDeal add ipfs sucess", "block", curBlockNum)
	} else {
		log.Error("ipfs RecvBlockToDeal add ipfs error ", "block", curBlockNum)
	}

	return nil
}
func (d *Downloader) GetBlockAndAnalysisSend(blockhash string, stype string) bool {
	blockFile, err := IpfsGetBlockByHash(blockhash, true)
	//解压区块
	//
	defer func() {
		blockFile.Close()
		os.Remove(blockhash + ".unzip")
	}()
	if err != nil {
		log.Debug(" ipfs GetBlockAndAnalysis error in IpfsGetBlockByHash", "error", err)
		return false
	}
	//
	obj := new(types.BlockAllSt) //types.Block)
	errd := rlp.Decode(blockFile, obj)

	log.Info("ipfs dencode block info from GetBlockAndAnalysis", "err", errd, "stype", stype, "obj.Header", obj.Sblock.NumberU64())

	if errd != nil {
		return false
	}
	d.SynOrignDownload(obj, 0, 0)
	return true
}

func (d *Downloader) GetBlockHashFromCache(headhash string /*common.Hash*/, coinstr string, headNumber uint64) bool {
	//strHeadHash := string(headhash[:])
	if gIpfsCache.lastestCache == nil {
		log.Info("ipfs GetBlockHashFromCache lastestCache = nil")
		return false
	}

	if gIpfsCache.lastestNum-headNumber < LastestBlockStroeNum {

		_, ok := gIpfsCache.lastestCache.MapList.Numberstore[headNumber]
		if ok {
			// block hash
			blockhash, ok := gIpfsCache.lastestCache.MapList.Numberstore[headNumber].Blockhash[headhash]
			if ok {
				log.Info("ipfs GetBlockHashFromCache download block from lastest Cache")
				return d.GetBlockAndAnalysisSend(blockhash, "lastestCache")
			}
		}
	}

	if gIpfsCache.getipfsCache2 == nil {
		log.Info("ipfs GetBlockHashFromCache getipfsCache2 = nil")
		return false
	}
	newheadHash := coinstr + ":" + headhash
	if headNumber >= gIpfsCache.getipfsCache2.CurCacheBeignNum && gIpfsCache.getipfsCache2.CurCacheBlockNum >= headNumber {
		//map
		_, ok := gIpfsCache.getipfsCache2.MapList.Numberstore[headNumber]
		if ok {
			blockhash, ok := gIpfsCache.getipfsCache2.MapList.Numberstore[headNumber].Blockhash[newheadHash]
			if ok {
				log.Info("ipfs GetBlockHashFromCache download block from Cache2")
				return d.GetBlockAndAnalysisSend(blockhash, "secondCache")
			} else {
				log.Error("ipfs GetBlockHashFromCache  error map Blockhash", "headNumber", headNumber, "headhash", newheadHash)
				return false
			}
		} else {
			log.Error("ipfs GetBlockHashFromCache  error map ", "headNumber", headNumber)
			return false
		}
	}
	log.Warn("ipfs GetBlockHashFromCache  no find ", "lastnum", gIpfsCache.lastestNum)
	return false
}

//下载区块
// SyncBlockFromIpfs
func (d *Downloader) SyncBlockFromIpfs(strHeadHash string, headNumber uint64, coin string, index int) int {
	//curBlock := (*core.BlockChain).CurrentBlock()
	//CurLocalBlocknum := d.blockchain.CurrentBlock().NumberU64()
	//strHeadHash := string(headhash[:])
	log.Debug(" *** ipfs get download block ***  ", "number", headNumber, "headhash", strHeadHash, "gIpfsCache.bassign", gIpfsCache.bassign)
	if gIpfsCache.bassign {
		//bfind := d.GetBlockHashFromCache(coin+":"+strHeadHash, headNumber)
		bfind := d.GetBlockHashFromCache(strHeadHash, coin, headNumber)
		if bfind {
			return 0
		}
	}

	log.Debug(" ****** ipfs get download block number over ipfs  ******  ", "number", headNumber)
	var err error

	if coin != "0" {
		goto secondCache
	}
	gIpfsCache.lastestCache, gIpfsCache.lastestNum, err = d.IpfsSyncGetLatestBlock(index)
	if err != nil {

		log.Debug(" ipfs  SyncBlockFromIpfs error in  IpfsSyncGetFirstCache")
		goto secondCache
		//return
	}
	if gIpfsCache.lastestNum <= headNumber {
		log.Debug(" ipfs SyncBlockFromIpfs It is no need to update", "CurLocalBlocknum", headNumber, "ipfsLastestBlockNum", gIpfsCache.lastestNum)
		//return
	}
	if logMap {
		for key, value := range gIpfsCache.lastestCache.MapList.Numberstore {
			//fmt.Println(i.CurrentNum, i.HashNum)
			log.Trace("Lastest curLastestInfo key", "key", key, "gIpfsCache.lastestNum", gIpfsCache.lastestNum)
			//if key == headNumber {
			for key2, value2 := range value.Blockhash {
				log.Debug("key-value", "key2:", key2, "value2", value2, "value2", gIpfsCache.lastestCache.MapList.Numberstore[key].Blockhash[key2])
			}
		}
	}
	gIpfsCache.bassign = true
	if gIpfsCache.lastestNum-headNumber < LastestBlockStroeNum {

		if logMap {
			log.Debug("********linshi lastest lastest begin *********```\n")
			//for key, value := range gIpfsCache.getipfsCache2.MapList.Numberstore[headNumber] {
			for key, value := range gIpfsCache.lastestCache.MapList.Numberstore {
				//fmt.Println(i.CurrentNum, i.HashNum)
				//log.Trace("Lastest curLastestInfo key", "key", key)
				if key == headNumber {
					for key2, value2 := range value.Blockhash {
						log.Debug("key-value", "key2:", key2, "value2", value2, "value2", gIpfsCache.lastestCache.MapList.Numberstore[key].Blockhash[key2])
					}
				}

			}

			log.Debug("```linshi lastest lastest end`````\n")
		}
		_, ok := gIpfsCache.lastestCache.MapList.Numberstore[headNumber]
		if ok {
			//  block hash
			blockhash, ok := gIpfsCache.lastestCache.MapList.Numberstore[headNumber].Blockhash[strHeadHash]
			if ok {

				log.Debug(" ipfs  SyncBlockFromIpfs get new block form ipfs use by getlastest", "blockNum", headNumber)
				bsec := d.GetBlockAndAnalysisSend(blockhash, "getlastest")
				if bsec {
					return 0
				}

			} else {
				log.Debug(" ipfs  SyncBlockFromIpfs get new block form error", "headnumber", headNumber, "headhash", strHeadHash)
			}

		} else {
			log.Debug(" ipfs  SyncBlockFromIpfs get new blockerror", "headnumber", headNumber)
		}

	} /*else*/

secondCache:

	{
		log.Debug(" ipfs  SyncBlockFromIpfs begin in IpfsSyncGetFirstCache")
		err = nil
		gIpfsCache.getipfsCache1, err = d.IpfsSyncGetFirstCache(index)
		//readCache =(*Cache1StoreCfg)readCacheCfg//readCache1Info
		if err != nil {
			log.Error(" ipfs  SyncBlockFromIpfs error in IpfsSyncGetFirstCache")
			return 1
		}
		//
		if headNumber >= gIpfsCache.getipfsCache1.CurrentBlockNum {
			log.Warn(" ipfs error SyncBlockFromIpfs IpfsSyncGetFirstCache block num", "CurLocalBlocknum", headNumber, "readCache1Info.CurrentBlockNum", gIpfsCache.getipfsCache1.CurrentBlockNum)
			//return
		}
		//
		if headNumber < gIpfsCache.getipfsCache1.OriginBlockNum {
			log.Error(" ipfs  SyncBlockFromIpfs   OriginBlockNum error", "headNumber", headNumber, "OriginBlockNum", gIpfsCache.getipfsCache1.OriginBlockNum)
			return 2
		}
		arrayIndex := uint32((headNumber - gIpfsCache.getipfsCache1.OriginBlockNum) / Cache2StoreHashMaxNum) //
		if arrayIndex >= Cache1StoreCache2Num {
			log.Error(" ipfs error SyncBlockFromIpfs exceed capacity", "arrayIndex", arrayIndex, "Cache1StoreCache2Num", Cache1StoreCache2Num)
			return 2
		}
		stCache2Infohash := gIpfsCache.getipfsCache1.StCahce2Hash[arrayIndex]
		//
		cache2File, err := IpfsGetBlockByHash(stCache2Infohash, false)
		if err != nil {
			log.Error(" ipfs  SyncBlockFromIpfs error IpfsGetBlockByHash cache2File", "error", err)
			return 1
		}
		defer func() {
			cache2File.Close()
			os.Remove(stCache2Infohash)
		}()

		//cache2st := new(Caches2CfgMap) //eof
		gIpfsCache.getipfsCache2 = new(Caches2CfgMap)
		//for {
		err = loadCache(gIpfsCache.getipfsCache2, 0, cache2File)
		if err != nil {
			log.Error("ipfs SyncBlockFromIpfs loadCache error", "error", err)
			return 1
		}

		gIpfsCache.bassign = true
		//
		if logMap {
			log.Debug("********linshi lastest begin *********```\n")
			//for key, value := range gIpfsCache.getipfsCache2.MapList.Numberstore[headNumber] {
			for key, value := range gIpfsCache.getipfsCache2.MapList.Numberstore {
				//fmt.Println(i.CurrentNum, i.HashNum)
				log.Trace("Lastest curLastestInfo key", "key", key)
				//if key == headNumber
				{
					for key2, value2 := range value.Blockhash {
						log.Debug("key-value", "key2:", key2, "value:", value2, "value2", gIpfsCache.getipfsCache2.MapList.Numberstore[key].Blockhash[key2])
					}
				}

			}

			log.Debug("```linshi lastest end`````\n")
		}

		_, ok := gIpfsCache.getipfsCache2.MapList.Numberstore[headNumber]
		newStrHeadhash := coin + ":" + strHeadHash
		if ok {
			blockhash, ok := gIpfsCache.getipfsCache2.MapList.Numberstore[headNumber].Blockhash[newStrHeadhash]
			if ok {

				log.Debug(" ipfs  SyncBlockFromIpfs get new block form ipfs use by getCache2", "blockNum", headNumber, "newStrHeadhash", newStrHeadhash)
				d.GetBlockAndAnalysisSend(blockhash, "getCache2")
				return 0
			} else {
				log.Error("ipfs  SyncBlockFromIpfs  error map Blockhash", "headNumber", headNumber, "headhash", strHeadHash)
				return 1
			}
		} else {
			log.Error("ipfs  SyncBlockFromIpfs  error map ", "headNumber", headNumber)
			return 1
		}

	}
	//return 1
	//SynTest()
}
func (d *Downloader) SynOrignDownload(out interface{}, flag int, blockNum uint64) { //(obj *types.Block) {

	tmp := new(BlockIpfs)
	tmp.Flag = flag
	tmp.BlockNum = blockNum
	switch flag {
	case 0:
		obj := out.(*types.BlockAllSt)
		/*txs := make(types.CurrencyBlock,0)//make(types.SelfTransactions,0)
		res := make(types.Receipts,0)
		for _,cub := range obj.Sblock.Currencies(){
			txs = append(txs,cub.Transactions.GetTransactions()...)
			res = append(res,cub.Receipts.GetReceipts()...)
		}
		cointxs,coinres := types.GetCoinTXRS(txs,res)*/
		tmp.Headeripfs = obj.Sblock.Header()
		tmp.Transactionsipfs = obj.Sblock.Currencies()
		tmp.Unclesipfs = obj.Sblock.Uncles()
		//tmp.Receipt = coinres
		tmp.BlockNum = tmp.Headeripfs.Number.Uint64() //blockNum
		log.Debug(" ipfs send new block to syn BlockAllSt ", "flag", flag, "blockNum", tmp.Headeripfs.Number.Uint64())
	case 1:
	case 2:
		/*
			obj := out.(*types.Body)
			// txs := make(types.SelfTransactions,0)
			// for _,cub := range obj.CurrencyBody{
			// 	txs = append(txs,cub.Transactions.GetTransactions()...)
			// }
			// cointxs:= types.GetCoinTX(txs)
			tmp.Transactionsipfs = obj.CurrencyBody
			tmp.Unclesipfs = obj.Uncles
			log.Debug(" ipfs send new block to syn Body", "flag", flag, "blockNum", blockNum)*/
		obj := out.(*types.Block)
		tmp.Headeripfs = obj.Header()
		tmp.Transactionsipfs = obj.Currencies()
		tmp.Unclesipfs = obj.Uncles()
		log.Debug(" ipfs send new block to syn Body", "flag", flag, "blockNum", blockNum)

	case 3:
		obj := out.([]types.CoinReceipts)
		tmp.Receipt = obj
		log.Debug(" ipfs send new block to syn Receipts", "flag", flag, "blockNum", blockNum)

	case 33: //通知删除请求队列

	}

	//log.Debug(" ipfs send new block to syn", "number=%d", tmp.Headeripfs.Number.Uint64(), "flag", flag, "blockNum", blockNum)

	d.ipfsBodyCh <- *tmp
}

//区块同步
func (d *Downloader) IpfsProcessRcvHead() {

	log.Debug(" ipfs proc go IpfsProcessRcvHead enter")
	//recvSync := time.NewTicker(5 * time.Second)
	//defer recvSync.Stop()
	flg := -1
	for {
		select {
		//case headers := <-HeaderIpfsCh:
		case reqs := <-d.dpIpfs.HeaderIpfsCh: //
			log.Debug(" ipfs proc recv block headers", "len", len(reqs))
			//gIpfsCache.bassign = false
			d.dpIpfs.DownMutex.Lock()
			for _, req := range reqs {
				if req.Flag == 1 {
					flg = d.SyncBlockFromIpfs(req.HeadReqipfs.Hash().String(), req.HeadReqipfs.Number.Uint64(), req.coinstr, 0)
				} else if req.Flag == 2 {
					flg = d.DownloadBatchBlock(req.HeadReqipfs.Hash().String(), req.HeadReqipfs.Number.Uint64(), req.realBeginNum, req.ReqPendflg, 0)
				} else if req.ReqPendflg == 4 { //快照下载     Flag:   number,	coinstr: hash, //借存状态hash
					flg = d.DownloadBatchBlock(req.coinstr, req.Flag, req.realBeginNum, req.ReqPendflg, 0)
				}
				if flg == 1 && (req.ReqPendflg != 4) {
					failReTrans := &DownloadRetry{
						header:       req.HeadReqipfs,
						realBeginNum: req.realBeginNum,
						flag:         req.Flag,
						ReqPendflg:   req.ReqPendflg,
						coinstr:      req.coinstr,
						downNum:      1,
					} // 放入同步过程，要重新发送的
					log.Debug(" ipfs get block from  add retrans", "number", req.HeadReqipfs.Number.Uint64())
					d.dpIpfs.DownRetrans.PushBack(failReTrans)
					//d.queue.BlockRegetByOld(req.ReqPendflg, req.HeadReqipfs)
				}
				if flg == 0 && (req.ReqPendflg != 4) {
					//d.queue.BlockIpfsdeletePool(req.HeadReqipfs.Number.Uint64())
					d.SynOrignDownload(nil, 33, req.realBeginNum) //req.HeadReqipfs.Number.Uint64()) //BlockIpfsdeletePool
				}
			}
			d.dpIpfs.DownMutex.Unlock()
			//case <-recvSync.C: //
			//	go d.SynDealblock()
			//case cancel
		}
	}
	log.Debug(" ipfs proc go IpfsProcessRcvHead out")
}
func (d *Downloader) IpfsSyncSaveBatchSecondCache(newFlg bool, blockNum uint64, headHash string /*common.Hash*/, snapshootNum uint64, allhash string, file *os.File) error {
	var cache2st = new(Caches2CfgMap)
	//var tgh bool = false

	if newFlg { //
		cache2st.CurCacheBeignNum = blockNum
		cache2st.CurCacheBlockNum = blockNum
		cache2st.NumHashStore = 0
		cache2st.MapList.Numberstore = make(map[uint64]NumberMapingCoupledHash)

	} else {

		err := loadCache(cache2st, 0, file)
		if err != nil {
			log.Error("ipfs IpfsSyncSaveSecondCache loadCache error", "error", err)
			file.Close()
			return err
		}

		file.Close()
		//tgh = true
		file, _ = os.OpenFile(file.Name(), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644) //strTmpCache2File, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644) //O_TRUNC

	}

	//insert
	err, newNumber := insertNewValue(blockNum, headHash, allhash, true, &cache2st.MapList)

	/*if cache2st.NumHashStore == 0 {
		cache2st.CurCacheBeignNum = blockNum
	}*/
	if newNumber {
		cache2st.NumHashStore++
		cache2st.CurCacheBlockNum = blockNum
	}
	if snapshootNum != 0 {
		//cache2st.SnapStatusNum = snapshootNum
	}
	if err == nil {
		//err2 := storeCache(cache2st, file)
		return storeCache(cache2st, file)
		//return file.Close()
	} else {
		//file.Close()
		return nil
	}

}

//IpsfAddNewBlockToCache
func (d *Downloader) IpsfAddNewBatchBlockToCache(stCfg *Cache1StoreCfg, blockNum uint64, snapshootNum uint64, strheadHash string, allhash string) error {

	var calArrayPos uint32 = 0
	if blockNum < stCfg.OriginBlockNum {
		calArrayPos = 0
	} else {
		calArrayPos = uint32(blockNum / BATCH_NUM / Cache2StoreBatchBlockMaxNum)
	}

	log.Warn("ipfs IpsfAddNewBatchBlockToCache&2 batch", "calArrayPos", calArrayPos, "Cache2FileNum ", stCfg.Cache2FileNum, "blockNum", blockNum)

	if calArrayPos >= Cache1StoreBatchCache2Num {
		log.Error("ipfs error IpsfAddNewBatchBlockToCache calc error,calArrayPos exeeed capacity", "calArrayPos", calArrayPos)
		return fmt.Errorf("calArrayPos > IpsfAddNewBatchBlockToCache")
	}

	//var tmpBlockFile *os.File

	tmpBlockFile, newFileFlg, err := IpfsGetFileCache2ByHash(stCfg.StBatchCahce2Hash[calArrayPos], strTmpBatchCache2File)

	if err != nil {
		tmpBlockFile.Close()
		log.Error("ipfs IpsfAddNewBatchBlockToCache use IpfsGetFileCache2ByHash error", "error", err)
		return err
	}

	err = d.IpfsSyncSaveBatchSecondCache(newFileFlg, blockNum, strheadHash /*string(headHash[:])*/, snapshootNum, allhash, tmpBlockFile)
	if err != nil {
		tmpBlockFile.Close()
		log.Error("ipfs IpsfAddNewBatchBlockToCache use IpfsSyncSaveSecondCache error", "error", err)
		return err
	}
	tmpBlockFile.Close()

	if logMap {
		log.Trace("`^^^^^^IpsfAddNewBatchBlockToCache cache begin^^^^^^^```")
		ffile, _ := os.OpenFile(strTmpBatchCache2File, os.O_RDWR, 0644)
		cache2st := new(Caches2CfgMap)
		cache2st.MapList.Numberstore = make(map[uint64]NumberMapingCoupledHash)
		loadCache(cache2st, 0, ffile)
		for key, value := range cache2st.MapList.Numberstore {

			log.Trace("SaveSecondCache info key", "key", key)
			for key2, value2 := range value.Blockhash {
				log.Trace("key-value", "key2", key2, "value:", value2, "k-v", cache2st.MapList.Numberstore[key].Blockhash[key2])
			}

		}
		ffile.Close()
		log.Trace("`^^^^^^cache end^^^^^^^```\n")

	}

	// add file
	newHash, _, err := IpfsAddNewFile(strTmpBatchCache2File, false)
	if err != nil {
		log.Error("ipfs IpsfAddNewBatchBlockToCache error IpfsAddNewFile", "err", err)
		return err
	}

	stCfg.StBatchCahce2Hash[calArrayPos] = string(newHash[0:IpfsHashLen])
	err = WriteJsFile(path.Join(strCacheDirectory, strCache1BlockFile), stCfg)
	if err != nil {
		log.Error("ipfs IpsfAddNewBatchBlockToCache WriteJsFile error", "error", err)
		return err
	}
	return nil
	//return d.IPfsDirectoryUpdate()
	//return nil

}

//批量存储区块
func (d *Downloader) AddNewBatchBlockToIpfs() {
	/*d.dpIpfs.BatchStBlock.headerStoreFile
	d.dpIpfs.BatchStBlock.bodyStoreFile
	d.dpIpfs.BatchStBlock.receiptStoreFile*/
	if d.dpIpfs.BatchStBlock.curBlockNum%BATCH_NUM != 0 {
		log.Error(" ipfs AddNewBatchBlockToIpfs error ", "blockNum", d.dpIpfs.BatchStBlock.curBlockNum)

	}
	d.dpIpfs.BatchStBlock.headerStoreFile.Close()
	d.dpIpfs.BatchStBlock.bodyStoreFile.Close()
	d.dpIpfs.BatchStBlock.receiptStoreFile.Close()
	//压缩区块 //批量
	//
	fhandler, _ := os.Stat(strBatchBodyFile)
	batchBodySize := fhandler.Size()
	gIpfsStat.totalBatchBlockSize += batchBodySize

	bHeadHash, _, err1 := IpfsAddNewFile(strBatchHeaderFile, true)
	if err1 != nil {
		log.Error(" ipfs AddNewBatchBlockToIpfs error", "strBatchHeaderFile", bHeadHash)
		return
	}
	bBodyHash, batchZipSize, err1 := IpfsAddNewFile(strBatchBodyFile, true)
	if err1 != nil {
		log.Error(" ipfs AddNewBatchBlockToIpfs error", "strBatchBodyFile", bBodyHash)
		return
	}

	gIpfsStat.totalZipBatchBlockSize += batchZipSize
	log.Warn("static ipfs AddNewBatchBlockToIpfs body info", "blockNum", d.dpIpfs.BatchStBlock.curBlockNum, "batchsize", batchBodySize, "zipsize", batchZipSize, "batchtoatalsize", gIpfsStat.totalBatchBlockSize, "zipTotalsize", gIpfsStat.totalZipBatchBlockSize)
	bReceiptHash, _, err1 := IpfsAddNewFile(strBatchReceiptFile, true)
	if err1 != nil {
		log.Error(" ipfs AddNewBatchBlockToIpfs error", "strBatchReceiptFile", bReceiptHash)
		return
	}
	strAllHash := "1:" + string(bHeadHash[0:IpfsHashLen]) + ",2:" + string(bBodyHash[0:IpfsHashLen]) + ",3:" + string(bReceiptHash[0:IpfsHashLen]) + ","

	if gIpfsStoreCache.storeipfsCache1 == nil {
		readCacheCfg, err := d.IpfsSyncGetFirstCache(0) //"firstCacheInfo.jn"
		if err != nil {
			fmt.Println("cache1 is nil, create it", err)
			log.Debug("ipfs RecvBlockToDeal IpfsSyncGetFirstCache batchblock cache1 is nil, create it  ", "error=", err)
			//readCacheCfg2 := Cache1StoreCfg{}
			readCacheCfg.OriginBlockNum = 0
			readCacheCfg.CurrentBlockNum = 0
			readCacheCfg.Cache2FileNum = 0
			if d.dpIpfs.BatchStBlock.ExpectBeginNum != 1 {
				//获取失败从本地文件载入
				ReadJsFile(path.Join(strCacheDirectory, strCache1BlockFile), readCacheCfg)
				log.Debug("ipfs RecvBlockToDeal IpfsSyncGetFirstCache  load localfile", "OriginBlockNum=", readCacheCfg.OriginBlockNum, "CurrentBlockNum=", readCacheCfg.CurrentBlockNum)
			}
		}
		gIpfsStoreCache.storeipfsCache1 = readCacheCfg
	}

	if d.dpIpfs.BatchStBlock.ExpectBeginNum%BATCH_NUM == 1 {
		d.IpsfAddNewBatchBlockToCache(gIpfsStoreCache.storeipfsCache1, d.dpIpfs.BatchStBlock.ExpectBeginNum, 0, d.dpIpfs.BatchStBlock.ExpectBeginNumhash, strAllHash)
	} else {
		d.IpsfAddNewBatchBlockToCache(gIpfsStoreCache.storeipfsCache1, d.dpIpfs.BatchStBlock.minBlockNum, 0, d.dpIpfs.BatchStBlock.minBlockNumHash, strAllHash)
	}
	log.Debug("ipfs AddNewBatchBlockToIpfs sucess ", "blockNum", d.dpIpfs.BatchStBlock.ExpectBeginNum)

	gIpfsProcessBlockNumber = d.dpIpfs.BatchStBlock.curBlockNum
	//file test
	if d.dpIpfs.BatchStBlock.ExpectBeginNum == 1 {
		//	compareFiletoBatchBlock()
	}
	d.BatchBlockStoreInit(true)
}
func (d *Downloader) AddStateRootInfoToIpfs(blockNum uint64, strheadHash string, filePath string) {

	newBlock := (blockNum/BATCH_NUM)*BATCH_NUM + 1 //对其取300 整

	/*if newBlock == d.dpIpfs.BatchStBlock.ExpectBeginNum {

	} else {
		gIpfsStoreCache.storeipfsCache1
	}*/
	//压缩区块 //压缩快照
	//var snapSize int64

	fhandler, _ := os.Stat(filePath)
	snapSize := fhandler.Size()
	gIpfsStat.totalSnapDataSize += snapSize
	bHash, zipSize, err1 := IpfsAddNewFile(filePath, true)
	if err1 != nil {
		log.Error(" ipfs AddStateRootInfoToIpfs error IpfsAddNewFile", "filePath", filePath)
		return
	}
	gIpfsStat.totalZipSnapDataSize += zipSize

	log.Warn("static ipfs AddStateRootInfoToIpfs snap", "blockNum", blockNum, "newblock", newBlock, "bHash", string(bHash[0:IpfsHashLen]),
		"snapSize", snapSize, "zipSize", zipSize, "snaptotalSize", gIpfsStat.totalSnapDataSize, "snapTotalzipSize", gIpfsStat.totalZipSnapDataSize)
	if gIpfsStoreCache.storeipfsCache1 == nil {
		readCacheCfg, err := d.IpfsSyncGetFirstCache(0) //"firstCacheInfo.jn"
		if err != nil {
			fmt.Println("cache1 is nil, create it", err)
			log.Debug("ipfs RecvBlockToDeal IpfsSyncGetFirstCache  snap cache1 is nil, create it  ", "error=", err)
			//readCacheCfg2 := Cache1StoreCfg{}
			readCacheCfg.OriginBlockNum = 0
			readCacheCfg.CurrentBlockNum = 0
			readCacheCfg.Cache2FileNum = 0
			//获取失败从本地文件载入
			ReadJsFile(path.Join(strCacheDirectory, strCache1BlockFile), readCacheCfg)
			log.Debug("ipfs RecvBlockToDeal IpfsSyncGetFirstCache  load localfile", "OriginBlockNum=", readCacheCfg.OriginBlockNum, "CurrentBlockNum=", readCacheCfg.CurrentBlockNum)

		}
		gIpfsStoreCache.storeipfsCache1 = readCacheCfg
	}
	strHash := "4:" + string(bHash[0:IpfsHashLen]) + ","
	d.IpsfAddNewBatchBlockToCache(gIpfsStoreCache.storeipfsCache1, newBlock, blockNum, strheadHash, strHash)
	err := d.IPfsDirectoryUpdate()
	if err != nil {
		log.Error(" ipfs AddStateRootInfoToIpfs error IPfsDirectoryUpdate", "blockNum", blockNum, "filePath", strheadHash)
		return
	}
	log.Warn(" ipfs snapshoot AddStateRootInfoToIpfs sucess", "blockNum", blockNum, "strheadHash", strheadHash, "filePath", filePath)
}

func (d *Downloader) TestSnapshoot() {
	testSnap := time.NewTicker(1 * time.Minute)
	defer testSnap.Stop()
	var beginNum uint64 = 51
	//var testHash common.Hash //:= "qazwsxedcrfv"
	fmt.Println("ipfs TestSnapshoot timer", testSnap)
	path := "D:\\panic.log"

	for {
		select {
		case <-testSnap.C:
			beginNum += 300
			strTmp := strconv.Itoa(int(beginNum))
			fmt.Println("TestSnapshoot beginNum", beginNum, strTmp)
			d.AddStateRootInfoToIpfs(beginNum, strTmp, path)
			//case cancel
		}
	}

}

var headerBatchBuf [][]byte = make([][]byte, 302, 302)
var bodyBatchBuf [][]byte = make([][]byte, 302, 302)
var receiptBatchBuf [][]byte = make([][]byte, 302, 302)

func (d *Downloader) BatchStoreAllBlock(stBlock *types.BlockAllSt) bool {

	/*rlp.Encode(d.dpIpfs.BatchStBlock.headerStoreFile, stBlock.Sblock.Header())
	rlp.Encode(d.dpIpfs.BatchStBlock.bodyStoreFile, stBlock.Sblock.Body())
	rlp.Encode(d.dpIpfs.BatchStBlock.receiptStoreFile, stBlock.SReceipt)*/
	var offset uint64 = 0

	blockNum := stBlock.Sblock.NumberU64()
	d.dpIpfs.BatchStBlock.curBlockNum = blockNum
	blockNumint := int(blockNum) //file test
	/*if blockNum == 1 {
		log.Debug(" ipfs BatchStoreAllBlock write ExpectBeginNum ", "blockNum", blockNum)
		d.dpIpfs.BatchStBlock.ExpectBeginNum = blockNum
		d.dpIpfs.BatchStBlock.ExpectBeginNumhash = stBlock.Sblock.Hash()
		d.dpIpfs.BatchStBlock.minBlockNum = blockNum
		d.dpIpfs.BatchStBlock.minBlockNumHash = stBlock.Sblock.Header().Hash()
	} else if blockNum%BATCH_NUM == 0 {
		log.Debug(" ipfs BatchStoreAllBlock write ExpectBeginNum ", "blockNum", blockNum)
		d.dpIpfs.BatchStBlock.ExpectBeginNum = blockNum
		d.dpIpfs.BatchStBlock.ExpectBeginNumhash = stBlock.Sblock.Hash()
	}*/
	//log.Trace(" ipfs BatchStoreAllBlock recv  block", "ExpectBeginNum", d.dpIpfs.BatchStBlock.ExpectBeginNum, "blockNum", blockNum)
	if d.dpIpfs.BatchStBlock.ExpectBeginNum != 0 {
		beginNum := d.dpIpfs.BatchStBlock.ExpectBeginNum
		//读文件时，可能原来文件中携带的值较大，而实际测试删除数据又从1开始
		//if blockNum-d.dpIpfs.BatchStBlock.ExpectBeginNum > BATCH_NUM || d.dpIpfs.BatchStBlock.ExpectBeginNum-blockNum > BATCH_NUM {

		//distance := math.Abs(float64(blockNum - d.dpIpfs.BatchStBlock.ExpectBeginNum))
		//if distance > 300 {
		if ((blockNum > beginNum) && (blockNum-beginNum >= BATCH_NUM)) || ((beginNum > blockNum) && (beginNum-blockNum >= BATCH_NUM)) {
			log.Warn(" ipfs BatchStoreAllBlock write file error illage", "blockNum", blockNum, "ExpectBeginNum", beginNum)
			//后加
			if blockNum == 1 {
				log.Warn(" ipfs BatchStoreAllBlock recv new block num=1 ,then clear file")
				d.BatchBlockStoreInit(true)
			}
			//return false
			//如文件保存到325， 重启后变为 621开始
			/*log.Warn(" ipfs BatchStoreAllBlock write file ExpectBeginNum illage ,then clear ", "blockNum", blockNum, "ExpectBeginNum", d.dpIpfs.BatchStBlock.ExpectBeginNum)
			if d.dpIpfs.BatchStBlock.headerStoreFile != nil {

				d.dpIpfs.BatchStBlock.headerStoreFile.Close()
				d.dpIpfs.BatchStBlock.bodyStoreFile.Close()
				d.dpIpfs.BatchStBlock.receiptStoreFile.Close()
			}
			d.BatchBlockStoreInit(true)*/

		}
	} else {
		if blockNum == 1 {
			log.Warn(" ipfs BatchStoreAllBlock recv new block num=1 ,then clear file")
			d.BatchBlockStoreInit(true)
			gIpfsStoreCache.storeipfsCache1 = new(Cache1StoreCfg)
		}
	}
	if gIpfsProcessBlockNumber == 0 {
		gIpfsProcessBlockNumber = blockNum
	}
	if blockNum%BATCH_NUM == 1 {
		log.Debug(" ipfs BatchStoreAllBlock write ExpectBeginNum ", "blockNum", blockNum)
		d.dpIpfs.BatchStBlock.ExpectBeginNum = blockNum
		d.dpIpfs.BatchStBlock.ExpectBeginNumhash = stBlock.Sblock.Hash().String()
	}
	if d.dpIpfs.BatchStBlock.minBlockNum == 0 {
		log.Debug(" ipfs BatchStoreAllBlock write minBlockNum ", "blockNum", blockNum)
		d.dpIpfs.BatchStBlock.minBlockNum = blockNum
		d.dpIpfs.BatchStBlock.minBlockNumHash = stBlock.Sblock.Header().Hash().String()
	}

	bhead, err := rlp.EncodeToBytes(stBlock.Sblock.Header())
	if err != nil {
		log.Error(" ipfs BatchStoreAllBlock error rlp.EncodeToBytes", "blockNum", blockNum)
		return false
	}
	//file test
	if blockNum <= BATCH_NUM {
		headerBatchBuf[blockNumint] = make([]byte, len(bhead))
		copy(headerBatchBuf[blockNum], bhead)
	}

	offset = uint64(len(bhead))
	//log.Trace(" ipfs BatchStoreAllBlock write header ", "blockNum", blockNum, "offset", offset)
	binary.Write(d.dpIpfs.BatchStBlock.headerStoreFile, binary.BigEndian, HeadBatchFlag)
	binary.Write(d.dpIpfs.BatchStBlock.headerStoreFile, binary.BigEndian, offset)
	binary.Write(d.dpIpfs.BatchStBlock.headerStoreFile, binary.BigEndian, blockNum)
	d.dpIpfs.BatchStBlock.headerStoreFile.Write(bhead)
	recpts := make([]types.CoinReceipts, 0)
	//txsCount := 0
	for _, curr := range stBlock.Sblock.Currencies() {
		recpts = append(recpts, types.CoinReceipts{CoinType: curr.CurrencyName, Receiptlist: curr.Receipts.GetReceipts()})
		//txsCount += len(curr.Transactions.GetTransactions())
	}

	/*if txsCount == 0 {
		return true
	}

	bbody, err := rlp.EncodeToBytes(stBlock.Sblock.Body())
	*/
	// body 改为存 整个sblock
	bbody, err := rlp.EncodeToBytes(stBlock.Sblock)
	if err != nil {
		log.Error(" ipfs BatchStoreAllBlock error bbody  rlp.EncodeToBytes", "blockNum", blockNum)
	}
	//file test
	if blockNum <= BATCH_NUM {
		bodyBatchBuf[blockNumint] = make([]byte, len(bbody))
		copy(bodyBatchBuf[blockNum], bbody)
	}

	offset = uint64(len(bbody))
	log.Trace(" ipfs BatchStoreAllBlock write body ", "blockNum", blockNum, "offset", offset)
	binary.Write(d.dpIpfs.BatchStBlock.bodyStoreFile, binary.BigEndian, BodyBatchFlag)
	binary.Write(d.dpIpfs.BatchStBlock.bodyStoreFile, binary.BigEndian, offset)
	binary.Write(d.dpIpfs.BatchStBlock.bodyStoreFile, binary.BigEndian, blockNum)
	d.dpIpfs.BatchStBlock.bodyStoreFile.Write(bbody)

	breceipt, err := rlp.EncodeToBytes(recpts)
	if err != nil {
		log.Error(" ipfs BatchStoreAllBlock error breceipt rlp.EncodeToBytes", "blockNum", blockNum)
	}
	//file test
	if blockNum <= BATCH_NUM {
		receiptBatchBuf[blockNumint] = make([]byte, len(breceipt))
		copy(receiptBatchBuf[blockNum], breceipt)
	}

	offset = uint64(len(breceipt))
	//log.Trace(" ipfs BatchStoreAllBlock write receipt ", "blockNum", blockNum, "offset", offset)

	binary.Write(d.dpIpfs.BatchStBlock.receiptStoreFile, binary.BigEndian, ReceiptBatchFlag)
	binary.Write(d.dpIpfs.BatchStBlock.receiptStoreFile, binary.BigEndian, offset)
	binary.Write(d.dpIpfs.BatchStBlock.receiptStoreFile, binary.BigEndian, blockNum)
	d.dpIpfs.BatchStBlock.receiptStoreFile.Write(breceipt)

	return true
}
func compareFiletoBatchBlock() {
	var err error = nil

	var blockNum, offset uint64
	var errb error = nil
	headerStoreFile, err := os.OpenFile(strBatchHeaderFile, os.O_RDWR, 0644)
	if err != nil {
		log.Error(" ipfs  error headerStoreFile")
	}
	bodyStoreFile, err := os.OpenFile(strBatchBodyFile, os.O_RDWR, 0644)
	if err != nil {
		log.Error(" ipfs  error bodyStoreFile")
	}
	receiptStoreFile, err := os.OpenFile(strBatchReceiptFile, os.O_RDWR, 0644)
	if err != nil {
		log.Error(" ipfs  error receiptStoreFile")
	}
	defer func() {
		headerStoreFile.Close()
		bodyStoreFile.Close()
		receiptStoreFile.Close()
	}()

	log.Info(" compareFiletoBatchBlock begin header")
	var offsetflag uint64
	for {
		errb = binary.Read(headerStoreFile, binary.BigEndian, &offsetflag)
		if offsetflag != HeadBatchFlag {
			log.Info(" compareFiletoBatchBlock  header offset flag error ", "offsetflag", offsetflag)
			break
		}

		errb = binary.Read(headerStoreFile, binary.BigEndian, &offset)
		if errb == io.EOF || offset > 1024000000 {
			log.Debug(" file over")
			break
		}
		errb = binary.Read(headerStoreFile, binary.BigEndian, &blockNum)
		if errb == io.EOF {
			log.Debug(" file over")
			break
		}
		if blockNum > BATCH_NUM {
			return
		}

		log.Info(" compareFiletoBatchBlock begin header", "offset", offset, "blockNum", blockNum)
		//offset = 64
		blockBuf := make([]byte, int(offset))
		leng, errb := headerStoreFile.Read(blockBuf)
		log.Debug(" file flag", "err", errb, "leng", leng)
		fmt.Println(errb, len(blockBuf))
		/*leng, errb = headerStoreFile.ReadAt(blockBuf, 16)
		log.Debug(" file flag", "err", errb, "leng", leng)
		fmt.Println(errb, len(blockBuf))*/
		if errb == io.EOF || leng != len(blockBuf) {
			log.Debug(" file over")
			break
		}
		obj := new(types.Header)
		//errd := rlp.Decode(blockFile, obj)
		errd := rlp.DecodeBytes(blockBuf, obj)
		if errd != nil {
			log.Error("ipfs dencode block info from ParseBatchHeader", "err", errd)
		}
		//d.SynOrignDownload(obj,1,blockNum)
		//headerStoreFile.ReadAt
		if bytes.Equal(blockBuf, headerBatchBuf[int(blockNum)]) {
			log.Warn(" compareFiletoBatchBlock head is equal ", "len", len(blockBuf), "blockNum", blockNum, "objblockNum", obj.Number.Uint64())
		} else {
			log.Warn(" compareFiletoBatchBlock head is not not equal ", "len", len(blockBuf), "blockNum", blockNum, "objblockNum", obj.Number.Uint64())
		}
		if blockNum == BATCH_NUM {
			break
		}

	}
	log.Info(" compareFiletoBatchBlock begin body")
	for {

		errb = binary.Read(bodyStoreFile, binary.BigEndian, &offsetflag)
		if offsetflag != BodyBatchFlag {
			log.Info(" compareFiletoBatchBlock  body offset flag error ", "offsetflag", offsetflag)
			break
		}

		errb = binary.Read(bodyStoreFile, binary.BigEndian, &offset)
		if errb == io.EOF {
			log.Debug(" file over")
			break
		}
		errb = binary.Read(bodyStoreFile, binary.BigEndian, &blockNum)
		if errb == io.EOF {
			log.Debug(" file over")
			break
		}
		blockBuf := make([]byte, offset)
		leng, errb := bodyStoreFile.Read(blockBuf)
		if errb == io.EOF || leng != len(blockBuf) {
			log.Debug(" file over")
			break
		}
		obj := new(types.Body)
		//errd := rlp.Decode(blockFile, obj)
		errd := rlp.DecodeBytes(blockBuf, obj)
		if errd != nil {
			log.Error("ipfs dencode block info from ParseBatchHeader", "err", errd)
		}

		if bytes.Equal(blockBuf, bodyBatchBuf[int(blockNum)]) {
			log.Warn(" compareFiletoBatchBlock body is equal ", "len", len(blockBuf), "blockNum", blockNum)
		} else {
			log.Warn(" compareFiletoBatchBlock body is not not equal ", "len", len(blockBuf), "blockNum", blockNum)
		}
		if blockNum == BATCH_NUM {
			break
		}
	}
	log.Info(" compareFiletoBatchBlock begin receipt")
	for {
		errb = binary.Read(receiptStoreFile, binary.BigEndian, &offsetflag)
		if offsetflag != ReceiptBatchFlag {
			log.Info(" compareFiletoBatchBlock  body offset flag error ", "offsetflag", offsetflag)
			break
		}

		errb = binary.Read(receiptStoreFile, binary.BigEndian, &offset)
		if errb == io.EOF {
			log.Debug(" file over")
			break
		}
		errb = binary.Read(receiptStoreFile, binary.BigEndian, &blockNum)
		if errb == io.EOF {
			log.Debug(" file over")
			break
		}
		blockBuf := make([]byte, offset)
		leng, errb := receiptStoreFile.Read(blockBuf)
		if errb == io.EOF || leng != len(blockBuf) {
			log.Debug(" file over")
			break
		}
		obj := new(types.Receipts)
		//errd := rlp.Decode(blockFile, obj)
		errd := rlp.DecodeBytes(blockBuf, obj)
		if errd != nil {
			log.Error("ipfs dencode block info from ParseBatchHeader", "err", errd)
		}
		if bytes.Equal(blockBuf, receiptBatchBuf[int(blockNum)]) {
			log.Warn(" compareFiletoBatchBlock receipt is equal ", "len", len(blockBuf), "blockNum", blockNum)
		} else {
			log.Warn(" compareFiletoBatchBlock receipt is not not equal ", "len", len(blockBuf), "blockNum", blockNum)
		}
		if blockNum == BATCH_NUM {
			break
		}
		//d.SynOrignDownload(obj,1,blockNum)
	}

}
func (d *Downloader) BatchBlockStoreInit(bNeedClear bool) {
	var err error = nil
	var fileFlg int
	d.dpIpfs.BatchStBlock.curBlockNum = 0
	d.dpIpfs.BatchStBlock.ExpectBeginNum = 0
	d.dpIpfs.BatchStBlock.minBlockNum = 0
	if bNeedClear == true {
		fileFlg = os.O_CREATE | os.O_TRUNC | os.O_RDWR
	} else {
		fileFlg = os.O_CREATE | os.O_RDWR
	}
	d.dpIpfs.BatchStBlock.headerStoreFile, err = os.OpenFile(strBatchHeaderFile, fileFlg, 0644)
	if err != nil {
		log.Error(" ipfs BatchBlockStoreInit error headerStoreFile")
	}
	d.dpIpfs.BatchStBlock.bodyStoreFile, err = os.OpenFile(strBatchBodyFile, fileFlg, 0644)
	if err != nil {
		log.Error(" ipfs BatchBlockStoreInit error bodyStoreFile")
	}
	d.dpIpfs.BatchStBlock.receiptStoreFile, err = os.OpenFile(strBatchReceiptFile, fileFlg, 0644)
	if err != nil {
		log.Error(" ipfs BatchBlockStoreInit error receiptStoreFile")
	}
	if bNeedClear == true {
		log.Warn(" ipfs BatchBlockStoreInit clear file ok")
		return
	}
	if GetFileSize(strBatchHeaderFile) == 0 {
		log.Error(" ipfs BatchBlockStoreInit error file is zero")
		return
	}
	hflg, hbegin, hhash, hlast := d.checkStoreFile(d.dpIpfs.BatchStBlock.headerStoreFile, HeadBatchFlag, true)
	bflg, bbegin, _, blast := d.checkStoreFile(d.dpIpfs.BatchStBlock.bodyStoreFile, BodyBatchFlag, false)
	fflg, rbegin, _, rlast := d.checkStoreFile(d.dpIpfs.BatchStBlock.receiptStoreFile, ReceiptBatchFlag, false)
	log.Debug(" ipfs BatchBlockStoreInit", "hflg", hflg, "bflg", bflg, "fflg", fflg, "hbegin", hbegin, "bbegin", bbegin, "rbegin", rbegin, "hlast", hlast, "blast", blast, "rlast", rlast)

	if bbegin%BATCH_NUM == 1 { //读取文件中的值赋值
		d.dpIpfs.BatchStBlock.ExpectBeginNum = bbegin
		d.dpIpfs.BatchStBlock.ExpectBeginNumhash = hhash
		//d.dpIpfs.BatchStBlock.curBlockNum = blast //后面先不考虑
		log.Debug(" ipfs BatchBlockStoreInit ExpectBeginNum and ExpectBeginNumhash", "bbegin", bbegin, hhash)
	}
	if hflg && hbegin == bbegin && bbegin == rbegin {
		log.Debug(" ipfs BatchBlockStoreInit ExpectBeginNum begin", "bbegin", bbegin, "hash", hhash)
		if bbegin%BATCH_NUM != 1 {
			log.Error(" ipfs BatchBlockStoreInit  read file error")
			return
		}
		d.dpIpfs.BatchStBlock.ExpectBeginNum = bbegin
		d.dpIpfs.BatchStBlock.minBlockNum = bbegin
		d.dpIpfs.BatchStBlock.ExpectBeginNumhash = hhash
		d.dpIpfs.BatchStBlock.minBlockNumHash = hhash
	}
	if hlast == blast && blast == rlast {
		log.Debug(" ipfs BatchBlockStoreInit ExpectBeginNum", "bbegin", bbegin)
		d.dpIpfs.BatchStBlock.curBlockNum = blast
	}
}

func (d *Downloader) checkStoreFile(blockFile *os.File, BatchFlag uint64, headFlg bool) (bool, uint64, string /*common.Hash*/, uint64) {
	var errb error = nil
	var blockNum, offset, offsetflag, beginNum, lastestblock uint64
	var headHash string //common.Hash
	var okFlg bool = false
	for {
		errb = binary.Read(blockFile, binary.BigEndian, &offsetflag)
		if errb == io.EOF {
			log.Info(" checkStoreFile  head normal over")
			okFlg = true
			break
		}

		if offsetflag != BatchFlag {
			log.Info(" checkStoreFile  head offset flag error ", "offsetflag", offsetflag)
			break
		}
		errb = binary.Read(blockFile, binary.BigEndian, &offset)
		if errb == io.EOF || offset > 102400000000 {
			log.Debug(" checkStoreFile file over", "offset", offset)
			break
		}
		errb = binary.Read(blockFile, binary.BigEndian, &blockNum)
		if errb == io.EOF {
			log.Debug(" checkStoreFile file over", "blockNum", blockNum)
			break
		}

		lastestblock = blockNum

		blockBuf := make([]byte, offset)
		leng, errb := blockFile.Read(blockBuf)
		if errb == io.EOF || leng != len(blockBuf) {
			log.Debug(" checkStoreFile file over", "length", leng, "blockNum", blockNum)
			break
		}
		if beginNum == 0 {
			beginNum = blockNum
			if headFlg {
				obj := new(types.Header)
				errd := rlp.DecodeBytes(blockBuf, obj)
				if errd != nil {
					break
				}
				if blockNum != obj.Number.Uint64() {
					log.Error("checkStoreFile DecodeBytes number error", "blockNum", blockNum, "packetblocknum", obj.Number.Uint64())
					break
				} else if errd == nil {
					headHash = obj.Hash().String()
				}
			}

		}
	}

	return okFlg, beginNum, headHash, lastestblock
}
func (d *Downloader) ParseBatchHeader(batchblockhash string, beginReqNumber uint64) bool {
	//var batchblockhash string
	var blockNum, offset, offsetflag uint64
	blockFile, err := IpfsGetBlockByHash(batchblockhash, true)
	//解压区块
	defer func() {
		blockFile.Close()
		os.Remove(batchblockhash + ".unzip")
	}()
	if err != nil {
		log.Debug(" ParseBatchHeader error in IpfsGetBlockByHash", "error", err)
		return false
	}
	log.Debug("ipfs  ParseBatchHeader begin", "beginReqNumber", beginReqNumber)
	var errb error = nil
	for {
		errb = binary.Read(blockFile, binary.BigEndian, &offsetflag)
		if errb == io.EOF {
			log.Info(" ParseBatchHeader  head normal over")
			break
		}
		if offsetflag != HeadBatchFlag {
			log.Info(" ParseBatchHeader  head offset flag error ", "offsetflag", offsetflag)
			break
		}
		errb = binary.Read(blockFile, binary.BigEndian, &offset)
		if errb == io.EOF || offset > 10240000000 {
			log.Debug(" ParseBatchHeader file over", "offset", offset)
			break
		}
		errb = binary.Read(blockFile, binary.BigEndian, &blockNum)
		if errb == io.EOF {
			log.Debug(" ParseBatchHeader file over", "blockNum", blockNum)
			break
		}
		if (beginReqNumber > blockNum) || ((blockNum > beginReqNumber) && (blockNum-beginReqNumber > BATCH_NUM)) {
			log.Debug(" ParseBatchHeader file error,blockNum illegality ", "blockNum", blockNum, "beginReqNumber", beginReqNumber)
			return false
		}
		blockBuf := make([]byte, offset)
		leng, errb := blockFile.Read(blockBuf)
		if errb == io.EOF || leng != len(blockBuf) {
			log.Debug(" ParseBatchHeader file over", "length", leng, "blockNum", blockNum)
			break
		}
		obj := new(types.Header)
		//errd := rlp.Decode(blockFile, obj)
		errd := rlp.DecodeBytes(blockBuf, obj)
		if errd != nil {
			log.Error("ipfs dencode block info from ParseBatchHeader", "err", errd, "blockNum", blockNum)
		}
		if blockNum != obj.Number.Uint64() {
			log.Error("ipfs dencode block info error", "blockNum", blockNum, "packetblocknum", obj.Number.Uint64())
		}
		d.SynOrignDownload(obj, 1, obj.Number.Uint64())
	}

	return true
}
func (d *Downloader) ParseBatchBody(batchblockhash string, beginReqNumber uint64, realBeginNum uint64, flg int) bool {

	var blockNum, offset, offsetflag uint64
	blockFile, err := IpfsGetBlockByHash(batchblockhash, true)
	//解压区块
	defer func() {
		blockFile.Close()
		os.Remove(batchblockhash + ".unzip")
	}()
	if err != nil {
		log.Debug(" ParseBatchBody error in IpfsGetBlockByHash", "error", err)
		return false
	}
	log.Debug("ipfs  ParseBatchBody begin", "beginReqNumber", beginReqNumber)
	var errb error = nil
	for {
		errb = binary.Read(blockFile, binary.BigEndian, &offsetflag)
		if errb == io.EOF {
			log.Info(" ParseBatchBody  head normal over")
			break
		}
		if offsetflag != BodyBatchFlag {
			log.Info(" ParseBatchBody  body offset flag error ", "offsetflag", offsetflag)
			break
		}

		errb = binary.Read(blockFile, binary.BigEndian, &offset)
		if errb == io.EOF || offset > 102400000000 {
			log.Debug(" ParseBatchBody file over", "offset", offset)
			break
		}
		errb = binary.Read(blockFile, binary.BigEndian, &blockNum)
		if errb == io.EOF {
			log.Debug(" ParseBatchBody file over", "blockNum", blockNum)
			break
		}
		/*if (beginReqNumber > blockNum) || ((blockNum > beginReqNumber) && (blockNum-beginReqNumber > BATCH_NUM)) {
			log.Debug(" ParseBatchBody file error,blockNum illegality ", "blockNum", blockNum, "beginReqNumber", beginReqNumber)
			return false
		}*/
		blockBuf := make([]byte, offset)
		leng, errb := blockFile.Read(blockBuf)
		if errb == io.EOF || leng != len(blockBuf) {
			log.Debug(" ParseBatchBody file over", "length", leng, "blockNum", blockNum)
			break
		}
		if blockNum < realBeginNum {
			log.Debug(" ParseBatchBody file contine", "realBeginNum", realBeginNum, "blockNum", blockNum)
			continue
		}
		/*obj := new(types.Body)
		//errd := rlp.Decode(blockFile, obj)*/
		//body改为stblock
		obj := new(types.Block)
		errd := rlp.DecodeBytes(blockBuf, obj)
		if errd != nil {
			log.Error("ipfs dencode block info from ParseBatchBody", "err", errd, "blockNum", blockNum)
		}
		if (beginReqNumber > blockNum) || ((blockNum > beginReqNumber) && (blockNum-beginReqNumber >= BATCH_NUM)) {
			log.Debug(" ParseBatchBody file error,blockNum illegality ", "blockNum", blockNum, "beginReqNumber", beginReqNumber)
		} else {
			if flg == 4 {

			} else {
				d.SynOrignDownload(obj, 2, blockNum)
			}
		}
	}
	return true
}
func (d *Downloader) ParseBatchReceipt(batchblockhash string, beginReqNumber uint64, realBeginNum uint64) bool {

	var blockNum, offset, offsetflag uint64
	blockFile, err := IpfsGetBlockByHash(batchblockhash, true)
	//解压区块
	defer func() {
		blockFile.Close()
		os.Remove(batchblockhash + ".unzip")
	}()
	if err != nil {
		log.Debug(" ParseBatchReceipt error in IpfsGetBlockByHash", "error", err)
		return false
	}
	log.Debug("ipfs  ParseBatchReceipt begin", "beginReqNumber", beginReqNumber)
	var errb error = nil
	for {

		errb = binary.Read(blockFile, binary.BigEndian, &offsetflag)
		if errb == io.EOF {
			log.Info(" ParseBatchReceipt  head normal over")
			break
		}
		if offsetflag != ReceiptBatchFlag {
			log.Info(" ParseBatchReceipt  receipt offset flag error ", "offsetflag", offsetflag)
			break
		}

		errb = binary.Read(blockFile, binary.BigEndian, &offset)
		if errb == io.EOF || offset > 10240000000 {
			log.Debug(" ParseBatchReceipt file over", "offset", offset)
			break
		}
		errb = binary.Read(blockFile, binary.BigEndian, &blockNum)
		if errb == io.EOF {
			log.Debug(" ParseBatchReceipt file over", "blockNum", blockNum)
			break
		}
		/*if (beginReqNumber > blockNum) || ((blockNum > beginReqNumber) && (blockNum-beginReqNumber > BATCH_NUM)) {
			log.Debug(" ParseBatchReceipt file error,blockNum illegality ", "blockNum", blockNum, "beginReqNumber", beginReqNumber)
			return false
		}*/
		blockBuf := make([]byte, offset)
		leng, errb := blockFile.Read(blockBuf)
		if errb == io.EOF || leng != len(blockBuf) {
			log.Debug(" ParseBatchReceipt file over", "length", leng, "blockNum", blockNum)
			break
		}
		if blockNum < realBeginNum {
			log.Debug(" ParseBatchReceipt file contine", "realBeginNum", realBeginNum, "blockNum", blockNum)
			continue
		}
		obj := new(types.Receipts)
		//errd := rlp.Decode(blockFile, obj)
		errd := rlp.DecodeBytes(blockBuf, obj)
		if errd != nil {
			log.Error("ipfs dencode block info from ParseBatchReceipt", "err", errd, "blockNum", blockNum)
		}
		if (beginReqNumber > blockNum) || ((blockNum > beginReqNumber) && (blockNum-beginReqNumber >= BATCH_NUM)) {
			log.Debug(" ParseBatchReceipt file error,blockNum illegality ", "blockNum", blockNum, "beginReqNumber", beginReqNumber)
		} else {
			d.SynOrignDownload(obj, 3, blockNum)
		}
	}
	return true
}
func (d *Downloader) ParseMPTstatus(batchblockhash string, beginReqNumber uint64, realstatusNumber uint64) bool {
	blockFile, err := IpfsGetBlockByHash(batchblockhash, true)
	//解压区块
	filepath := blockFile.Name()
	defer func() {
		os.Remove(batchblockhash + ".unzip")
	}()
	blockFile.Close()

	if err != nil {
		log.Debug(" ParseMPTstatus error in IpfsGetBlockByHash", "error", err)
		return false
	}
	log.Debug("ipfs  ParseMPTstatus begin", "beginReqNumber", beginReqNumber, "realstatusNumber", realstatusNumber)
	d.blockchain.SynSnapshot(realstatusNumber, batchblockhash, filepath)
	return true
}
func (d *Downloader) dealSnaperr() {
	log.Debug(" ipfs  DownloadBatchBlock snapshoot  failed")
	//写回管道
	d.WaitSnapshoot <- 0
}
func (d *Downloader) DownloadBatchBlock(headhash string /*common.Hash*/, headNumber uint64, realBeginNum uint64, pendflag int, index int) int {
	//  后续需要 优化获取cache 流程
	realReqNumber := headNumber
	strHeadhash := headhash //string(headhash[:])
	log.Debug(" *** ipfs  DownloadBatchBlock begin in IpfsSyncGetFirstCache &&&", "headnumber", headNumber, "pendingflg", pendflag)
	var err error = nil
	gIpfsCache.getipfsCache1, err = d.IpfsSyncGetFirstCache(index)
	//readCache =(*Cache1StoreCfg)readCacheCfg//readCache1Info
	if err != nil {
		log.Error(" ipfs  DownloadBatchBlock error in IpfsSyncGetFirstCache")
		if pendflag == 4 {
			d.dealSnaperr()
		}
		return 1
	}
	//
	if pendflag == 4 { //若是状态树快照，则 转换 区块数目
		headNumber = headNumber/BATCH_NUM*BATCH_NUM + 1
		log.Debug(" *** ipfs  DownloadBatchBlock begin status mpt block", "newBlockNum", headNumber)
	}

	calArrayPos := uint32(headNumber / BATCH_NUM / Cache2StoreBatchBlockMaxNum)
	//arrayIndex := uint32((headNumber - gIpfsCache.getipfsCache1.OriginBlockNum) / Cache2StoreHashMaxNum) //
	if calArrayPos >= Cache1StoreBatchCache2Num {
		log.Error(" ipfs error DownloadBatchBlock exceed capacity", "arrayIndex", calArrayPos, "Cache1StoreCache2Num", Cache1StoreBatchCache2Num)
		if pendflag == 4 {
			d.dealSnaperr()
		}
		return 2
	}
	stCache2Infohash := gIpfsCache.getipfsCache1.StBatchCahce2Hash[calArrayPos]
	//
	cache2File, err := IpfsGetBlockByHash(stCache2Infohash, false)
	if err != nil {
		log.Error(" ipfs  DownloadBatchBlock error IpfsGetBlockByHash cache2File", "error", err)
		if pendflag == 4 {
			d.dealSnaperr()
		}
		return 1
	}
	defer func() {
		cache2File.Close()
		os.Remove(stCache2Infohash)
	}()

	//cache2st := new(Caches2CfgMap) //eof
	gIpfsCache.getipfsBatchCache2 = new(Caches2CfgMap)
	//for {
	err = loadCache(gIpfsCache.getipfsBatchCache2, 0, cache2File)
	if err != nil {
		log.Error("ipfs DownloadBatchBlock loadCache error", "error", err)
		if pendflag == 4 {
			d.dealSnaperr()
		}
		return 1
	}

	//gIpfsCache.bassign = true
	//
	if logMap {
		log.Debug("********linshi lastest begin *********```\n")
		//for key, value := range gIpfsCache.getipfsCache2.MapList.Numberstore[headNumber] {
		for key, value := range gIpfsCache.getipfsBatchCache2.MapList.Numberstore {
			//fmt.Println(i.CurrentNum, i.HashNum)
			log.Trace("Lastest curLastestInfo key", "key", key)
			//if key == headNumber
			{
				for key2, value2 := range value.Blockhash {
					log.Debug("key-value", "key2:", key2, "value:", value2, "value2", gIpfsCache.getipfsBatchCache2.MapList.Numberstore[key].Blockhash[key2])
				}
			}

		}

		log.Debug("```linshi lastest end`````\n")
	}

	_, ok := gIpfsCache.getipfsBatchCache2.MapList.Numberstore[headNumber]
	if ok {

		blockhash, ok := gIpfsCache.getipfsBatchCache2.MapList.Numberstore[headNumber].Blockhash[strHeadhash] // 1:hash,2:hash,3:hash,4:statushash
		if ok {
			if pendflag == 4 { //获取快照,先解析处理，后续可下载区块
				fmt.Println("ipfs DownloadBatchBlock get status MPT", blockhash)
				sidx := strings.Index(blockhash, "4:")
				if sidx >= 0 {
					strStatusHash := string([]byte(blockhash)[(sidx + 2):(sidx + 2 + IpfsHashLen)])
					log.Debug(" ipfs  DownloadBatchBlock get status MPT info form ipfs use by getCache2", "blockNum", headNumber, "strStatusHash", strStatusHash)
					if d.ParseMPTstatus(strStatusHash, headNumber, realReqNumber) {
						log.Debug(" ipfs  DownloadBatchBlock get status MPT over sucess")
						d.WaitSnapshoot <- 1
						return 0
					}
				}
				log.Debug(" ipfs  DownloadBatchBlock get status MPT over failed")
				//写回管道
				d.WaitSnapshoot <- 0
				return 0 //
			}

			strSet := strings.Split(blockhash, ",")
			for _, str := range strSet {
				hstr := string([]byte(str)[:2])
				/*if hstr == "4:" && pendflag == 4 {
					log.Debug(" ipfs  DownloadBatchBlock get status MPT info form ipfs use by getCache2", "blockNum", headNumber)
					if d.ParseMPTstatus(string([]byte(str)[2:]), headNumber) {
						return 1
					}
				}else */
				if hstr == "2:" {
					log.Debug(" ipfs  DownloadBatchBlock get new batchblock body form ipfs use by getCache2", "blockNum", headNumber)
					if d.ParseBatchBody(string([]byte(str)[2:]), headNumber, realBeginNum, pendflag) == false {
						return 1
					}
				} else if hstr == "3:" && pendflag == 2 { // 若是含有3,比含有2，且前面已下载
					log.Debug(" ipfs  DownloadBatchBlock get new batchblock receipt form ipfs use by getCache2", "blockNum", headNumber)
					if d.ParseBatchReceipt(string([]byte(str)[2:]), headNumber, realBeginNum) == false {
						return 1
					}
				}
			}
			//strings.Index()
			//d.queue.BlockIpfsdeletePool(headNumber)

			return 0
		} else {
			if pendflag == 4 {
				d.dealSnaperr()
			}
			log.Error("ipfs  DownloadBatchBlock  error map Blockhash", "headNumber", headNumber, "headhash", headhash)
			return 1
		}
	} else {
		log.Error("ipfs  DownloadBatchBlock  error map ", "headNumber", headNumber, "headhash", headhash)
		if pendflag == 4 {
			d.dealSnaperr()
		}
		return 1
	}

}

func (d *Downloader) SynBlockFormBlockchain() {
	log.Debug(" ipfs proc go SynBlockFormBlockchain enter")
	d.BatchBlockStoreInit(false)
	/*fmt.Println("aaaaaa")
	time.Sleep(10 * time.Second)
	fmt.Println("bbbbb down ")
	RestartIpfsDaemon()
	fmt.Println("ccccc")*/
	recvSync := time.NewTicker(5 * time.Second)
	defer recvSync.Stop()
	for {
		select {
		case <-recvSync.C:
			go d.SynDealblock()
			//case cancel
		}
	}
}

//存储过程
func (d *Downloader) SynDealblock() {

	if !atomic.CompareAndSwapInt32(&d.dpIpfs.IpfsDealBlocking, 0, 1) {
		return
	}
	defer atomic.StoreInt32(&d.dpIpfs.IpfsDealBlocking, 0)

	queueBlock := d.blockchain.GetStoreBlockInfo()
	log.Trace(" ipfs get block from blockchain", "blockNum=", queueBlock.Size())

	//队列不空 则弹出元素处理
	if !queueBlock.Empty() {
		d.dpIpfs.DownMutex.Lock()
		d.RecvBlockSaveToipfs(queueBlock)
		d.dpIpfs.DownMutex.Unlock()
	}
	/*for !queueBlock.Empty() {

		err := d.RecvBlockToDeal(queueBlock.PopItem().(*types.Block))
		if err != nil {
			log.Error(" ipfs period save block error")
		}

	}*/
}
func (d *Downloader) StatusSnapshootDeal() {

	log.Debug(" StatusSnapshootDeal enter")
	if d.dpIpfs == nil {
		log.Error(" StatusSnapshootDeal enter error dpIpfs is nil ")
		return
	}
	for {
		select {
		case tmpsnap := <-d.dpIpfs.SnapshootInfoCh:
			d.dpIpfs.DownMutex.Lock()
			d.AddStateRootInfoToIpfs(tmpsnap.BlockNumber, tmpsnap.Hashstr, tmpsnap.FilePath)
			d.dpIpfs.DownMutex.Unlock()
		}
	}

}
func (d *Downloader) SaveSnapshootStatus(blockNum uint64, strheadHash string, filePath string) bool {
	newSnap := new(SnapshootReq)
	if d.dpIpfs == nil {
		log.Error(" SaveSnapshootStatus enter error dpIpfs is nil ")
		return false
	}
	newSnap.BlockNumber = blockNum
	newSnap.Hashstr = strheadHash
	newSnap.FilePath = filePath
	d.dpIpfs.SnapshootInfoCh <- (*newSnap)
	return true
}

func (d *Downloader) SynIPFSCheck() {
	//DownRetrans:    make([]DownloadRetry, 6),
	//DownMutex:      new(sync.Mutex),
	log.Debug(" ipfs proc go SynIPFSCheck enter")
	CheckSync := time.NewTicker(10 * time.Second)
	defer CheckSync.Stop()
	for {
		select {
		case <-CheckSync.C:
			{
				log.Trace(" ipfs proc go CheckSync")
				d.dpIpfs.DownMutex.Lock()
				if lsize := d.dpIpfs.DownRetrans.Len(); lsize > 0 {
					log.Debug(" ipfs get block from  retrans", "lsize", lsize)
					//for index := 0; index < lsize; index++ {
					for element := d.dpIpfs.DownRetrans.Front(); element != nil; element = element.Next() {
						//tmp := d.dpIpfs.DownRetrans.PopItem().(*DownloadRetry)
						var res int
						tmpReq := element.Value.(*DownloadRetry)
						//res := d.SyncBlockFromIpfs(tmpReq.header.Hash(), tmpReq.header.Number.Uint64(), tmpReq.downNum%2)
						if tmpReq.flag == 1 {
							res = d.SyncBlockFromIpfs(tmpReq.header.Hash().String(), tmpReq.header.Number.Uint64(), tmpReq.coinstr, 0)
						} else if tmpReq.flag == 2 {
							res = d.DownloadBatchBlock(tmpReq.header.Hash().String(), tmpReq.header.Number.Uint64(), tmpReq.realBeginNum, tmpReq.ReqPendflg, 0)
						} else if tmpReq.ReqPendflg == 4 {
							res = d.DownloadBatchBlock(tmpReq.coinstr, tmpReq.flag, tmpReq.realBeginNum, tmpReq.ReqPendflg, 0)
						}

						log.Debug(" ipfs get block from  retrans SyncBlockFromIpfs ", "res", res, "downNum", tmpReq.downNum, "blockNum", tmpReq.header.Number.Uint64())
						if res == 1 {
							if tmpReq.downNum < 1 {
								//d.dpIpfs.DownRetrans.PushBack(tmp)
								//log.Debug(" ipfs get block from  retrans again", "num", tmp.downNum, "blockNum", tmp.header.Number.Uint64())
							} else {
								d.dpIpfs.DownRetrans.Remove(element)
								//加入原始下载方式
								if tmpReq.flag != 4 {
									d.queue.BlockRegetByOldMode(int(tmpReq.flag), tmpReq.ReqPendflg, tmpReq.header, tmpReq.realBeginNum)
								}
								//lb d.queue.BlockIpfsdeletePool(tmpReq.header.Number.Uint64())
							}
							tmpReq.downNum++
						} else if res == 0 {
							//lb d.queue.BlockIpfsdeletePool(tmpReq.header.Number.Uint64())
							d.SynOrignDownload(nil, 33, tmpReq.realBeginNum) // tmpReq.header.Number.Uint64())
							d.dpIpfs.DownRetrans.Remove(element)
						}
					}
				}
				d.dpIpfs.DownMutex.Unlock()
			}
		}
	}

}
func (d *Downloader) ClearIpfsQueue() {
	//d.dpIpfs.HeaderIpfsCh = d.dpIpfs.HeaderIpfsCh[0:0]
	if d.dpIpfs != nil && d.dpIpfs.DownRetrans != nil {
		if lsize := d.dpIpfs.DownRetrans.Len(); lsize > 0 {
			d.dpIpfs.DownMutex.Lock()
			for element := d.dpIpfs.DownRetrans.Front(); element != nil; element = element.Next() {
				d.dpIpfs.DownRetrans.Remove(element)
			}
			d.dpIpfs.DownMutex.Unlock()
		}
	}
}
func (d *Downloader) GetfirstcacheByIPFS() {
	fmt.Println("ipfs broadcast id ", d.dpIpfs.StrIpfspeerID)
	//var out bytes.Buffer
	var outerr bytes.Buffer
	//ipnsPath := "/ipns/" + d.dpIpfs.StrIpfspeerID + "/" + strCache1BlockFile //"firstCacheSync.ha"
	ipnsPath := "/ipns/" + d.dpIpfs.StrIpfspeerID + "/" + strCache1BlockFile
	c := exec.Command(d.dpIpfs.StrIPFSExecName, "cat", ipnsPath) //或cat

	//IpfsStartTimer("first")
	c.Stderr = &outerr
	outbuf, err := c.Output()
	//IpfsStopTimer()

	curCache1Info := new(Cache1StoreCfg) // Cache1StoreCfg{}
	//stdErr := outerr.String()
	if err != nil {
		fmt.Println("ipfs error IpfsSyncGetFirstCache error", err, outerr.String())
		return
	}
	err = json.Unmarshal(outbuf, curCache1Info)
	if err != nil {
		log.Error("ipfs IpfsSyncGetFirstCache json.Unmarshal error", "error", err)
		return
	}

	fmt.Println("ipfs storage block", curCache1Info.CurrentBlockNum)
	fmt.Println("ipfs cache2  block list")
	for i := 0; i < 30; i++ {
		if curCache1Info.StCahce2Hash[i] != "" {
			fmt.Println(curCache1Info.StCahce2Hash[i])
		}
	}
	fmt.Println("ipfs cache2  batch block list")
	for i := 0; i < 10; i++ {
		if curCache1Info.StBatchCahce2Hash[i] != "" {
			fmt.Println(curCache1Info.StBatchCahce2Hash[i])
		}
	}
	return
}
func (d *Downloader) GetsecondcacheByIPFS(strHash string) {
	fmt.Println("ipfs second cache  ", strHash)
	file, err := IpfsGetBlockByHash(strHash, false)
	if err != nil {
		fmt.Println("ipfs second cache  error", err)
		return
	}

	cache2st := new(Caches2CfgMap)
	loadCache(cache2st, 0, file)
	fmt.Println("ipfs second cache  block num", cache2st.CurCacheBeignNum)
	for key, value := range cache2st.MapList.Numberstore {

		for key2, value2 := range value.Blockhash {
			fmt.Println("ipfs second cache blockNum=%d  blockHash=%s, ipfshash=%s", key, key2, value2) //cache2st.MapList.Numberstore[key].Blockhash[key2],
		}
	}
	file.Close()
}
func (d *Downloader) GetBlockByIPFS(strHash string) {
	fmt.Println("ipfs block", strHash)
	file, err := IpfsGetBlockByHash(strHash, true)
	//解压
	defer func() {
		file.Close()
		os.Remove(strHash + ".unzip")
	}()
	if err != nil {
		fmt.Println("ipfs block error", err)
		return
	}
	obj := new(types.BlockAllSt) //types.Block)
	errd := rlp.Decode(file, obj)

	if errd != nil {
		fmt.Println("ipfs block rlp decode error", err)
		return
	}

	fmt.Println("ipfs block number", obj.Sblock.NumberU64())
	//fmt.Println("ipfs block Root", obj.Sblock.Root().String())
	fmt.Println("ipfs block ParentHash", obj.Sblock.ParentHash().String())
	//fmt.Println("ipfs block TxHash", obj.Sblock.TxHash().String())
	//fmt.Println("ipfs block ReceiptHash", obj.Sblock.ReceiptHash().String())

	fmt.Println("ipfs block Nonce", obj.Sblock.Nonce())
	fmt.Println("ipfs block Size", obj.Sblock.Size())
	//fmt.Println("ipfs block tx len ", len(obj.Sblock.Transactions()))
}
func (d *Downloader) GetsanpByIPFS(strHash string) {
	fmt.Println("ipfs sanpshoot", strHash)
	file, err := IpfsGetBlockByHash(strHash, true)
	//解压
	file.Close()

	if err != nil {
		fmt.Println("ipfs sanpshoot error", err)
		return
	}
	//if d.blockchain {
	d.blockchain.PrintSnapshotAccountMsg(0, "", strHash)
	//}
	os.Remove(strHash + ".unzip")
}
