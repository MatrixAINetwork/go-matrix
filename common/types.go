// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php

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

	"github.com/matrix/go-matrix/common/hexutil"
	"github.com/matrix/go-matrix/crypto/sha3"
)

const (
	HashLength      = 32
	AddressLength   = 20
	SignatureLength = 65
)

var (
	hashT    = reflect.TypeOf(Hash{})
	addressT = reflect.TypeOf(Address{})
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

//hezi账户属性定义
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

// Sets the hash to the value of b. If b is larger than len(h), 'b' will be cropped (from the left).
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

// Address represents the 20 byte address of an Matrix account.
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
// Matrix address or not.
func IsHexAddress(s string) bool {
	if hasHexPrefix(s) {
		s = s[2:]
	}
	return len(s) == 2*AddressLength && isHex(s)
}

// Get the string representation of the underlying address
func (a Address) Str() string   { return string(a[:]) }
func (a Address) Bytes() []byte { return a[:] }
func (a Address) Big() *big.Int { return new(big.Int).SetBytes(a[:]) }
func (a Address) Hash() Hash    { return BytesToHash(a[:]) }

func (a Address) Equal(other Address) bool {
	return bytes.Equal(a[:], other[:])
}

// Hex returns an EIP55-compliant hex string representation of the address.
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

// Sets the address to the value of b. If b is larger than len(a) it will panic
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

// UnmarshalText decodes the address from hex. The 0x prefix is optional.
func (a *UnprefixedAddress) UnmarshalText(input []byte) error {
	return hexutil.UnmarshalFixedUnprefixedText("UnprefixedAddress", input, a[:])
}

// MarshalText encodes the address as hex.
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
		return nil, fmt.Errorf("Invalid address")
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

// MarshalJSON marshals the original value
func (ma *MixedcaseAddress) MarshalJSON() ([]byte, error) {
	if strings.HasPrefix(ma.original, "0x") || strings.HasPrefix(ma.original, "0X") {
		return json.Marshal(fmt.Sprintf("0x%s", ma.original[2:]))
	}
	return json.Marshal(fmt.Sprintf("0x%s", ma.original))
}

// Address returns the address
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

// ValidChecksum returns true if the address has valid checksum
func (ma *MixedcaseAddress) ValidChecksum() bool {
	return ma.original == ma.addr.Hex()
}

// Original returns the mixed-case input string
func (ma *MixedcaseAddress) Original() string {
	return ma.original
}

/////////// Signature
type Signature [SignatureLength]byte

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
}

//hezi
type Elect1 struct {
	Account string
	Stock   uint16
	Type    ElectRoleType
}

//hezi
type NetTopology1 struct {
	Type            uint8
	NetTopologyData []NetTopologyData1
}

//hezi
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
	CoinType string
	Fromaddr Address
	To_Amont map[Address]*big.Int
}

const (
	StateDBRevocableBtree string = "RevcBTree"
	StateDBTimeBtree      string = "TimeBtree"
)

var (
	BlkMinerRewardAddress     Address = HexToAddress("0x8000000000000000000000000000000000000000") //区块奖励
	BlkValidatorRewardAddress Address = HexToAddress("0x8000000000000000000000000000000000000001") //leader奖励
	TxGasRewardAddress        Address = HexToAddress("0x8000000000000000000000000000000000000002") //交易费
	LotteryRewardAddress      Address = HexToAddress("0x8000000000000000000000000000000000000003") //彩票
)

const (
	//byte can not be 1,because 1 is occupied
	ExtraNormalTxType  byte = 0   //普通交易
	ExtraBroadTxType   byte = 1   //广播交易(内部交易，钱包无用)
	ExtraUnGasTxType   byte = 2   //无gas的奖励交易(内部交易，钱包无用)
	ExtraRevocable     byte = 3   //可撤销的交易
	ExtraRevertTxType  byte = 4   //撤销交易
	ExtraAuthTx        byte = 5   //授权委托
	ExtraCancelEntrust byte = 6   //取消委托
	ExtraTimeTxType    byte = 7   //定时交易
	ExtraAItxType      byte = 8   //AI 交易
	ExtraSuperBlockTx  byte = 120 //超级区块交易
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
	From Address
	Adam []AddrAmont
	Tim  uint32
}

type EntrustType struct {
	//委托地址
	EntrustAddres string //被委托人from
	//委托权限
	IsEntrustGas  bool //委托gas
	IsEntrustSign bool //委托签名

	//委托限制
	StartHeight uint64 //委托起始时间
	EndHeight   uint64 //委托结束时间
}

type EntrustType1 struct {
	//委托地址
	EntrustAddres Address //被委托人from
	//委托权限
	IsEntrustGas  bool //委托gas
	IsEntrustSign bool //委托签名
	//IsEntrustTx     bool	//委托交易（取消）
	//委托限制
	//PeerMaxAmount   *big.Int //单笔金额(取消)
	//TotalAmount     *big.Int //总额(取消)
	StartHeight uint64 //委托起始时间
	EndHeight   uint64 //委托结束时间
	//EntrustCount    uint32   //委托次数(取消)
}

type AuthType struct {
	AuthAddres    Address //授权人from
	IsEntrustGas  bool    //委托gas
	IsEntrustSign bool    //委托签名
	StartHeight   uint64  //委托起始时间
	EndHeight     uint64  //委托结束时间
}
