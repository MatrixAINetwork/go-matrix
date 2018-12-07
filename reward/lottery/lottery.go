package lottery

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/log"
	"math/big"
	"math/rand"
	"sort"
)

const (
	N           = 6
	FIRST       = 1 //一等奖数目
	SECOND      = 2 //二等奖数目
	THIRD       = 3 //三等奖数目
	PackageName = "彩票奖励"
)

var (
	FIRSTPRIZE   *big.Int = big.NewInt(6e+18) //一等奖金额  5man
	SENCONDPRIZE *big.Int = big.NewInt(3e+18) //二等奖金额 2man
	THIRDPRIZE   *big.Int = big.NewInt(1e+18) //三等奖金额 1man
)

type TxCmpResult struct {
	Tx        types.SelfTransaction
	CmpResult uint64
}

// A slice of Pairs that implements sort.Interface to sort by Value.
type TxCmpResultList []TxCmpResult

func (p TxCmpResultList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p TxCmpResultList) Len() int           { return len(p) }
func (p TxCmpResultList) Less(i, j int) bool { return p[i].CmpResult < p[j].CmpResult }

type ChainReader interface {
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
	txsCmpResultList := tlr.getLotteryList(num, N)
	tlr.lotteryChoose(txsCmpResultList, LotteryAccount)

	return LotteryAccount
}

func (tlr *TxsLottery) getLotteryList(num uint64, lotteryNum int) TxCmpResultList {
	originBlockNum := common.GetLastReElectionNumber(num) - 1

	if num < common.GetReElectionInterval() {
		originBlockNum = 0
	}
	randSeed := tlr.seed.GetSeed(num)
	rand.Seed(randSeed.Int64())
	txsCmpResultList := make(TxCmpResultList, 0)
	for originBlockNum < num {
		txs := tlr.chain.GetBlockByNumber(originBlockNum).Transactions()
		for _, tx := range txs {
			extx := tx.GetMatrix_EX()
			if (extx != nil) && len(extx) > 0 && extx[0].TxType == common.ExtraNormalTxType || extx == nil {
				txCmpResult := TxCmpResult{tx, tx.Hash().Big().Uint64()}
				txsCmpResultList = append(txsCmpResultList, txCmpResult)
			}

		}
		originBlockNum++
	}
	if 0 == len(txsCmpResultList) {
		return nil
	}
	sort.Sort(txsCmpResultList)
	chooseResultList := make(TxCmpResultList, 0)
	for i := 0; i < lotteryNum && i < len(txsCmpResultList); i++ {
		randUint64 := rand.Uint64()
		index := randUint64 % (uint64(len(txsCmpResultList)))
		log.INFO(PackageName, "交易序号", index)
		chooseResultList = append(chooseResultList, txsCmpResultList[index])
	}

	return chooseResultList
}

func (tlr *TxsLottery) lotteryChoose(txsCmpResultList TxCmpResultList, LotteryAccountMap map[string]map[common.Address]*big.Int) {
	firstLottery := make(map[common.Address]*big.Int, FIRST)
	secondLottery := make(map[common.Address]*big.Int, SECOND)
	thirdLottery := make(map[common.Address]*big.Int, THIRD)
	for _, v := range txsCmpResultList {
		from := v.Tx.From()

		//抽取一等奖
		LotteryAccount, _ := LotteryAccountMap["First"]
		if len(LotteryAccount) < FIRST {

			firstLottery[from] = FIRSTPRIZE
			LotteryAccountMap["First"] = firstLottery
			log.INFO(PackageName, "一等奖", from.String(), "金额", FIRSTPRIZE)
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
			LotteryAccountMap["Second"] = secondLottery
			log.INFO(PackageName, "二等奖", from.String(), "金额", SENCONDPRIZE)

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
			LotteryAccountMap["third"] = thirdLottery
			log.INFO(PackageName, "三等奖", from.Hex())
			continue
		}
		break

	}

}
