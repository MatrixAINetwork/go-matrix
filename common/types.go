// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package common

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"math/rand"
	"reflect"
	"strings"

	"bytes"

	"unicode"

	"github.com/MatrixAINetwork/go-matrix/common/hexutil"
	"github.com/MatrixAINetwork/go-matrix/crypto/sha3"
)

const (
	HashLength      = 32
	AddressLength   = 20
	SignatureLength = 65
)

var (
	hashT      = reflect.TypeOf(Hash{})
	addressT   = reflect.TypeOf(Address{})
	signatureT = reflect.TypeOf(Signature{})
)

const (
	MainAccount     = iota //主账户
	FreezeAccount          //冻结账户
	LockAccount            //锁仓账户
	WithdrawAccount        //可撤销账户
	EntrustAccount         //委托账户
)

var LastAccount uint32 = EntrustAccount //必须赋值最后一个账户

// Hash represents the 32 byte Keccak256 hash of arbitrary data.
type Hash [HashLength]byte

//账户属性定义
type BalanceSlice struct {
	AccountType uint32
	Balance     *big.Int
}
type BalanceType []BalanceSlice

func BytesToHash(b []byte) Hash {
	var h Hash
	h.SetBytes(b)
	return h
}

func BigToHash(b *big.Int) Hash { return BytesToHash(b.Bytes()) }
func HexToHash(s string) Hash   { return BytesToHash(FromHex(s)) }

// Get the string representation of the underlying hash
func (h Hash) Str() string   { return string(h[:]) }
func (h Hash) Bytes() []byte { return h[:] }
func (h Hash) Big() *big.Int { return new(big.Int).SetBytes(h[:]) }
func (h Hash) Hex() string   { return hexutil.Encode(h[:]) }

// TerminalString implements log.TerminalStringer, formatting a string for console
// output during logging.
func (h Hash) TerminalString() string {
	return fmt.Sprintf("%x…%x", h[:3], h[29:])
}

// String implements the stringer interface and is used also by the logger when
// doing full logging into a file.
func (h Hash) String() string {
	return h.Hex()
}

// Format implements fmt.Formatter, forcing the byte slice to be formatted as is,
// without going through the stringer interface used for logging.
func (h Hash) Format(s fmt.State, c rune) {
	fmt.Fprintf(s, "%"+string(c), h[:])
}

// UnmarshalText parses a hash in hex syntax.
func (h *Hash) UnmarshalText(input []byte) error {
	return hexutil.UnmarshalFixedText("Hash", input, h[:])
}

// UnmarshalJSON parses a hash in hex syntax.
func (h *Hash) UnmarshalJSON(input []byte) error {
	return hexutil.UnmarshalFixedJSON(hashT, input, h[:])
}

// MarshalText returns the hex representation of h.
func (h Hash) MarshalText() ([]byte, error) {
	return hexutil.Bytes(h[:]).MarshalText()
}

// Sets the hash to the Value of b. If b is larger than len(h), 'b' will be cropped (from the left).
func (h *Hash) SetBytes(b []byte) {
	if len(b) > len(h) {
		b = b[len(b)-HashLength:]
	}

	copy(h[HashLength-len(b):], b)
}

// Set string `s` to h. If s is larger than len(h) s will be cropped (from left) to fit.
func (h *Hash) SetString(s string) { h.SetBytes([]byte(s)) }

// Sets h to other
func (h *Hash) Set(other Hash) {
	for i, v := range other {
		h[i] = v
	}
}

// Generate implements testing/quick.Generator.
func (h Hash) Generate(rand *rand.Rand, size int) reflect.Value {
	m := rand.Intn(len(h))
	for i := len(h) - 1; i > m; i-- {
		h[i] = byte(rand.Uint32())
	}
	return reflect.ValueOf(h)
}

func (h Hash) Equal(other Hash) bool {
	return bytes.Equal(h[:], other[:])
}

func EmptyHash(h Hash) bool {
	return h == Hash{}
}

// UnprefixedHash allows marshaling a Hash without 0x prefix.
type UnprefixedHash Hash

// UnmarshalText decodes the hash from hex. The 0x prefix is optional.
func (h *UnprefixedHash) UnmarshalText(input []byte) error {
	return hexutil.UnmarshalFixedUnprefixedText("UnprefixedHash", input, h[:])
}

// MarshalText encodes the hash as hex.
func (h UnprefixedHash) MarshalText() ([]byte, error) {
	return []byte(hex.EncodeToString(h[:])), nil
}

/////////// Address

// Address represents the 20 byte Address of an Matrix account.
type Address [AddressLength]byte

func BytesToAddress(b []byte) Address {
	var a Address
	a.SetBytes(b)
	return a
}

func HashToAddress(hash Hash) Address {
	return BytesToAddress(hash[11:])
}

func BigToAddress(b *big.Int) Address { return BytesToAddress(b.Bytes()) }
func HexToAddress(s string) Address   { return BytesToAddress(FromHex(s)) }

// IsHexAddress verifies whether a string can represent a valid hex-encoded
// Matrix Address or not.
func IsHexAddress(s string) bool {
	if hasHexPrefix(s) {
		s = s[2:]
	}
	return len(s) == 2*AddressLength && isHex(s)
}

// Get the string representation of the underlying Address
func (a Address) Str() string   { return string(a[:]) }
func (a Address) Bytes() []byte { return a[:] }
func (a Address) Big() *big.Int { return new(big.Int).SetBytes(a[:]) }
func (a Address) Hash() Hash    { return BytesToHash(a[:]) }

func (a Address) Equal(other Address) bool {
	return bytes.Equal(a[:], other[:])
}

// Hex returns an EIP55-compliant hex string representation of the Address.
func (a Address) Hex() string {
	unchecksummed := hex.EncodeToString(a[:])
	sha := sha3.NewKeccak256()
	sha.Write([]byte(unchecksummed))
	hash := sha.Sum(nil)

	result := []byte(unchecksummed)
	for i := 0; i < len(result); i++ {
		hashByte := hash[i/2]
		if i%2 == 0 {
			hashByte = hashByte >> 4
		} else {
			hashByte &= 0xf
		}
		if result[i] > '9' && hashByte > 7 {
			result[i] -= 32
		}
	}
	return "0x" + string(result)
}

// String implements the stringer interface and is used also by the logger.
func (a Address) String() string {
	return a.Hex()
}

// Format implements fmt.Formatter, forcing the byte slice to be formatted as is,
// without going through the stringer interface used for logging.
func (a Address) Format(s fmt.State, c rune) {
	fmt.Fprintf(s, "%"+string(c), a[:])
}

// Sets the Address to the Value of b. If b is larger than len(a) it will panic
func (a *Address) SetBytes(b []byte) {
	if len(b) > len(a) {
		b = b[len(b)-AddressLength:]
	}
	copy(a[AddressLength-len(b):], b)
}

// Set string `s` to a. If s is larger than len(a) it will panic
func (a *Address) SetString(s string) { a.SetBytes([]byte(s)) }

// Sets a to other
func (a *Address) Set(other Address) {
	for i, v := range other {
		a[i] = v
	}
}

// MarshalText returns the hex representation of a.
func (a Address) MarshalText() ([]byte, error) {
	return hexutil.Bytes(a[:]).MarshalText()
}

// UnmarshalText parses a hash in hex syntax.
func (a *Address) UnmarshalText(input []byte) error {
	return hexutil.UnmarshalFixedText("Address", input, a[:])
}

// UnmarshalJSON parses a hash in hex syntax.
func (a *Address) UnmarshalJSON(input []byte) error {
	return hexutil.UnmarshalFixedJSON(addressT, input, a[:])
}

// UnprefixedHash allows marshaling an Address without 0x prefix.
type UnprefixedAddress Address

// UnmarshalText decodes the Address from hex. The 0x prefix is optional.
func (a *UnprefixedAddress) UnmarshalText(input []byte) error {
	return hexutil.UnmarshalFixedUnprefixedText("UnprefixedAddress", input, a[:])
}

// MarshalText encodes the Address as hex.
func (a UnprefixedAddress) MarshalText() ([]byte, error) {
	return []byte(hex.EncodeToString(a[:])), nil
}

// MixedcaseAddress retains the original string, which may or may not be
// correctly checksummed
type MixedcaseAddress struct {
	addr     Address
	original string
}

// NewMixedcaseAddress constructor (mainly for testing)
func NewMixedcaseAddress(addr Address) MixedcaseAddress {
	return MixedcaseAddress{addr: addr, original: addr.Hex()}
}

// NewMixedcaseAddressFromString is mainly meant for unit-testing
func NewMixedcaseAddressFromString(hexaddr string) (*MixedcaseAddress, error) {
	if !IsHexAddress(hexaddr) {
		return nil, fmt.Errorf("Invalid Address")
	}
	a := FromHex(hexaddr)
	return &MixedcaseAddress{addr: BytesToAddress(a), original: hexaddr}, nil
}

// UnmarshalJSON parses MixedcaseAddress
func (ma *MixedcaseAddress) UnmarshalJSON(input []byte) error {
	if err := hexutil.UnmarshalFixedJSON(addressT, input, ma.addr[:]); err != nil {
		return err
	}
	return json.Unmarshal(input, &ma.original)
}

// MarshalJSON marshals the original Value
func (ma *MixedcaseAddress) MarshalJSON() ([]byte, error) {
	if strings.HasPrefix(ma.original, "0x") || strings.HasPrefix(ma.original, "0X") {
		return json.Marshal(fmt.Sprintf("0x%s", ma.original[2:]))
	}
	return json.Marshal(fmt.Sprintf("0x%s", ma.original))
}

// Address returns the Address
func (ma *MixedcaseAddress) Address() Address {
	return ma.addr
}

// String implements fmt.Stringer
func (ma *MixedcaseAddress) String() string {
	if ma.ValidChecksum() {
		return fmt.Sprintf("%s [chksum ok]", ma.original)
	}
	return fmt.Sprintf("%s [chksum INVALID]", ma.original)
}

// ValidChecksum returns true if the Address has valid checksum
func (ma *MixedcaseAddress) ValidChecksum() bool {
	return ma.original == ma.addr.Hex()
}

// Original returns the mixed-case input string
func (ma *MixedcaseAddress) Original() string {
	return ma.original
}

/////////// Signature
type Signature [SignatureLength]byte

func (a *Signature) UnmarshalJSON(input []byte) error {
	return hexutil.UnmarshalFixedJSON(signatureT, input, a[:])
}

/*
func (a *Signature) MarshalJSON() ([]byte, error) {
	return hexutil.Bytes(a[:]).MarshalText()
}
*/
func (a Signature) MarshalText() ([]byte, error) {
	return hexutil.Bytes(a[:]).MarshalText()
}

func BytesToSignature(b []byte) Signature {
	var s Signature
	s.SetBytes(b)
	return s
}

func (s Signature) Str() string   { return string(s[:]) }
func (s Signature) Bytes() []byte { return s[:] }

func (s *Signature) SetBytes(b []byte) {
	if len(b) > len(s) {
		b = b[len(b)-SignatureLength:]
	}

	copy(s[SignatureLength-len(b):], b)
}

// Sets h to other
func (h *Signature) Set(other Signature) {
	for i, v := range other {
		h[i] = v
	}
}

type VerifiedSign struct {
	Sign     Signature `json:"sign"`
	Account  Address   `json:"account"`
	Validate bool      `json:"validate"`
	Stock    uint16    `json:"stock"`
}

type VerifiedSign1 struct {
	Sign     Signature `json:"sign"`
	Account  string    `json:"account"`
	Validate bool      `json:"validate"`
	Stock    uint16    `json:"stock"`
}

//
type Elect struct {
	Account Address
	Stock   uint16
	Type    ElectRoleType
	VIP     VIPRoleType
}

//
type Elect1 struct {
	Account string
	Stock   uint16
	Type    ElectRoleType
	VIP     VIPRoleType
}

//
type NetTopology1 struct {
	Type            uint8
	NetTopologyData []NetTopologyData1
}

//
type NetTopologyData1 struct {
	Account  string
	Position uint16
}

const (
	PosOffline uint16 = 0xF000
	PosOnline  uint16 = 0xF001
)

type NetTopologyData struct {
	Account  Address
	Position uint16
}

const (
	NetTopoTypeAll    uint8 = 0
	NetTopoTypeChange uint8 = 1
)

type NetTopology struct {
	Type            uint8
	NetTopologyData []NetTopologyData
}
type RewarTx struct {
	CoinRange string
	CoinType  string
	Fromaddr  Address
	To_Amont  map[Address]*big.Int
	RewardTyp byte
}

const (
	StateDBRevocableBtree string = "RevcBTree"
	StateDBTimeBtree      string = "TimeBtree"
)

var (
	BlkMinerRewardAddress     Address = HexToAddress("0x8000000000000000000000000000000000000000") //区块奖励
	BlkValidatorRewardAddress Address = HexToAddress("0x8000000000000000000000000000000000000000") //leader奖励
	InterestRewardAddress     Address = HexToAddress("0x8000000000000000000000000000000000000000") //利息
	TxGasRewardAddress        Address = HexToAddress("0x8000000000000000000000000000000000000001") //交易费
	LotteryRewardAddress      Address = HexToAddress("0x8000000000000000000000000000000000000002") //彩票
	ContractAddress           Address = HexToAddress("0x000000000000000000000000000000000000000A") //合约账户

	DestroyAddress Address = HexToAddress("0xA27d3632A283c138C2F78ba21d1e473a500e4AF3") //创建币种的转账地址（MAN.3GJF5vPbrmbUG7ZTFyFogdiuXY3Lp）
)

const (
	ExtraNormalTxType         byte = 0   //普通交易
	ExtraBroadTxType          byte = 1   //广播交易(内部交易，钱包无用)
	ExtraUnGasMinerTxType     byte = 2   //矿工奖励类型
	ExtraRevocable            byte = 3   //可撤销的交易
	ExtraRevertTxType         byte = 4   //撤销交易
	ExtraAuthTx               byte = 5   //授权委托
	ExtraCancelEntrust        byte = 6   //取消委托
	ExtraTimeTxType           byte = 7   //定时交易
	ExtraAItxType             byte = 8   //AI 交易
	ExtraMakeCoinType         byte = 9   //创建币种交易
	ExtraUnGasValidatorTxType byte = 10  //验证者奖励类型
	ExtraUnGasInterestTxType  byte = 11  //利息奖励通过合约交易发放
	ExtraUnGasTxsType         byte = 12  //交易费奖励类型
	ExtraUnGasLotteryTxType   byte = 13  //彩票奖励类型
	ExtraSetBlackListTxType   byte = 14  //设置黑名单交易
	ExtraSuperBlockTx         byte = 120 //超级区块交易
)

var (
	WhiteAddrlist     = [1]Address{InterestRewardAddress}
	RewardAccounts    = [5]Address{BlkMinerRewardAddress, BlkValidatorRewardAddress, TxGasRewardAddress, LotteryRewardAddress, InterestRewardAddress}
	ConsensusAccounts []Address
	BlackList         []Address
	BlackListString   []string
	WorkPath          string
)

const (
	RewardMinerType     byte = 0 //矿工奖励类型
	RewardValidatorType byte = 1 //验证者奖励类型
	RewardInterestType  byte = 2 //利息奖励通过合约交易发放
	RewardTxsType       byte = 3 //交易费奖励类型
	RewardLotteryType   byte = 4 //彩票费奖励类型
)

type TxTypeInt uint8
type RetCallTxN struct {
	TXt   byte
	ListN []uint32
}
type AddrAmont struct {
	Addr  Address
	Amont *big.Int
}

type RecorbleTx struct {
	From    Address
	Cointyp string
	Adam    []AddrAmont
	Tim     uint32
	Typ     byte
}

//地址为matrix地址
type EntrustType struct {
	//委托地址
	EntrustAddres string //被委托人from
	//委托权限
	IsEntrustGas    bool //委托gas
	IsEntrustSign   bool //委托签名
	EnstrustSetType byte //0-按高度委托,1-按时间委托,2-按次数委托

	//委托限制
	StartHeight  uint64 //委托起始高度
	EndHeight    uint64 //委托结束高度
	StartTime    uint64
	EndTime      uint64
	EntrustCount uint32 //委托次数
}

type AuthType struct {
	AuthAddres      Address //授权人from
	EnstrustSetType byte    //0-按高度委托,1-按时间委托
	IsEntrustGas    bool    //委托gas
	IsEntrustSign   bool    //委托签名
	StartHeight     uint64  //委托起始高度
	EndHeight       uint64  //委托结束高度
	StartTime       uint64
	EndTime         uint64
	EntrustCount    uint32 //授权委托次数
}

type CoinRoot struct {
	Cointyp     string
	Root        Hash
	TxHash      Hash
	ReceiptHash Hash
	Bloom       [256]byte
}
type Coinbyte struct {
	Root    Hash
	Byte256 []Hash
}

type CoinSharding struct {
	CoinType  string
	Shardings []uint
}

type SMakeCoin struct {
	CoinName    string
	AddrAmount  map[string]*hexutil.Big
	CoinUnit    *hexutil.Big //单位
	PackNum     uint64
	CoinAddress Address
	//CoinTotal *big.Int  //总发行量
	PayCoinType string
}

type BroadTxkey struct {
	Key     string
	Address Address
}
type BroadTxValue struct {
	Key   BroadTxkey
	Value []byte
}

func Greater(a, b BroadTxkey) bool {
	if a.Key > b.Key {
		return true
	} else if a.Key == b.Key {
		return bytes.Compare(a.Address[:], b.Address[:]) > 0
	}
	return false
}
func Less(a, b BroadTxkey) bool {
	if a.Key < b.Key {
		return true
	} else if a.Key == b.Key {
		return bytes.Compare(a.Address[:], b.Address[:]) < 0
	}
	return false
}

type BroadTxSlice []BroadTxValue

func (si *BroadTxSlice) Insert(key string, address Address, value []byte) {
	insValue := BroadTxValue{BroadTxkey{key, address}, value}
	index, exist := find(insValue.Key, si)
	if exist {
		(*si)[index] = insValue
	} else {
		insert(si, index, insValue)
	}
}
func (si *BroadTxSlice) FindKey(key string) map[Address][]byte {
	firstKey := BroadTxkey{key, Address{}}
	endKey := BroadTxkey{key, Address{}}
	for i := 0; i < len(endKey.Address); i++ {
		endKey.Address[i] = 0xff
	}
	first, exist := find(firstKey, si)
	last, exist1 := find(endKey, si)
	if exist {
		first--
	}
	if exist1 {
		last++
	}
	valueMap := make(map[Address][]byte, last-first)
	for ; first < last; first++ {
		valueMap[(*si)[first].Key.Address] = (*si)[first].Value
	}
	return valueMap
}
func (si *BroadTxSlice) FindValue(key string, address Address) ([]byte, bool) {
	index, exist := find(BroadTxkey{key, address}, si)
	if exist {
		return (*si)[index].Value, true
	} else {
		return nil, false
	}
}
func find(k BroadTxkey, info *BroadTxSlice) (int, bool) {
	left, right, mid := 0, len(*info)-1, 0
	if right < 0 {
		return 0, false
	}
	for {
		mid = (left + right) / 2
		if Greater((*info)[mid].Key, k) {
			right = mid - 1
		} else if Less((*info)[mid].Key, k) {
			left = mid + 1
		} else {
			return mid, true
		}
		if left > right {
			return left, false
		}
	}
	return mid, false
}

//binary insert
func insert(info *BroadTxSlice, index int, value BroadTxValue) {
	*info = append(*info, value)
	end := len(*info) - 1
	for i := end; i > index; i-- {
		(*info)[i], (*info)[i-1] = (*info)[i-1], (*info)[i]
	}
}

//长度为3-8位,不能有小写字母，不能有特殊字符，不能有数字，不能有连续的"MAN"
func IsValidityCurrency(s string) bool {
	if len(s) < 3 || len(s) > 8 {
		return false
	}

	for i := 0; i < len(s); i++ {
		ch := s[i]
		if !unicode.IsLetter(int32(ch)) {
			return false
		}
		if !unicode.IsUpper(int32(ch)) {
			return false
		}
	}
	if strings.Contains(s, "MAN") {
		return false
	}
	return true
}

//长度为3-8位,不能有小写字母，不能有特殊字符，不能有数字
func IsValidityManCurrency(s string) bool {
	if len(s) < 3 || len(s) > 8 {
		return false
	}

	for i := 0; i < len(s); i++ {
		ch := s[i]
		if !unicode.IsLetter(int32(ch)) {
			return false
		}
		if !unicode.IsUpper(int32(ch)) {
			return false
		}
	}
	return true
}

type CoinConfig struct {
	CoinRange   string       `json:"CoinRange"`   //coinrange和cointype是一个类型，为了扩展方便保留该字段
	CoinType    string       `json:"CoinType"`    //支付币种
	PackNum     uint64       `json:"PackNum"`     //打包数量限制 如果为0则不打包
	CoinUnit    *hexutil.Big `json:"CoinUnit"`    //单位
	CoinTotal   *hexutil.Big `json:"CoinTotal"`   //总发行量
	CoinAddress Address      `json:"CoinAddress"` //币种交易费账户
	//PayCoinType	string 		 `json:"PayCoinType"` //发放币种
}

const COINPREFIX string = "ms_"

type LinkInfo struct {
	Sbs uint64
	Bn  uint64
	Bt  uint64
}

func IsGreaterLink(linkA, linkB LinkInfo) bool {
	if linkA.Sbs > linkB.Sbs {
		return true
	} else if linkA.Sbs == linkB.Sbs {
		if linkA.Bn > linkB.Bn {
			return true
		} else if linkA.Bn == linkB.Bn {
			if linkA.Bt > linkB.Bt {
				return true
			}
		}
	}
	return false
}

type OperationalInterestSlash struct {
	DepositAmount *big.Int
	OperAmount    *big.Int //用来增加的利息金额或者惩罚的金额
	Position      uint64
	DepositType   uint64
}

//Operational Interest and Slash
type CalculateDeposit struct {
	AddressA0   Address
	CalcDeposit []OperationalInterestSlash
}

//退选信息
type WithDrawInfo struct {
	WithDrawAmount *big.Int
	WithDrawTime   uint64 //退选时间
}

//没把定期DepositAmount和WithDrawTime放到ZeroDepositlist里，因为这样处理方便，暂时这样用
type DepositMsg struct {
	DepositType      uint64 //0-活期,1-定期1个月,3-定期3个月,6-定期6个月
	DepositAmount    *big.Int
	Interest         *big.Int
	Slash            *big.Int
	BeginTime        uint64 //定期起始时间，为当前确认时间(evm.Time)
	EndTime          uint64 //定期到期时间，(BeginTime+定期时长)
	Position         uint64 //仓位
	WithDrawInfolist []WithDrawInfo
}
type DepositBase struct {
	AddressA0     Address
	AddressA1     Address
	OnlineTime    *big.Int
	Role          *big.Int
	PositionNonce uint64
	Dpstmsg       []DepositMsg
}
type DepositRoles struct {
	Role    *big.Int
	Address []Address
}
type CheckDepositInfo struct {
	DepositAmount *big.Int
	Withdraw      uint64
	Role          *big.Int
	AddressA1     Address
}

//type WithDrawMsg struct {
//	DepositAmount *big.Int
//	WithDrawTime 	uint64
//}
