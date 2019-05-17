// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package types

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/crypto"
	"github.com/MatrixAINetwork/go-matrix/params"
	"runtime"
	"sync"
)

var (
	ErrInvalidChainId = errors.New("invalid chain id for signer")
)

// sigCache is used to cache the derived sender and contains
// the signer used to derive it.
type sigCache struct {
	signer Signer
	from   common.Address
}

//批量解签名
func BatchSender(txser SelfTransactions) {
	var waitG = &sync.WaitGroup{}
	//	maxProcs := runtime.NumCPU() //获取cpu个数
	//	if maxProcs >= 2 {
	//		runtime.GOMAXPROCS(maxProcs - 1) //限制同时运行的goroutines数量
	//	}
	for _, tx := range txser {
		if tx.GetMatrixType() == common.ExtraUnGasMinerTxType || tx.GetMatrixType() == common.ExtraUnGasValidatorTxType ||
			tx.GetMatrixType() == common.ExtraUnGasInterestTxType || tx.GetMatrixType() == common.ExtraUnGasTxsType || tx.GetMatrixType() == common.ExtraUnGasLotteryTxType {
			continue
		}
		sig := NewEIP155Signer(tx.ChainId())
		waitG.Add(1)
		go Sender_self(sig, tx, waitG)
	}
	waitG.Wait()
}

// MakeSigner returns a Signer based on the given chain config and block number.
func MakeSigner(config *params.ChainConfig, blockNumber *big.Int) Signer {
	var signer Signer
	switch {
	case config.IsEIP155(blockNumber):
		signer = NewEIP155Signer(config.ChainId)
	case config.IsHomestead(blockNumber):
		//signer = HomesteadSigner{}
	default:
		//signer = FrontierSigner{}
	}
	return signer
}

// SignTx signs the transaction using the given signer and private key
func SignTx(tx SelfTransaction, s Signer, prv *ecdsa.PrivateKey) (SelfTransaction, error) {
	h := s.Hash(tx)
	sig, err := crypto.Sign(h[:], prv)
	if err != nil {
		return nil, err
	}
	return tx.WithSignature(s, sig)
}

// Sender returns the address derived from the signature (V, R, S) using secp256k1
// elliptic curve and an error if it failed deriving or upon an incorrect
// signature.
//
// Sender may cache the address, allowing it to be used regardless of
// signing method. The cache is invalidated if the cached signer does
// not match the signer used in the current call.
func Sender(signer Signer, tx SelfTransaction) (common.Address, error) {
	if sc := tx.GetFromLoad(); sc != nil {
		sigCache, ok := sc.(sigCache)
		if ok {
			// If the signer used to derive from in a previous
			// call is not the same as used current, invalidate
			// the cache.
			if sigCache.signer.Equal(signer) {
				return sigCache.from, nil
			}
		}
	}

	addr, err := signer.Sender(tx)
	if err != nil {
		return common.Address{}, err
	}
	tx.SetFromLoad(sigCache{signer: signer, from: addr})
	return addr, nil
}

func BatchSender_self(txs []SelfTransaction, signer Signer, fiter func(SelfTransaction) bool) {
	if len(txs) == 0 {
		return
	}
	if signer == nil {
		signer = NewEIP155Signer(txs[0].ChainId())
	}
	var waitG = &sync.WaitGroup{}
	routineNum := len(txs)/100 + 1
	if routineNum > 1 {
		maxProcs := runtime.GOMAXPROCS(0) //获取cpu个数
		if maxProcs >= 2 {
			maxProcs--
		}
		if maxProcs < routineNum {
			routineNum = maxProcs
		}
	}
	routChan := make(chan SelfTransaction, 0)
	for i := 0; i < routineNum; i++ {
		waitG.Add(1)
		go Sender_sub(signer, routChan, waitG)
	}
	for _, tx := range txs {
		if fiter(tx) {
			routChan <- tx
		}
	}
	close(routChan)
	waitG.Wait()
}
func Sender_sub(signer Signer, txChan chan SelfTransaction, waitg *sync.WaitGroup) {
	defer waitg.Done()
	for {
		select {
		case tx, ok := <-txChan:
			if ok {
				if sc := tx.GetFromLoad(); sc != nil {
					sigCache := sc.(sigCache)
					if sigCache.signer.Equal(signer) {
						break
					}
				}

				addr, err := signer.Sender(tx)
				if err != nil {
					break
				}
				tx.SetFromLoad(sigCache{signer: signer, from: addr})

			} else {
				return
			}
		}
	}
}

//
func Sender_self(signer Signer, tx SelfTransaction, waitg *sync.WaitGroup) (common.Address, error) {
	defer waitg.Done()
	if sc := tx.GetFromLoad(); sc != nil {
		sigCache := sc.(sigCache)
		if sigCache.signer.Equal(signer) {
			return sigCache.from, nil
		}
	}

	addr, err := signer.Sender(tx)
	if err != nil {
		return common.Address{}, err
	}
	tx.SetFromLoad(sigCache{signer: signer, from: addr})
	return addr, nil
}

// Signer encapsulates transaction signature handling. Note that this interface is not a
// stable API and may change at any time to accommodate new protocol rules.
type Signer interface {
	// Sender returns the sender address of the transaction.
	Sender(tx SelfTransaction) (common.Address, error)
	// SignatureValues returns the raw R, S, V values corresponding to the
	// given signature.
	SignatureValues(tx SelfTransaction, sig []byte) (r, s, v *big.Int, err error)
	// Hash returns the hash to be signed.
	Hash(tx SelfTransaction) common.Hash
	// Equal returns true if the given signer is the same as the receiver.
	Equal(Signer) bool
}

// EIP155Transaction implements Signer using the EIP155 rules.
type EIP155Signer struct {
	chainId, chainIdMul *big.Int
}

func NewEIP155Signer(chainId *big.Int) EIP155Signer {
	if chainId == nil {
		chainId = new(big.Int)
	}
	return EIP155Signer{
		chainId:    chainId,
		chainIdMul: new(big.Int).Mul(chainId, big.NewInt(2)),
	}
}

func (s EIP155Signer) Equal(s2 Signer) bool {
	eip155, ok := s2.(EIP155Signer)
	return ok && eip155.chainId.Cmp(s.chainId) == 0
}

var big8 = big.NewInt(8)

func (s EIP155Signer) Sender(tx SelfTransaction) (common.Address, error) {
	if tx.ChainId().Cmp(s.chainId) != 0 {
		return common.Address{}, ErrInvalidChainId
	}
	V := new(big.Int).Set(tx.GetTxV())
	V.Sub(V, s.chainIdMul)
	V.Sub(V, big8)
	return recoverPlain(s.Hash(tx), tx.GetTxR(), tx.GetTxS(), V, true)
}

// WithSignature returns a new transaction with the given signature. This signature
// needs to be in the [R || S || V] format where V is 0 or 1.
func (s EIP155Signer) SignatureValues(tx SelfTransaction, sig []byte) (R, S, V *big.Int, err error) {
	if len(sig) != 65 {
		panic(fmt.Sprintf("wrong size for signature: got %d, want 65", len(sig)))
	}
	R = new(big.Int).SetBytes(sig[:32])
	S = new(big.Int).SetBytes(sig[32:64])
	V = new(big.Int).SetBytes([]byte{sig[64] + 27})

	if s.chainId.Sign() != 0 {
		V = big.NewInt(int64(sig[64] + 35))
		V.Add(V, s.chainIdMul)
	}
	return R, S, V, nil
}

// Hash returns the hash to be signed by the sender.
// It does not uniquely identify the transaction.
func (s EIP155Signer) Hash(txer SelfTransaction) common.Hash {
	return rlpHash(txer.GetMakeHashfield(s.chainId))
}

func recoverPlain(sighash common.Hash, R, S, Vb *big.Int, homestead bool) (common.Address, error) {
	if Vb.BitLen() > 8 {
		return common.Address{}, ErrInvalidSig
	}
	V := byte(Vb.Uint64() - 27)
	if !crypto.ValidateSignatureValues(V, R, S, homestead) {
		return common.Address{}, ErrInvalidSig
	}
	// encode the snature in uncompressed format
	r, s := R.Bytes(), S.Bytes()
	sig := make([]byte, 65)
	copy(sig[32-len(r):32], r)
	copy(sig[64-len(s):64], s)
	sig[64] = V
	// recover the public key from the snature
	pub, err := crypto.Ecrecover(sighash[:], sig)
	if err != nil {
		return common.Address{}, err
	}
	if len(pub) == 0 || pub[0] != 4 {
		return common.Address{}, errors.New("invalid public key")
	}
	var addr common.Address
	copy(addr[:], crypto.Keccak256(pub[1:])[12:])
	return addr, nil
}

// 将原来的deriveChainId方法改为deriveChainId1，然后重写deriveChainId方法
func deriveChainId(v *big.Int) *big.Int {
	v1 := new(big.Int).Set(v)
	return deriveChainId1(v1)
}

// deriveChainId derives the chain id from the given v parameter
func deriveChainId1(v *big.Int) *big.Int {
	if v.BitLen() <= 64 {
		v := v.Uint64()
		if v == 27 || v == 28 {
			return new(big.Int)
		}
		return new(big.Int).SetUint64((v - 35) / 2)
	}
	v = new(big.Int).Sub(v, big.NewInt(35))
	return v.Div(v, big.NewInt(2))
}
