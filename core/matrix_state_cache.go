package core

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/rawdb"
	"github.com/matrix/go-matrix/core/state"
	"github.com/pkg/errors"
)

func (bc *BlockChain) MatrixState() (*state.StateDB, error) {
	root := bc.GetMatrixRoot()
	if (root == common.Hash{}) {
		return nil, errors.New("get current block matrix root err!")
	}
	return bc.MatrixStateAt(root)
}

func (bc *BlockChain) MatrixStateAt(root common.Hash) (*state.StateDB, error) {
	return state.New(root, bc.matrixCache)
}

func (bc *BlockChain) GetMatrixRoot() common.Hash {
	return bc.GetMatrixRootByBlock(bc.GetCurrentHash())
}

func (bc *BlockChain) GetMatrixRootByBlock(blockHash common.Hash) common.Hash {
	if matrixRoot, ok := bc.matrixRootCache.Get(blockHash); ok {
		return matrixRoot.(common.Hash)
	}
	block := bc.GetBlockByHash(blockHash)
	if block == nil {
		return common.Hash{}
	}
	matrixRoot := rawdb.ReadMatrixRoot(bc.db, block.Hash(), block.NumberU64())
	if (matrixRoot != common.Hash{}) {
		bc.matrixRootCache.Add(block.Hash(), matrixRoot)
	}
	return matrixRoot
}

func (bc *BlockChain) GetMatrixRootByNumber(number uint64) common.Hash {
	block := bc.GetBlockByNumber(number)
	if block == nil {
		return common.Hash{}
	}
	if matrixRoot, ok := bc.matrixRootCache.Get(block.Hash()); ok {
		return matrixRoot.(common.Hash)
	}
	matrixRoot := rawdb.ReadMatrixRoot(bc.db, block.Hash(), block.NumberU64())
	if (matrixRoot != common.Hash{}) {
		bc.matrixRootCache.Add(block.Hash(), matrixRoot)
	}
	return matrixRoot
}
