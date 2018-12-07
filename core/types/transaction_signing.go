// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php

package types

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/crypto"
	"github.com/matrix/go-matrix/params"
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

	//if sc := tx.from.Load(); sc != nil {
	//	sigCache := sc.(sigCache)
	//	// If the signer used to derive from in a previous
	//	// call is not the same as used current, invalidate
	//	// the cache.
	//	if sigCache.signer.Equal(signer) {
	//		return sigCache.from, nil
	//	}
	//}
	//
	//addr, err := signer.Sender(tx)
	//if err != nil {
	//	return common.Address{}, err
	//}
	//tx.from.Store(sigCache{signer: signer, from: addr})
	//return addr, nil
}

//YY
func Sender_self(signer Signer, tx SelfTransaction, waitg *sync.WaitGroup) (common.Address, error) {
	defer waitg.Done()
	if sc := tx.GetFromLoad(); sc != nil {
		sigCache := sc.(sigCache)
		// If the signer used to derive from in a previous
		// call is not the same as used current, invalidate
		// the cache.
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
	//if sc := tx.from.Load(); sc != nil {
	//	sigCache := sc.(sigCache)
	//	// If the signer used to derive from in a previous
	//	// call is not the same as used current, invalidate
	//	// the cache.
	//	if sigCache.signer.Equal(signer) {
	//		return sigCache.from, nil
	//	}
	//}
	//
	//addr, err := signer.Sender(tx)
	//if err != nil {
	//	return common.Address{}, err
	//}
	//tx.from.Store(sigCache{signer: signer, from: addr})
	//return addr, nil
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
	//YY=====begin======
	V := new(big.Int).Set(tx.GetTxV())
	if V.Cmp(big.NewInt(128)) > 0 {
		V.Sub(V, big.NewInt(128))
	}
	V.Sub(V, s.chainIdMul)
	//=======end========
	V.Sub(V, big8)
	return recoverPlain(s.Hash(tx), tx.GetTxR(), tx.GetTxS(), V, true)
	//if !tx.Protected() {
	//	return HomesteadSigner{}.Sender(tx)
	//}
	//if tx.ChainId().Cmp(s.chainId) != 0 {
	//	return common.Address{}, ErrInvalidChainId
	//}
	////YY=====begin======
	//V := new(big.Int).Set(tx.data.V)
	//if V.Cmp(big.NewInt(128)) > 0 {
	//	V.Sub(V, big.NewInt(128))
	//}
	//V.Sub(V, s.chainIdMul)
	////V := new(big.Int).Sub(tx.data.V, s.chainIdMul) 注释原来的方式
	////=======end========
	//V.Sub(V, big8)
	//return recoverPlain(s.Hash(tx), tx.data.R, tx.data.S, V, true)
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

	//if err != nil {
	//	return nil, nil, nil, err
	//}
	if s.chainId.Sign() != 0 {
		V = big.NewInt(int64(sig[64] + 35))
		V.Add(V, s.chainIdMul)
	}
	return R, S, V, nil
}

// Hash returns the hash to be signed by the sender.
// It does not uniquely identify the transaction.
func (s EIP155Signer) Hash(txer SelfTransaction) common.Hash {
	switch txer.TxType() {
	case NormalTxIndex:
		tx, ok := txer.(*Transaction)
		if !ok {
			return common.Hash{}
		}
		if tx.Mtype == true {
			var data1 txdata1
			TxdataAddresToString(tx.Currency, &tx.data, &data1)
			return rlpHash([]interface{}{
				data1.AccountNonce,
				data1.Price,
				data1.GasLimit,
				data1.Recipient,
				data1.Amount,
				data1.Payload,
				s.chainId, uint(0), uint(0),
				data1.TxEnterType,
				data1.IsEntrustTx,
				data1.CommitTime,
				data1.Extra,
			})

		} else {
			return rlpHash([]interface{}{
				tx.data.AccountNonce,
				tx.data.Price,
				tx.data.GasLimit,
				tx.data.Recipient,
				tx.data.Amount,
				tx.data.Payload,
				s.chainId, uint(0), uint(0),
				tx.data.TxEnterType,
				tx.data.IsEntrustTx,
				tx.data.CommitTime,
				tx.data.Extra,
			})
		}
	case BroadCastTxIndex:
		tx, ok := txer.(*TransactionBroad)
		if !ok {
			return common.Hash{}
		}
		return rlpHash([]interface{}{
			tx.data.AccountNonce,
			tx.data.Price,
			tx.data.GasLimit,
			tx.data.Recipient,
			tx.data.Amount,
			tx.data.Payload,
			tx.data.Extra,
			s.chainId, uint(0), uint(0),
		})
	default:
		return common.Hash{}
	}

}

// HomesteadTransaction implements TransactionInterface using the
// homestead rules.
//type HomesteadSigner struct{ FrontierSigner }
//
//func (s HomesteadSigner) Equal(s2 Signer) bool {
//	_, ok := s2.(HomesteadSigner)
//	return ok
//}
//
//// SignatureValues returns signature values. This signature
//// needs to be in the [R || S || V] format where V is 0 or 1.
//func (hs HomesteadSigner) SignatureValues(tx *Transaction, sig []byte) (r, s, v *big.Int, err error) {
//	return hs.FrontierSigner.SignatureValues(tx, sig)
//}
//
//func (hs HomesteadSigner) Sender(tx *Transaction) (common.Address, error) {
//	return recoverPlain(hs.Hash(tx), tx.data.R, tx.data.S, tx.data.V, true)
//}
//
//type FrontierSigner struct{}
//
//func (s FrontierSigner) Equal(s2 Signer) bool {
//	_, ok := s2.(FrontierSigner)
//	return ok
//}

// SignatureValues returns signature values. This signature
// needs to be in the [R || S || V] format where V is 0 or 1.
//func (fs FrontierSigner) SignatureValues(tx *Transaction, sig []byte) (r, s, v *big.Int, err error) {
//	if len(sig) != 65 {
//		panic(fmt.Sprintf("wrong size for signature: got %d, want 65", len(sig)))
//	}
//	r = new(big.Int).SetBytes(sig[:32])
//	s = new(big.Int).SetBytes(sig[32:64])
//	v = new(big.Int).SetBytes([]byte{sig[64] + 27})
//	return r, s, v, nil
//}
//
//// Hash returns the hash to be signed by the sender.
//// It does not uniquely identify the transaction.
//func (fs FrontierSigner) Hash(tx *Transaction) common.Hash {
//	return rlpHash([]interface{}{
//		tx.data.AccountNonce,
//		tx.data.Price,
//		tx.data.GasLimit,
//		tx.data.Recipient,
//		tx.data.Amount,
//		tx.data.Payload,
//	})
//}
//
//func (fs FrontierSigner) Sender(tx *Transaction) (common.Address, error) {
//	return recoverPlain(fs.Hash(tx), tx.data.R, tx.data.S, tx.data.V, false)
//}

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

//YY 将原来的deriveChainId方法改为deriveChainId1，然后重写deriveChainId方法
func deriveChainId(v *big.Int) *big.Int {
	v1 := new(big.Int).Set(v)
	tmp := big.NewInt(128)
	if v1.Cmp(tmp) > 0 {
		v1.Sub(v1, tmp)
	}
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
