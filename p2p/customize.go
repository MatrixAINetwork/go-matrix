// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package p2p

import (
	"bytes"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/crypto"
	"github.com/MatrixAINetwork/go-matrix/crypto/secp256k1"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/p2p/netutil"

	"github.com/MatrixAINetwork/go-matrix/rlp"
)

type Custsend struct {
	FromIp string
	ToIp   string
	IsTcp  bool
	Data   Data_Format
	Code   uint64
	NodeId string
}

var Custsrv *Server

var RecvChan chan Custsend = make(chan Custsend)
var SendChan chan []Custsend = make(chan []Custsend)

type Data_Format struct {
	Type        uint64
	Seq         uint64
	Data_struct []byte
}

type Slove_Cycle struct {
	b  interface{}
	f  func(interface{}, Custsend) //string->Custsend
	b1 interface{}
	f1 func(interface{}, []byte) error
	b2 interface{}
	f2 func(interface{}, Custsend) //string->Custsend

}

var SC Slove_Cycle

func (this *Slove_Cycle) Register_boot(f func(interface{}, Custsend), b interface{}) {
	SC.b = b
	SC.f = f
}

func (this *Slove_Cycle) OK(Data Custsend) {
	//	var ss Custsend
	//	ss.ToIp = "66.66.66.66"
	//	ss.Data.Type = 1
	this.f(this.b, Data)
}

func (this *Slove_Cycle) Register_verifier(f func(interface{}, Custsend), b interface{}) {
	SC.b2 = b
	SC.f2 = f
}

func (this *Slove_Cycle) Recv_data(Data Custsend) {
	//	var ss Custsend
	//	ss.ToIp = "66.66.66.66"
	//	ss.Data.Type = 1
	this.f2(this.b2, Data)
}

func (this *Slove_Cycle) RegisterReceiveUDP(f1 func(interface{}, []byte) error, b1 interface{}) {
	SC.b1 = b1
	SC.f1 = f1
}

func (this *Slove_Cycle) LoadTx(Data []byte) {
	this.f1(this.b1, Data)
}

const (
	TCPSERV = 0x00
	UDPSERV = 0x01
	VPNSERV = 0x02
	SENSERV = 0x03
)

const (
	macSize  = 256 / 8
	sigSize  = 520 / 8
	headSize = macSize + sigSize // space of packet frame data
)

var (
	headSpace = make([]byte, headSize)

	// Neighbors replies are sent across multiple packets to
	// stay below the 1280 byte limit. We compute the maximum number
	// of entries by stuffing a packet until it grows too large.
	maxNeighbors int
)

type NodeID [NodeIDBits / 8]byte

const NodeIDBits = 512

const defaultMaxUdpConcurrent = 10000

var errPacketTooSmall = errors.New("too small")
var errBadHash = errors.New("bad hash")

/*
type Custsend struct {
	FromIp string
	ToIp   string
	IsTcp  bool
	Data   Data_Format
	Code   uint64
	NodeId string
}
*/
/*
type Custsendd struct {
	FromIp string
	ToIp   string
	IsTcp  bool
	Data   Data_Formatd
	Code   uint64
	NodeId string
}

type Data_Formatd struct {
	Type        uint64
	Seq         uint64
	Data_struct []byte
}*/

type Status_Re struct {
	Ip        string
	Ip_status bool
}

/*
type Data_Format struct {
	Type        uint64
	Seq         uint64
	Data_struct []byte
}
*/
type ElectionMsg struct {
	Ip              string
	PeerId          string
	AnnonceRate     uint32
	NodeId          uint32
	MortgageAccount uint64
	UpTime          uint64
	TxId            uint64
	RecElecMsgTime  uint64
}

type ElectionQueue struct {
	Ems  []ElectionMsg
	Lock sync.RWMutex
}

type Deliver struct {
	Iplist []string
	Data   string
	Flag   int
}

type Custdata struct {
	IP   string
	Data string
}

type Custnode struct {
	IP string
	//	UDP, TCP uint16
	NodeId string
	//	sha      common.Hash
	//	addedAt  time.Time
}

type Custconn struct {
	Nodelist []Custnode
	IsTcp    bool
}

type DataFormat struct {
	Msgtype int
	Msglen  int
	Msgdata int
}

/*
type Custconn struct {
	Ip    string
	IsTcp bool
}
*/

type doconn struct {
	iplist []string
	flag   int
}

type addr struct {
	ip   string
	port int
}

type tcplist struct {
	addtcplist []addr
	Lock       sync.Mutex
}

type udplist struct {
	addudplist []addr
	Lock       sync.Mutex
}
type vpnlist struct {
	addvpnlist []addr
	Lock       sync.Mutex
}
type dodo struct {
	TCPList tcplist
	UDPList udplist
	VPNList vpnlist
}

//udp send data
func Custencodedata(priv *ecdsa.PrivateKey, data string, req interface{}) (packet, hash []byte, err error) {
	b := new(bytes.Buffer)
	b.Write(headSpace)
	b.Write([]byte(data))
	if err := rlp.Encode(b, req); err != nil {
		log.Error("Can't encode diSCv4 packet", "err", err)
		return nil, nil, err
	}
	packet = b.Bytes()
	sig, err := crypto.Sign(crypto.Keccak256(packet[headSize:]), priv)
	if err != nil {
		log.Error("Can't sign diSCv4 packet", "err", err)
		return nil, nil, err
	}
	copy(packet[macSize:], sig)
	// add the hash to the front. Note: this doesn't protect the
	// packet in any way. Our public key will be part of this hash in
	// The future.
	hash = crypto.Keccak256(packet[macSize:])
	copy(packet, hash)
	return packet, hash, nil
}

//udp receive data

func recoverNodeID(hash, sig []byte) (id NodeID, err error) {
	pubkey, err := secp256k1.RecoverPubkey(hash, sig)
	if err != nil {
		return id, err
	}
	if len(pubkey)-1 != len(id) {
		return id, fmt.Errorf("recovered pubkey has %d bits, want %d bits", len(pubkey)*8, (len(id)+1)*8)
	}
	for i := range id {
		id[i] = pubkey[i+1]
	}
	return id, nil
}

type rpcEndpoint struct {
	IP  net.IP // len 4 for IPv4 or 16 for IPv6
	UDP uint16 // for diSCovery protocol
	TCP uint16 // for RLPx protocol
}

type CustUpdPacket struct {
	Version    uint
	From, To   rpcEndpoint
	Expiration uint64
	// Ignore additional fields (for forward compatibility).
	Data Custsend
	Rest []rlp.RawValue `rlp:"tail"`
}

func CustdecodePacket(buf []byte, Ip string) (*CustUpdPacket, NodeID, []byte, error) {
	if len(buf) < headSize+1 {
		return nil, NodeID{}, nil, errPacketTooSmall
	}
	//fmt.Println(buf)

	hash, sig, sigdata := buf[:macSize], buf[macSize:headSize], buf[headSize:]
	//	fmt.Println("*******hash", hash)
	//	fmt.Println("*******sig", sig)
	//	fmt.Println("*******sigdata", sigdata)
	shouldhash := crypto.Keccak256(buf[macSize:])
	//shouldhash := crypto.Keccak256(buf[macSize:headSize])
	if !bytes.Equal(hash, shouldhash) {
		return nil, NodeID{}, nil, errBadHash
	}
	fromID, err := recoverNodeID(crypto.Keccak256(buf[headSize:]), sig)
	FromNodeID := fmt.Sprintf("%x", fromID[:])
	//	var noID = (*string)(unsafe.Pointer(&fromID))
	//	fmt.Println("*******", noID, sig)
	if err != nil {
		return nil, NodeID{}, hash, err
	}

	var req *CustUpdPacket
	switch ptype := sigdata[0]; ptype {
	case 1:
		//fmt.Println("rrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrr")
		req = new(CustUpdPacket)
		s := rlp.NewStream(bytes.NewReader(sigdata[1:]), 0)
		err = s.Decode(req)
		if err != nil {
			return nil, NodeID{}, nil, errBadHash
		}
		req.Data.NodeId = FromNodeID
		req.Data.FromIp = Ip
		RecvChan <- req.Data
	/*
		case pongPacket:
			req = new(pong)
		case findnodePacket:
			req = new(findnode)
		case neighborsPacket:
			req = new(neighbors)
	*/
	case 4:
		req = new(CustUpdPacket)
		s := rlp.NewStream(bytes.NewReader(sigdata[1:]), 0)
		err = s.Decode(req)
		if err != nil {
			return nil, NodeID{}, nil, errBadHash
		}
		req.Data.NodeId = FromNodeID
		req.Data.FromIp = Ip
		SC.LoadTx(req.Data.Data.Data_struct)
	case 5:
		log.Info("------->Recv msg from verifier!")
		req = new(CustUpdPacket)
		s := rlp.NewStream(bytes.NewReader(sigdata[1:]), 0)
		err = s.Decode(req)
		if err != nil {
			return nil, NodeID{}, nil, errBadHash
		}
		req.Data.NodeId = FromNodeID
		req.Data.FromIp = Ip
		SC.Recv_data(req.Data)
	default:
		return req, fromID, hash, fmt.Errorf("unknown type: %d", ptype)
	}

	//	s := rlp.NewStream(bytes.NewReader(sigdata[1:]), 0)
	//	err = s.Decode(req)

	//	return req, fromID, hash, err
	return req, fromID, hash, err
}

func Receiveudp() error {
	//fmt.Println("receiveudp------------------------------------------")
	//	addr, err := net.ResolveUDPAddr("udp", "50505")
	//	conn, err := net.ListenUDP("udp", addr)
	//	udpaddr, err := net.ResolveUDPAddr("udp", "192.168.3.222:50505")

	udpaddr, err := net.ResolveUDPAddr("udp", ":40404")
	//	checkError(err)

	conn, err := net.ListenUDP("udp", udpaddr)

	if err != nil {
		///fmt.Println("******************", err)
		return err
	}
	//	realaddr := conn.LocalAddr().(*net.UDPAddr)

	tokens := defaultMaxUdpConcurrent

	slots := make(chan struct{}, tokens)
	for i := 0; i < tokens; i++ {
		slots <- struct{}{}
	}

	for {
		<-slots

		buf := make([]byte, 65535)

		nbytes, from, err := conn.ReadFromUDP(buf)
		//fmt.Println(buf[:nbytes])
		Ip := from.IP.String()
		if netutil.IsTemporaryError(err) {
			// Ignore temporary read errors.
			log.Debug("Temporary UDP read error", "err", err)
			continue
		} else if err != nil {
			// Shut down the loop for permament errors.
			log.Debug("UDP read error", "err", err)
			return err
		}

		go func() {
			//fmt.Println(conn.RemoteAddr().String())
			_, _, _, err := CustdecodePacket(buf[:nbytes], Ip)
			//log.Info("***************", "hash", hash, "ss", ss, "err", err)
			//fromID, hash, ss, err := CustdecodePacket(buf[:nbytes])
			//log.Info("***************", "fromID", fromID.Data, "hash", hash, "ss", ss, "err", err)

			//log.Info("-------------------", "nbytes", nbytes, "from", from)
			if err != nil {
				log.Debug("Bad diSCv4 packet", "addr", from, "err", err)
				slots <- struct{}{}
			}
			slots <- struct{}{}
		}()
	}
	return nil
}

///////////////////////////////////////////////////////////////////

//udp send data

func CustencodePacket(priv *ecdsa.PrivateKey, ptype byte, req interface{}) (packet, hash []byte, err error) {

	b := new(bytes.Buffer)
	b.Write(headSpace)
	b.WriteByte(ptype)

	//fmt.Println(b, b.Len())

	if err := rlp.Encode(b, req); err != nil {
		fmt.Println("Can't encode diSCv4 packet", "err", err)
		return nil, nil, err
	}

	//fmt.Println(b, b.Len())

	//fmt.Println(b.Bytes())
	packet = b.Bytes()
	sig, err := crypto.Sign(crypto.Keccak256(packet[headSize:]), priv)
	if err != nil {
		log.Error("Can't sign diSCv4 packet", "err", err)
		return nil, nil, err
	}
	//fmt.Println(sig)
	copy(packet[macSize:], sig)
	//fmt.Println(packet)

	//fmt.Println(b, b.Len())
	hash = crypto.Keccak256(packet[macSize:])
	copy(packet, hash)
	//fmt.Println(b, b.Len())
	return packet, hash, nil
}

type packet interface {
	//	handle(t *udp, from *net.UDPAddr, fromID NodeID, mac []byte) error
	name() string
}

func (c *Config) name() string {
	if c.Name == "" {
		progname := strings.TrimSuffix(filepath.Base(os.Args[0]), ".exe")
		if progname == "" {
			panic("empty executable name, set Config.Name")
		}
		return progname
	}
	return c.Name
}

func makeEndpoint(addr *net.UDPAddr, tcpPort uint16) rpcEndpoint {
	ip := addr.IP.To4()
	if ip == nil {
		ip = addr.IP.To16()
	}
	return rpcEndpoint{IP: ip, UDP: uint16(addr.Port), TCP: tcpPort}
}

const expiration = 20 * time.Second

const (
	datadirPrivateKey      = "nodekey"            // Path within the datadir to the node's private key
	datadirDefaultKeyStore = "keystore"           // Path within the datadir to the keystore
	datadirStaticNodes     = "static-nodes.json"  // Path within the datadir to the static node list
	datadirTrustedNodes    = "trusted-nodes.json" // Path within the datadir to the trusted node list
	datadirNodeDatabase    = "nodes"              // Path within the datadir to store the node infos
)

//node config
type CustConfig struct {
	Name string `toml:"-"`

	UserIdent string `toml:",omitempty"`

	Version string `toml:"-"`

	DataDir string

	CustConf Config
}

var isOldGmanResource = map[string]bool{
	"chaindata":          true,
	"nodes":              true,
	"nodekey":            true,
	"static-nodes.json":  true,
	"trusted-nodes.json": true,
}

func (c *CustConfig) instanceDir() string {
	if c.DataDir == "" {
		return ""
	}
	return filepath.Join(c.DataDir, c.name())
}

func (c *CustConfig) resolvePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	if c.DataDir == "" {
		return ""
	}
	// Backwards-compatibility: ensure that data directory files created
	// by gman 1.4 are used if they exist.
	if c.name() == "gman" && isOldGmanResource[path] {
		oldpath := ""
		if c.Name == "gman" {
			oldpath = filepath.Join(c.DataDir, path)
		}
		if oldpath != "" && common.FileExist(oldpath) {
			// TODO: print warning
			return oldpath
		}
	}
	return filepath.Join(c.instanceDir(), path)
}

func (c *CustConfig) name() string {
	if c.Name == "" {
		progname := strings.TrimSuffix(filepath.Base(os.Args[0]), ".exe")
		if progname == "" {
			panic("empty executable name, set Config.Name")
		}
		return progname
	}
	return c.Name
}

func CustNodeKey() *ecdsa.PrivateKey {
	var c CustConfig
	c.DataDir = "C:\\Users\\user\\go\\src\\github.com\\matrix\\go-matrix\\cmd\\gman"
	// Use any specifically configured key.
	if c.CustConf.PrivateKey != nil {
		return c.CustConf.PrivateKey
	}
	// Generate ephemeral key if no datadir is being used.
	if c.DataDir == "" {
		key, err := crypto.GenerateKey()
		if err != nil {
			log.Crit(fmt.Sprintf("Failed to generate ephemeral node key: %v", err))
		}
		return key
	}

	keyfile := c.resolvePath(datadirPrivateKey)
	//fmt.Println(keyfile)
	keyfile = "C:\\Users\\user\\go\\src\\github.com\\matrix\\go-matrix\\cmd\\gman\\chaindata\\debug\\nodekey"
	if key, err := crypto.LoadECDSA(keyfile); err == nil {
		return key
	}
	// No persistent key found, generate and store a new one.
	key, err := crypto.GenerateKey()

	//log.Info("**********************NodeKey", "node/config.go GenerateKey", key)
	if err != nil {
		log.Crit(fmt.Sprintf("Failed to generate node key: %v", err))
	}
	//	instanceDir := filepath.Join(c.DataDir, c.name())
	instanceDir := filepath.Join(c.DataDir, "chaindata\\debug")
	//fmt.Println(instanceDir)
	if err := os.MkdirAll(instanceDir, 0700); err != nil {
		log.Error(fmt.Sprintf("Failed to persist node key: %v", err))
		return key
	}
	keyfile = filepath.Join(instanceDir, datadirPrivateKey)
	if err := crypto.SaveECDSA(keyfile, key); err != nil {
		log.Error(fmt.Sprintf("Failed to persist node key: %v", err))
	}
	//log.Info("**********************NodeKey", "node/config.go return key", key)
	return key
}

func CustUdpSend(data Custsend) error {

	var mtoaddr net.UDPAddr
	mtoaddr.IP = []byte("192.168.3.7:50505")
	var ourEndpoint rpcEndpoint
	ourEndpoint.IP = mtoaddr.IP
	ourEndpoint.UDP = 1
	ourEndpoint.TCP = 0
	var toaddr net.UDPAddr
	toaddr.IP = []byte(data.ToIp)
	toaddr.Port = 50505

	req := &CustUpdPacket{
		Version:    1,
		From:       ourEndpoint,
		To:         makeEndpoint(&toaddr, 0),
		Data:       data,
		Expiration: uint64(time.Now().Add(expiration).Unix()),
	}

	prk := Custsrv.PrivateKey //CustNodeKey()
	//fmt.Println("*****************", "prk", prk)
	packet, _, err := CustencodePacket(prk, byte(data.Code), *req)
	if err != nil {
		log.Error("p2p CustUdpSend: err", "err", err)
	}

	var a int
	if a == 0 {
		//fmt.Println(hash)
	}
	if err != nil {
		return err
	}

	dest := data.ToIp + ":50505"
	//fmt.Println("***************To Ip", data.ToIp)
	conn, err := net.Dial("udp", dest)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer conn.Close()

	conn.Write(packet)
	//fmt.Println("张伟：发送成功", packet)
	return nil
}

func CustSend() {
	for {
		select {
		case SendData := <-SendChan:
			//fmt.Println("张伟：resvdata", SendData)
			for _, elm := range SendData {
				//		if elm.IsTcp == false {
				CustUdpSend(elm)
				//		}
			}
		}
	}
}

func CustStat(qestat []string) []Status_Re {

	var restat []Status_Re
	for _, elm := range qestat {
		dest := elm + ":50505"
		conn, err := net.DialTimeout("tcp", dest, 2*time.Second)
		var tem Status_Re
		if err != nil {
			log.Info("********************", "CustStat err", err)
			tem.Ip = elm
			tem.Ip_status = false
			restat = append(restat, tem)
			//			conn.Close()
			continue
		}

		tem.Ip = elm
		tem.Ip_status = true
		restat = append(restat, tem)
		conn.Close()
	}
	return restat
}

/*
func main() {

	/////////////////////////////

	if len(*emailAddress) == 0 {
		log.Info("Missing required --email-address parameter")
	}

	var toaddr net.UDPAddr
	toaddr.IP = []byte("192.168.3.222")
	toaddr.Port = 50505

	pingg("bedf4041cdafe7ea1178b936dab972a0db1009855456c1f60c59e0a8ccb5dcbc8d617ea8a3b507356750557434aaf22bf5e262970cb6a51312bb28dfca300a02", &toaddr)


}
*/
