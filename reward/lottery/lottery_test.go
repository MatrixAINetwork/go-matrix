package lottery

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/log"
	"math/big"
	"strconv"
	"testing"
)


type Chain struct {

}

type randSeed  struct{

}
func (r* randSeed)GetSeed(num uint64) *big.Int{

	return big.NewInt(1000)
}

func (chain *Chain)  GetBlockByNumber(num uint64 ) *types.Block{
	header := &types.Header{

	}
    txs:=make([]types.SelfTransaction,0)
    if num==298{
	for i := 0; i < 25; i++ {

		tx:=types.NewTransactions(uint64(i),common.Address{},big.NewInt(100), 100,big.NewInt(int64(100)),nil,nil,0,common.ExtraNormalTxType)
		addr :=common.Address{}
		addr.SetString(strconv.Itoa(i))
		tx.SetFromLoad(addr)
		txs=append(txs,tx)

	}
}

	return types.NewBlockWithTxs(header, txs)
}

func TestTxsLottery_LotteryCalc(t *testing.T) {
	log.InitLog(3)
	lotterytest:=New(&Chain{},&randSeed{})
	lotterytest.LotteryCalc(299)

}
