// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php

package downloader

import (
	"bytes"
	"container/list"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/rlp"
	"gopkg.in/karalabe/cookiejar.v2/collections/prque"
)

const (
	IpfsHashLen           = 46
	LastestBlockStroeNum  = 100   //100
	Cache2StoreHashMaxNum = 2000  //2628000 //24 hour* (3600 second /12 second）*365
	Cache1StoreCache2Num  = 10000 //
)

//var logPrint bool = false
//var logOriginMode bool = false
//var gBlockFile *File

type Hash []byte //[IpfsHashLen]byte
type BlockStore struct {
	Numberstore map[uint64]NumberMapingCoupledHash
}
type NumberMapingCoupledHash struct {
	Blockhash map[common.Hash]string
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
	MapList          BlockStore
}

//cache1
type Cache1StoreCfg struct {
	OriginBlockNum  uint64
	CurrentBlockNum uint64
	Cache2FileNum   uint32
	StCahce2Hash    [Cache1StoreCache2Num]string
}
type DownloadRetry struct {
	header  *types.Header
	downNum int
}
type IPfsDownloader struct {
	BIpfsIsRunning      bool
	IpfsDealBlocking    int32
	StrIpfspeerID       string
	StrIpfsSecondpeerID string
	StrIPFSLocationPath string
	StrIPFSServerInfo   string
	StrIPFSExecName     string
	HeaderIpfsCh        chan []*types.Header
	BlockRcvCh          chan *types.Block
	DownMutex           *sync.Mutex
	DownRetrans         *list.List //*prque.Prque // []DownloadRetry prque.New()
}

type listBlockInfo struct {
	blockNum      uint64
	blockHeadHash common.Hash
	blockIpfshash Hash
}

//var HeaderIpfsCh chan []*types.Header

var (
	strLastestBlockFile  = "lastestblockInfo.gb" //存放最近100块缓存文件
	strCache1BlockFile   = "firstCacheInfo.jn"   //存放一级缓存文件
	strCacheDirectory    = "ipfsCachecommon"     //发布的目录
	strTmpCache2File     = "secondCacheInfo.gb"  //暂时存放查询的二级缓存文件
	strNewBlockStoreFile = "NewTmpBlcok.rp"      //新来的block 暂时保存文件
//	StrIpfspeerID        = "QmPXtaMvY6ZB67Xgeb8M2D8KuyPBXbyVEyTzaxs5TpjuNi" //peer id

)

type GetIpfsCache struct {
	bassign       bool
	lastestCache  *LastestBlcokCfg
	lastestNum    uint64
	getipfsCache1 *Cache1StoreCfg
	getipfsCache2 *Caches2CfgMap
}

type DownloadFileInfo struct {
	Downloadflg          bool
	IpfsPath             string
	StrIPFSServerInfo    string
	StrIPFSServer2Info   string
	StrIPFSServer3Info   string
	StrIPFSServer4Info   string
	PrimaryDescription   string
	SecondaryDescription string
}

var gIpfsCache GetIpfsCache
var IpfsInfo DownloadFileInfo
var logMap bool
var listPeerId [2]string
var testShowlog int

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
	err := ReadJsFile("ipfsinfo.json", &IpfsInfo)
	fmt.Println("read ipfs ", err, IpfsInfo.Downloadflg, IpfsInfo.StrIPFSServerInfo)
	if /*IpfsInfo.IpfsPath == "" ||*/ IpfsInfo.StrIPFSServerInfo == "" || IpfsInfo.PrimaryDescription == "" {
		IpfsInfo.Downloadflg = false
	}
}
func GetIpfsMode() bool {
	return IpfsInfo.Downloadflg
}
func newIpfsDownload() *IPfsDownloader {

	return &IPfsDownloader{
		BIpfsIsRunning: false,
		HeaderIpfsCh:   make(chan []*types.Header, 1),
		BlockRcvCh:     make(chan *types.Block, 1),
		DownRetrans:    list.New(), //prque.New(), //make([]DownloadRetry, 6),
		DownMutex:      new(sync.Mutex),
	}

}

var gStrIpfsName string

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
	gStrIpfsName = d.dpIpfs.StrIPFSExecName
	d.dpIpfs.StrIPFSServerInfo = "/ip4/192.168.3.30/tcp/4001/ipfs/QmQSazdGapokSejxeTTQc4tCRcHgqRPtoMeW3trRk4zA1S"
	return nil // foe test
}

//IpfsDownloadInit

func (d *Downloader) IpfsDownloadInit() error {

	var out bytes.Buffer
	var outerr bytes.Buffer
	// Directory
	CheckDirAndCreate(strCacheDirectory)
	fmt.Println("IpfsDownloadInit enter")
	//
	d.dpIpfs.StrIPFSLocationPath = IpfsInfo.IpfsPath      //"D:\\lb\\go-ipfs"
	d.dpIpfs.StrIPFSExecName = IpfsInfo.IpfsPath + "ipfs" //d.dpIpfs.StrIPFSLocationPath + "\\ipfs.exe"
	gStrIpfsName = d.dpIpfs.StrIPFSExecName               //gStrIpfsPath = d.dpIpfs.StrIPFSLocationPath
	d.dpIpfs.StrIpfspeerID = IpfsInfo.PrimaryDescription
	d.dpIpfs.StrIpfsSecondpeerID = IpfsInfo.SecondaryDescription
	d.dpIpfs.StrIPFSServerInfo = IpfsInfo.StrIPFSServerInfo //"/ip4/192.168.3.30/tcp/4001/ipfs/QmQSazdGapokSejxeTTQc4tCRcHgqRPtoMeW3trRk4zA1S"
	//	return nil // foe test
	listPeerId[0] = d.dpIpfs.StrIpfspeerID
	listPeerId[1] = d.dpIpfs.StrIpfspeerID
	if d.dpIpfs.StrIpfsSecondpeerID != "" {
		listPeerId[1] = d.dpIpfs.StrIpfsSecondpeerID
	}
	fmt.Println("peer ID ", listPeerId[0], listPeerId[1])
	log.Warn("ipfs Downloader init", "peerid0", listPeerId[0], "peerid1", listPeerId[1])

	out.Reset()
	outerr.Reset()
	c := exec.Command(d.dpIpfs.StrIPFSExecName, "init")
	c.Stdout = &out
	c.Stderr = &outerr
	err := c.Run()

	strErrInfo := outerr.String()

	if err != nil {
		log.Warn("ipfs IpfsDownloadInit init error", "error", err, "ipfs err", strErrInfo)
		//return err
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
	out.Reset()
	outerr.Reset()

	fmt.Println("ipfs daemon run")

	c = exec.Command(d.dpIpfs.StrIPFSExecName, "daemon") //"add", "D:\\melog3332.txt")
	d.dpIpfs.BIpfsIsRunning = true
	err = c.Run()
	strErrInfo = outerr.String()

	d.dpIpfs.BIpfsIsRunning = false
	if err != nil {
		log.Error("ipfs IpfsDownloadInit daemon error,exit init", "error", err, "ipfs err", outerr.String())
		//return err
	}
	//d.IpfsMode = false //启动失败时 置为false
	fmt.Println("ipfsDownloadInit error", err)

	return nil
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

// IpfsGetBlockByHash get block
func IpfsGetBlockByHash(strHash string) (*os.File, error) {
	var out bytes.Buffer
	var outerr bytes.Buffer

	c := exec.Command(gStrIpfsName, "get", strHash)
	c.Stdout = &out
	c.Stderr = &outerr
	err := c.Run()
	c.StdinPipe()
	//strErrInfo := outerr.String()

	log.Debug("ipfs IpfsGetBlockByHash info", "error", err, "strHash", strHash)

	if err != nil {
		log.Error("ipfs IpfsGetBlockByHash error", "error", err, "ipfs err", outerr.String())
		return nil, err
	}

	return os.OpenFile(strHash, os.O_RDONLY /*|os.O_APPEND*/, 0644)

}

//IpfsAddNewFile
func IpfsAddNewFile(filePath string) (Hash, error) {
	var out bytes.Buffer
	var outerr bytes.Buffer
	//1M
	c := exec.Command(gStrIpfsName, "add", "-q", "-s", "size-1048576", filePath)
	c.Stdout = &out
	c.Stderr = &outerr
	err := c.Run()
	//strErrInfo := outerr.String()

	log.Trace("ipfs IpfsAddNewFile to ipfs network", "filePath", filePath)

	if err != nil {
		log.Error("ipfs IpfsAddNewFile to  ipfs network", "error", err, "ipfs err", outerr.String())
		return nil, err
	}
	return out.Bytes(), nil
}

//IpfsGetFileCache2ByHash

func IpfsGetFileCache2ByHash(strhash string) (*os.File, bool, error) {
	var out bytes.Buffer
	var outerr bytes.Buffer
	if strhash == "" {
		//var errf error = nil
		tmpBlockFile, errf := os.OpenFile(strTmpCache2File, os.O_WRONLY|os.O_CREATE, 0644) //"secondCacheInfo.gb"
		if errf != nil {
			return nil, false, fmt.Errorf("ipfs error IpfsGetFileCache2ByHash OpenFile error")
		} else {
			return tmpBlockFile, true, nil
		}
	}

	//tmp := []byte(strhash)
	//strhash2 := string(tmp[0:IpfsHashLen])
	c := exec.Command(gStrIpfsName, "get", "-o=secondCacheInfo.gb", strhash) //strhash)
	c.Stdout = &out
	c.Stderr = &outerr
	err := c.Run()

	//strErrInfo := outerr.String()

	if err != nil {
		log.Error("ipfs IpfsGetFileCache2ByHash get error", "error", err, "ipfs err", outerr.String())
		return nil, false, err
	}

	tmpBlockFile, errf := os.OpenFile(strTmpCache2File, os.O_RDWR /*os.O_APPEND*/, 0644)

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
				newHash, err1 := IpfsAddNewFile(strTmpCache2File) //"secondCacheInfo.gb"

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

			tmpCache2File, newFileFlg, err = IpfsGetFileCache2ByHash(stCfg.StCahce2Hash[calArrayPos]) //"secondCacheInfo.gb"
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
		err1, _ := insertNewValue(tmpBlock.blockNum, tmpBlock.blockHeadHash, tmpBlock.blockIpfshash, &cache2st.MapList)
		if err1 != nil {
			continue
		}
		//if newNumber {
		cache2st.NumHashStore++
		cache2st.CurCacheBlockNum = tmpBlock.blockNum
		//}
		lastArrayPos = calArrayPos
		stCfg.CurrentBlockNum = tmpBlock.blockNum
	}
	storeCache(cache2st, tmpCache2File)
	tmpCache2File.Close()
	newHash, err := IpfsAddNewFile(strTmpCache2File)
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
func (d *Downloader) IpsfAddNewBlockToCache(stCfg *Cache1StoreCfg, blockNum uint64, headHash common.Hash, fhash Hash) error {

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

	if calArrayPos >= Cache1StoreCache2Num {
		log.Error("ipfs error IpsfAddNewBlockToCache calc error,calArrayPos exeeed capacity", "calArrayPos", calArrayPos)
		return fmt.Errorf("calArrayPos > Cache1StoreCache2Num")
	}

	//var tmpBlockFile *os.File

	tmpBlockFile, newFileFlg, err := IpfsGetFileCache2ByHash(stCfg.StCahce2Hash[calArrayPos]) //stCfg.Cache2FileNum]) //.CurCachehash)
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
	newHash, err := IpfsAddNewFile(strTmpCache2File)

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
func (d *Downloader) IpfsSyncSaveSecondCache(newFlg bool, blockNum uint64, headHash common.Hash, fhash Hash, file *os.File) error {
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
	err, newNumber := insertNewValue(blockNum, headHash, fhash, &cache2st.MapList)

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
		return err
	}
	publishHash := string(out.Bytes()[0 : out.Len()-1]) //0：IpfsHashLen

	out.Reset()
	outerr.Reset()
	//run
	c = exec.Command(d.dpIpfs.StrIPFSExecName, "name", "publish", publishHash) //string(outbuf[:]))
	err = c.Run()
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
	c.Stderr = &outerr
	//err := c.Run()
	outbuf, err := c.Output()

	//strErrInfo := outerr.String()

	//new
	curCache1Info := new(Cache1StoreCfg) // Cache1StoreCfg{}

	if err != nil {
		log.Error("ipfs error IpfsSyncGetFirstCache error", "error", err, "ipfs err", outerr.String())
		return curCache1Info, err
	}
	//ReadJsFile()

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

	c := exec.Command(gStrIpfsName, "get", "-o=ipfsCachecommon/", ipnsPath) //
	//c.Stdout = &out
	c.Stderr = &outerr
	err := c.Run()
	//outbuf, err := c.Output()

	//strErrInfo := outerr.String()

	curLastestInfo := new(LastestBlcokCfg) //LastestBlcokCfg{}
	if err != nil {

		log.Error("ipfs IpfsSyncGetLatestBlock run cmd error", "error", err, "ipfs err", outerr.String())
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
func insertNewValue(blockNum uint64, headHash common.Hash, blockhash Hash, newBlock *BlockStore) (error, bool) {
	var BnumberNoExist bool = false

	if newBlock.Numberstore == nil {
		newBlock.Numberstore = make(map[uint64]NumberMapingCoupledHash) //
	}
	_, ok := (newBlock.Numberstore)[blockNum]
	if ok == false {
		//newBlock.Numberstore = make(map[uint64]NumberMapingCoupledHash)
		var tmpH NumberMapingCoupledHash
		tmpH.Blockhash = make(map[common.Hash]string)
		(newBlock.Numberstore)[blockNum] = tmpH
		//log.Trace("insertNewValue  not exist head hash", "blocknum", blockNum)
		BnumberNoExist = true
	} else {
		_, ok = (newBlock.Numberstore)[blockNum].Blockhash[headHash]
		if ok {

			log.Warn("insertNewValue  Blockhash[headHash] value exist ", "blocknum", blockNum)
			//return fmt.Errorf("block hash already exist")
		}
	}
	newBlock.Numberstore[blockNum].Blockhash[headHash] = string(blockhash[0:IpfsHashLen])
	log.Trace("ipfs insertNewValue  head hash", "blocknum", blockNum, "headHash", headHash, "blockhash", string(blockhash[0:IpfsHashLen]))

	return nil, BnumberNoExist
}

func (d *Downloader) addNewBlockBatchToLastest(curLastestInfo *LastestBlcokCfg, blockList []listBlockInfo) error {
	for _, tmpBlock := range blockList {
		curLastestInfo.CurrentNum = tmpBlock.blockNum
		_, newNumber := insertNewValue(tmpBlock.blockNum, tmpBlock.blockHeadHash, tmpBlock.blockIpfshash, &curLastestInfo.MapList) //
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
func (d *Downloader) IpfsSynInsertNewBlockHashToLastest(curLastestInfo *LastestBlcokCfg, blockNum uint64, headHash common.Hash, blockhash Hash) error {
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
	err, newNumber := insertNewValue(blockNum, headHash, blockhash, &curLastestInfo.MapList) //
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

	log.Warn("~~~~~ ipfs RecvBlockToDeal recv new block ~~~~~~", "blockNum=", blockqueue.Size())

	stCache1Cfg, _, err := d.IpfsSyncGetLatestBlock(0) //"lastestblockInfo.gb"

	if err != nil {

		//if curBlockNum < hasBlocbNum {
		//	log.Warn("ipfs warn RecvBlockToDeal blockNum already exist ", "blockNum=", curBlockNum)
		//return err
		//}
	}

	curlistBlockInfo := make([]listBlockInfo, 0)
	for {
		var tmplistBlockInfo listBlockInfo //:= new(listBlockInfo)
		if blockqueue.Empty() {
			break
		}
		newBlock := blockqueue.PopItem().(*types.Block)
		if newBlock == nil {
			break
		}
		tmplistBlockInfo.blockNum = newBlock.NumberU64()

		tmpBlockFile, errf := os.OpenFile(strNewBlockStoreFile, os.O_WRONLY|os.O_CREATE, 0644) //"NewTmpBlcok.rp"
		if errf != nil {
			log.Error("ipfs RecvBlockToDeal error in open file ", "error=", errf)
			return errf
		}

		errd := rlp.Encode(tmpBlockFile, newBlock)
		log.Debug("ipfs block encode info", "error", errd, "blockNum", tmplistBlockInfo.blockNum)
		tmpBlockFile.Close()
		err = nil
		tmplistBlockInfo.blockIpfshash, err = IpfsAddNewFile(strNewBlockStoreFile)
		if err != nil {
			log.Error("ipfs RecvBlockToDeal error IpfsAddNewFile  ", "error=", err)
			return err
		}
		tmplistBlockInfo.blockHeadHash = newBlock.Hash()
		curlistBlockInfo = append(curlistBlockInfo, tmplistBlockInfo)
	}
	if len(curlistBlockInfo) == 0 {
		return fmt.Errorf("curlistBlockInfo len = 0")
	}

	d.addNewBlockBatchToLastest(stCache1Cfg, curlistBlockInfo)
	//d.IPfsDirectoryUpdate()
	//go func() error {
	dealcacheFunc := func() error {
		//readCacheCg := Cache1StoreCfg{}	//readCache =(*Cache1StoreCfg)readCacheCfg
		readCacheCfg, err := d.IpfsSyncGetFirstCache(0) //"firstCacheInfo.jn"
		if err != nil {
			fmt.Println("cache1 is nil, create it", err)
			log.Debug("ipfs RecvBlockToDeal IpfsSyncGetFirstCache  cache1 is nil, create it  ", "error=", err)
			//readCacheCfg2 := Cache1StoreCfg{}
			readCacheCfg.OriginBlockNum = curlistBlockInfo[0].blockNum
			readCacheCfg.CurrentBlockNum = curlistBlockInfo[0].blockNum
			readCacheCfg.Cache2FileNum = 0
		}
		return d.IpsfAddNewBlockBatchToCache(readCacheCfg, curlistBlockInfo)

	} //()
	err = dealcacheFunc()
	err = d.IPfsDirectoryUpdate()
	if err == nil {
		log.Trace("ipfs RecvBlockToDeal add ipfs sucess")
	} else {
		log.Error("ipfs RecvBlockToDeal add ipfs error ")
	}
	testShowlog++

	if testShowlog == 6 || testShowlog == 18 || testShowlog == 30 || testShowlog == 50 || testShowlog == 70 {
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
			tmpCache2File, _, _ := IpfsGetFileCache2ByHash(tmpCache1.StCahce2Hash[idx]) //secondCacheInfo.gb
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

	bHash, err := IpfsAddNewFile(strNewBlockStoreFile)
	if err != nil {
		log.Error("ipfs RecvBlockToDeal error IpfsAddNewFile  ", "error=", errf)
		return err
	}

	headHash := newBlock.Hash()

	d.IpfsSynInsertNewBlockHashToLastest(stCache1Cfg, curBlockNum, headHash, bHash[0:IpfsHashLen])

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
		}
		return d.IpsfAddNewBlockToCache(readCacheCfg, curBlockNum, headHash, bHash[0:IpfsHashLen]) /* &readCacheCfg,*/

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
	blockFile, err := IpfsGetBlockByHash(blockhash)
	defer func() {
		blockFile.Close()
		os.Remove(blockhash)
	}()
	if err != nil {
		log.Debug(" ipfs GetBlockAndAnalysis error in IpfsGetBlockByHash", "error", err)
		return false
	}
	//
	obj := new(types.Block)
	errd := rlp.Decode(blockFile, obj)

	log.Info("ipfs dencode block info from GetBlockAndAnalysis", "err", errd, "stype", stype, "obj.Header", obj.NumberU64())

	if errd != nil {
		return false
	}
	d.SynOrignDownload(obj)
	return true
}

func (d *Downloader) GetBlockHashFromCache(headhash common.Hash, headNumber uint64) bool {
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
	if headNumber >= gIpfsCache.getipfsCache2.CurCacheBeignNum && gIpfsCache.getipfsCache2.CurCacheBlockNum >= headNumber {
		//map
		_, ok := gIpfsCache.getipfsCache2.MapList.Numberstore[headNumber]
		if ok {
			blockhash, ok := gIpfsCache.getipfsCache2.MapList.Numberstore[headNumber].Blockhash[headhash]
			if ok {
				log.Info("ipfs GetBlockHashFromCache download block from Cache2")
				return d.GetBlockAndAnalysisSend(blockhash, "secondCache")
			} else {
				log.Error("ipfs GetBlockHashFromCache  error map Blockhash", "headNumber", headNumber, "headhash", headhash)
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

// SyncBlockFromIpfs
func (d *Downloader) SyncBlockFromIpfs(headhash common.Hash, headNumber uint64, index int) int {
	//curBlock := (*core.BlockChain).CurrentBlock()
	//CurLocalBlocknum := d.blockchain.CurrentBlock().NumberU64()
	log.Debug(" *** ipfs get download block ***  ", "number", headNumber, "headhash", headhash, "gIpfsCache.bassign", gIpfsCache.bassign)
	if gIpfsCache.bassign {
		bfind := d.GetBlockHashFromCache(headhash, headNumber)
		if bfind {
			return 0
		}
	}

	log.Debug(" ****** ipfs get download block number over ipfs  ******  ", "number", headNumber)
	var err error
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
			blockhash, ok := gIpfsCache.lastestCache.MapList.Numberstore[headNumber].Blockhash[headhash]
			if ok {

				log.Debug(" ipfs  SyncBlockFromIpfs get new block form ipfs use by getlastest", "blockNum", headNumber)
				bsec := d.GetBlockAndAnalysisSend(blockhash, "getlastest")
				if bsec {
					return 0
				}

			} else {
				log.Debug(" ipfs  SyncBlockFromIpfs get new block form error", "headnumber", headNumber, "headhash", headhash)
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
		cache2File, err := IpfsGetBlockByHash(stCache2Infohash)
		if err != nil {
			log.Error(" ipfs  SyncBlockFromIpfs error IpfsGetBlockByHash cache2File", "error", err)
			return 1
		}
		defer cache2File.Close()

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
		if ok {
			blockhash, ok := gIpfsCache.getipfsCache2.MapList.Numberstore[headNumber].Blockhash[headhash]
			if ok {

				log.Debug(" ipfs  SyncBlockFromIpfs get new block form ipfs use by getCache2", "blockNum", headNumber)
				d.GetBlockAndAnalysisSend(blockhash, "getCache2")
				return 0
			} else {
				log.Error("ipfs  SyncBlockFromIpfs  error map Blockhash", "headNumber", headNumber, "headhash", headhash)
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
func (d *Downloader) SynOrignDownload(obj *types.Block) {

	tmp := new(BlockIpfs)
	tmp.Headeripfs = obj.Header()
	tmp.Transactionsipfs = obj.Transactions()
	tmp.Unclesipfs = obj.Uncles()

	log.Debug(" ipfs send new block to syn", "number=%d", tmp.Headeripfs.Number.Uint64())

	d.ipfsBodyCh <- *tmp
}

func (d *Downloader) IpfsProcessRcvHead() {

	log.Debug(" ipfs proc go IpfsProcessRcvHead enter")
	//recvSync := time.NewTicker(5 * time.Second)
	//defer recvSync.Stop()
	flg := 0
	for {
		select {
		//case headers := <-HeaderIpfsCh:
		case headers := <-d.dpIpfs.HeaderIpfsCh: //
			log.Debug(" ipfs proc recv block headers", "len", len(headers))
			//gIpfsCache.bassign = false
			d.dpIpfs.DownMutex.Lock()
			for _, headerd := range headers {
				flg = d.SyncBlockFromIpfs(headerd.Hash(), headerd.Number.Uint64(), 0)
				if flg == 1 {
					failReTrans := &DownloadRetry{
						header:  headerd,
						downNum: 1,
					}
					log.Debug(" ipfs get block from  add retrans", "number", headerd.Number.Uint64())
					d.dpIpfs.DownRetrans.PushBack(failReTrans)
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
func (d *Downloader) SynBlockFormBlockchain() {
	log.Debug(" ipfs proc go SynBlockFormBlockchain enter")
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
func (d *Downloader) SynDealblock() {

	if !atomic.CompareAndSwapInt32(&d.dpIpfs.IpfsDealBlocking, 0, 1) {
		return
	}
	defer atomic.StoreInt32(&d.dpIpfs.IpfsDealBlocking, 0)

	queueBlock := d.blockchain.GetStoreBlockInfo()
	log.Trace(" ipfs get block from blockchain", "blockNum=", queueBlock.Size())

	//队列不空 则弹出元素处理
	if !queueBlock.Empty() {
		d.RecvBlockSaveToipfs(queueBlock)
	}
	/*for !queueBlock.Empty() {

		err := d.RecvBlockToDeal(queueBlock.PopItem().(*types.Block))
		if err != nil {
			log.Error(" ipfs period save block error")
		}

	}*/
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
						tmpReq := element.Value.(*DownloadRetry)
						res := d.SyncBlockFromIpfs(tmpReq.header.Hash(), tmpReq.header.Number.Uint64(), tmpReq.downNum%2)
						log.Debug(" ipfs get block from  retrans SyncBlockFromIpfs ", "res", res, "downNum", tmpReq.downNum, "blockNum", tmpReq.header.Number.Uint64())
						if res == 1 {
							if tmpReq.downNum < 50 {
								//d.dpIpfs.DownRetrans.PushBack(tmp)
								//log.Debug(" ipfs get block from  retrans again", "num", tmp.downNum, "blockNum", tmp.header.Number.Uint64())
							} else {
								d.dpIpfs.DownRetrans.Remove(element)
							}
							tmpReq.downNum++
						} else if res == 0 {
							d.dpIpfs.DownRetrans.Remove(element)
						}
					}
				}
				d.dpIpfs.DownMutex.Unlock()
			}
		}
	}

}
