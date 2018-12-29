package election

import (
	"testing"

	"fmt"
	"github.com/matrix/go-matrix/common/mt19937"
	"github.com/matrix/go-matrix/log"
	"math/big"

	"github.com/matrix/go-matrix/baseinterface"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/vm"
	_ "github.com/matrix/go-matrix/election/layered"
	_ "github.com/matrix/go-matrix/election/nochoice"
	_ "github.com/matrix/go-matrix/election/stock"
	"github.com/matrix/go-matrix/run/utils"

	"encoding/json"
	"github.com/matrix/go-matrix/core"
	"github.com/matrix/go-matrix/mc"
	"io/ioutil"
	"os"
	"strconv"
)

func GetDepositDetatil(num int, m int, n int, onlineFlag bool) []vm.DepositDetail {
	mList := []vm.DepositDetail{}
	for i := 0; i < num; i++ {
		temp := vm.DepositDetail{}
		temp.Address = common.BigToAddress(big.NewInt(int64(i)))

		if m > 0 {
			temp.Deposit = new(big.Int).Mul(big.NewInt(10000000), common.ManValue)
			m--
		} else if n > 0 {
			temp.Deposit = new(big.Int).Mul(big.NewInt(1000000), common.ManValue)
			n--
		} else {
			temp.Deposit = new(big.Int).Mul(big.NewInt(10000), common.ManValue)
		}

		if onlineFlag == true {
			temp.OnlineTime = big.NewInt(int64(i))
		} else {
			temp.OnlineTime = big.NewInt(int64(0))
		}

		temp.WithdrawH = big.NewInt(int64(i))

		tNodeID := "000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
		if i < 10 {
			tNodeID += "0"
		}
		tNodeID += strconv.Itoa(i)
		//fmt.Println("i", i, "err", err, len(tNodeID), "nodeId-string", tNodeID, "address-string", temp.Address.String())

		mList = append(mList, temp)

	}
	return mList
}

func MakeValidatorTopReq(num int, Seed uint64, vip1Num int, vip2Num int, white []common.Address, black []common.Address, onlineFlag bool) *mc.MasterValidatorReElectionReqMsg {
	mList := GetDepositDetatil(num, vip1Num, vip2Num, onlineFlag)

	ans := &mc.MasterValidatorReElectionReqMsg{
		SeqNum:        Seed,
		RandSeed:      big.NewInt(int64(Seed)),
		ValidatorList: mList,
		//	FoundationValidatoeList: []vm.DepositDetail{},
	}
	ans.ElectConfig = mc.ElectConfigInfo_All{
		ValidatorNum:  11,
		BackValidator: 5,
		WhiteList:     white,
		BlackList:     black,
	}
	ans.VIPList = []mc.VIPConfig{

		mc.VIPConfig{
			MinMoney:     0,
			InterestRate: 100,
			ElectUserNum: 0,
			StockScale:   1000,
		},
		mc.VIPConfig{
			MinMoney:     1000000,
			InterestRate: 100,
			ElectUserNum: 3,
			StockScale:   1700,
		},
		mc.VIPConfig{
			MinMoney:     10000000,
			InterestRate: 100,
			ElectUserNum: 5,
			StockScale:   2000,
		},
	}
	return ans

}
func MakeMinerTopReq(num int, Seed uint64, vip1Num int, vip2Num int, white []common.Address, black []common.Address, onlineFlag bool) *mc.MasterMinerReElectionReqMsg {
	mList := GetDepositDetatil(num, vip1Num, vip2Num, onlineFlag)

	ans := &mc.MasterMinerReElectionReqMsg{
		SeqNum:    Seed,
		RandSeed:  big.NewInt(int64(Seed)),
		MinerList: mList,
	}
	ans.ElectConfig = mc.ElectConfigInfo_All{
		ValidatorNum:  11,
		BackValidator: 5,
		MinerNum:      21,
		WhiteList:     white,
		BlackList:     black,
	}
	return ans

}

func PrintMiner(miner *mc.MasterMinerReElectionRsp) {

	fmt.Println("MasterMiner")
	for _, v := range miner.MasterMiner {
		fmt.Println(v.Account, v.Position, v.Type, v.Stock)
	}
	fmt.Println("BackUpMiner")
	fmt.Println("\n\n\n\n")

}

func PrintValidator(validator *mc.MasterValidatorReElectionRsq) {

	fmt.Println("MasterValidator")
	for _, v := range validator.MasterValidator {
		fmt.Println(v.Account, v.Position, v.Type, v.Stock)
	}
	fmt.Println("BackupValidator")
	for _, v := range validator.BackUpValidator {
		fmt.Println(v.Account, v.Position, v.Type, v.Stock)
	}

	fmt.Println("CandidateValidator")
	for _, v := range validator.CandidateValidator {
		fmt.Println(v.Account, v.Position, v.Type, v.Stock)
	}
	fmt.Println("\n\n\n\n")
}

func TestUnit1(t *testing.T) {
	////矿工生成单元测试
	//
	//for Num := 20; Num <= 22; Num++ {
	//	for Key := 101; Key <= 105; Key++ {
	//		req := MakeMinerTopReq(Num, uint64(Key))
	//		fmt.Println("矿工备选列表个数", len(req.MinerList), "随机数", req.RandSeed)
	//		rspMiner := baseinterface.NewElect("layered").MinerTopGen(req)
	//		PrintMiner(rspMiner)
	//	}
	//}

}

func GOTestV(vip1Num int, vip2Num int, white []common.Address, black []common.Address, plug string, onlineFlag bool) {
	//验证者拓扑生成
	mapMaster := make(map[common.Address]int, 0)
	mapBackup := make(map[common.Address]int, 0)
	mapCand := make(map[common.Address]int, 0)
	//股权方案-（10-12）

	for Num := 50; Num <= 50; Num++ {
		for Key := 0; Key < 1; Key++ {
			req := MakeValidatorTopReq(Num, uint64(Key*2000+1), vip1Num, vip2Num, white, black, onlineFlag)
			//if Key==0{
			//	for _,v:=range req.ValidatorList{
			//		fmt.Println("账户",v.Address.String(),"NodeId",v.NodeID.String(),"抵押值",v.Deposit.String(),"在线时长",v.OnlineTime.String(),"withdraw",v.WithdrawH.String())
			//	}
			//}

			rspValidator := baseinterface.NewElect(plug).ValidatorTopGen(req)
			for _, v := range rspValidator.MasterValidator {
				mapMaster[v.Account]++
			}
			for _, v := range rspValidator.BackUpValidator {
				mapBackup[v.Account]++
			}
			for _, v := range rspValidator.CandidateValidator {
				mapCand[v.Account]++
			}
			//fmt.Println(len(rspValidator.MasterValidator),len(rspValidator.BackUpValidator),len(rspValidator.CandidateValidator))
			//PrintValidator(rspValidator)
		}
	}

	ListAddr := []common.Address{}
	for i := 0; i < 50; i++ {
		ListAddr = append(ListAddr, common.BigToAddress(big.NewInt(int64(i))))
	}
	all := 0
	fmt.Println()
	for _, v := range ListAddr {
		fmt.Println("账户", v.String(), "选择验证者次数", mapMaster[v], "选择备份验证者次数", mapBackup[v], "选择候选验证者次数", mapCand[v])
		all += mapMaster[v]
		all += mapBackup[v]
		all += mapCand[v]
	}
	fmt.Println("所有节点被选中的总次数", all)
}

func GOTestM(vip1Num int, vip2Num int, white []common.Address, black []common.Address, plug string, onlineFlag bool) {
	//矿工拓扑生成
	mapMaster := make(map[common.Address]int, 0)

	for Num := 50; Num <= 50; Num++ {
		for Key := 0; Key < 1000; Key++ {
			req := MakeMinerTopReq(Num, uint64(Key*2000+1), vip1Num, vip2Num, white, black, onlineFlag)
			//if Key==0{
			//	for _,v:=range req.ValidatorList{
			//		fmt.Println("账户",v.Address.String(),"NodeId",v.NodeID.String(),"抵押值",v.Deposit.String(),"在线时长",v.OnlineTime.String(),"withdraw",v.WithdrawH.String())
			//	}
			//}

			rspValidator := baseinterface.NewElect(plug).MinerTopGen(req)
			for _, v := range rspValidator.MasterMiner {
				mapMaster[v.Account]++
			}

			//PrintValidator(rspValidator)
		}
	}

	ListAddr := []common.Address{}
	for i := 0; i < 50; i++ {
		ListAddr = append(ListAddr, common.BigToAddress(big.NewInt(int64(i))))
	}
	all := 0
	fmt.Println()
	for _, v := range ListAddr {
		fmt.Println("账户", v.String(), "选择矿工次数", mapMaster[v])
		all += mapMaster[v]

	}
	fmt.Println("所有节点被选中的总次数", all)
}

func TestUnit2(t *testing.T) {
	log.InitLog(3)
	GOTestV(5, 3, []common.Address{}, []common.Address{}, "layerd", true)
	//	GOTestV(4,4,[]common.Address{},[]common.Address{},"layerd",true)
	//GOTestV(4,4,[]common.Address{},[]common.Address{},"layerd",false)
	//	GOTestV(6,3,[]common.Address{},[]common.Address{},"layerd",true)
	//	GOTestV(6,3,[]common.Address{},[]common.Address{},"layerd",false)
	//GOTestV(5, 3, []common.Address{}, []common.Address{}, "layerd", true)
}
func Test3(t *testing.T) {
	//white:=[]common.Address{
	//	common.BigToAddress(big.NewInt(2)),
	//}
	//black:=[]common.Address{
	//
	//}
	//GOTestV(0,0,white,black,"layerd",true)

	//white:=[]common.Address{
	//	//common.BigToAddress(big.NewInt(2)),
	//}
	//black:=[]common.Address{
	//	common.BigToAddress(big.NewInt(2)),
	//}
	//GOTestV(0,0,white,black,"layerd",true)

	white := []common.Address{
		common.BigToAddress(big.NewInt(1)),
	}
	black := []common.Address{
		common.BigToAddress(big.NewInt(2)),
	}
	GOTestV(0, 0, white, black, "layerd", true)

}
func Test4(t *testing.T) {
	//white:=[]common.Address{
	//	//common.BigToAddress(big.NewInt(1)),
	//}
	//black:=[]common.Address{
	//	//common.BigToAddress(big.NewInt(2)),
	//}
	//GOTestM(0,0,white,black,"layerd",true)
	//
	//white:=[]common.Address{
	////common.BigToAddress(big.NewInt(1)),
	//}
	//black:=[]common.Address{
	//common.BigToAddress(big.NewInt(2)),
	//}
	//GOTestM(0,0,white,black,"layerd",true)

	white := []common.Address{
		common.BigToAddress(big.NewInt(1)),
	}
	black := []common.Address{
		//	common.BigToAddress(big.NewInt(2)),
	}
	GOTestM(0, 0, white, black, "layerd", true)

}

func Test5(t *testing.T) {
	///log.InitLog(3)

	//	GOTestV(0,0,[]common.Address{},[]common.Address{},"nochoice",true)
	//GOTestV(0,0,[]common.Address{},[]common.Address{},"nochoice",false)
	//white:=[]common.Address{
	//	common.BigToAddress(big.NewInt(1)),
	//}
	//black:=[]common.Address{
	//	common.BigToAddress(big.NewInt(2)),
	//}
	//
	//GOTestV(0,0,white,black,"nochoice",false)

	//GOTestM(0,0,[]common.Address{},[]common.Address{},"nochoice",true)
	//	GOTestM(0,0,[]common.Address{},[]common.Address{},"nochoice",false)

	white := []common.Address{
		common.BigToAddress(big.NewInt(1)),
	}
	black := []common.Address{
		common.BigToAddress(big.NewInt(2)),
	}
	GOTestM(0, 0, white, black, "nochoice", false)

}
func Test6(t *testing.T) {
	//GOTestV(0,0,[]common.Address{},[]common.Address{},"stock",true)
	//GOTestV(3,0,[]common.Address{},[]common.Address{},"stock",true)
	//GOTestV(3,3,[]common.Address{},[]common.Address{},"stock",true)

	white := []common.Address{
		common.BigToAddress(big.NewInt(1)),
	}
	black := []common.Address{
		common.BigToAddress(big.NewInt(2)),
	}

	GOTestV(0, 0, white, black, "stock", true)
}

func Test7(t *testing.T) {
	//	GOTestM(3,0,[]common.Address{},[]common.Address{},"stock",true)
	//GOTestM(0,0,[]common.Address{},[]common.Address{},"stock",true)
	white := []common.Address{
		common.BigToAddress(big.NewInt(1)),
	}
	black := []common.Address{
		common.BigToAddress(big.NewInt(2)),
	}
	GOTestM(0, 0, white, black, "stock", true)
}

func Savefile(genesis *core.Genesis, filename string) {
	marshalData, err := json.Marshal(genesis)
	err = ioutil.WriteFile(filename, marshalData, os.ModeAppend)
	if err != nil {
		fmt.Println("测试支持", "生成test文件成功")
	}

}
func Savefile1(genesis1 *core.Genesis1, filename string) {
	marshalData, err := json.Marshal(genesis1)
	err = ioutil.WriteFile(filename, marshalData, os.ModeAppend)
	if err != nil {
		fmt.Println("测试支持", "生成test文件成功")
	}

}

func TestDefaultGenesisCfg(t *testing.T) {
	genesisPath := "MANGenesis.json"
	file, err := os.Open(genesisPath)
	if err != nil {
		utils.Fatalf("Failed to read genesis file: %v", err)
	}
	defer file.Close()
	genesis1 := new(core.Genesis1)

	if err := json.NewDecoder(file).Decode(genesis1); err != nil {
		utils.Fatalf("invalid genesis file: %v", err)
	}
	genesis, err := core.GetDefaultGeneis()

	Savefile(genesis, "init.json")
	core.ManGenesisToEthGensis(genesis1, genesis)
	Savefile(genesis, "end.json")
	fmt.Println(genesis.MState.Broadcast)

}

func TestNew(t *testing.T) {
	fmt.Println("daas", 0xffff)
	A := new(core.Genesis1)
	err := json.Unmarshal([]byte(core.DefaultJson), A)
	fmt.Println("err", err)
	fmt.Println(A)
}

func Test111(t *testing.T) {
	aimRatio := []float64{}
	total := -0.02
	for index := 0; index < 30; index++ {
		total += 0.02
		aimRatio = append(aimRatio, total)
	}
	mapUsed := make(map[int]bool)
	rand := mt19937.RandUniformInit(10)
	for time := 0; time < 1000; time++ {
		rr := float64(rand.Uniform(0.0, 1.0))
		fmt.Println("rr", rr)
		for index := len(aimRatio) - 1; index >= 0; index-- {
			if rr > aimRatio[index] {
				mapUsed[index] = true
				fmt.Println("rr", rr, "time", time, "index", index, "aimRatio[index]", aimRatio[index], len(mapUsed))
				break
			}
		}
		if len(mapUsed) == 30 {
			break
		}
	}

}

var mapMoney = make(map[common.Address]int)

func MakeValidatorReq(vipList []mc.VIPConfig) *mc.MasterValidatorReElectionReqMsg {
	blackList := []common.Address{}
	index := []int{90, 88, 86, 47, 1}
	for _, v := range index {
		blackList = append(blackList, common.BigToAddress(big.NewInt(int64(v))))
	}
	req := &mc.MasterValidatorReElectionReqMsg{
		SeqNum:        1,
		RandSeed:      big.NewInt(100),
		ValidatorList: []vm.DepositDetail{},
		ElectConfig: mc.ElectConfigInfo_All{
			MinerNum:      21,
			ValidatorNum:  19,
			BackValidator: 5,
			ElectPlug:     "layerd",
			WhiteList:     []common.Address{},
			BlackList:     []common.Address{},
		},
		VIPList: vipList,
	}
	for index := 10; index <= 49; index++ {
		depos := index * 10000
		req.ValidatorList = append(req.ValidatorList, vm.DepositDetail{

			Address: common.BigToAddress(big.NewInt(int64(len(req.ValidatorList) + 1))),
			Deposit: new(big.Int).Mul(big.NewInt(int64(depos)), common.ManValue),
		})
		mapMoney[common.BigToAddress(big.NewInt(int64(len(req.ValidatorList)+1)))] = depos
	}
	for index := 1000; index <= 1490; index += 10 {
		depos := index * 10000
		req.ValidatorList = append(req.ValidatorList, vm.DepositDetail{
			Address: common.BigToAddress(big.NewInt(int64(len(req.ValidatorList) + 1))),
			Deposit: new(big.Int).Mul(big.NewInt(int64(depos)), common.ManValue),
		})
		mapMoney[common.BigToAddress(big.NewInt(int64(len(req.ValidatorList)+1)))] = depos
	}
	return req
}

var (
	VIPList = [][]mc.VIPConfig{
		[]mc.VIPConfig{
			mc.VIPConfig{
				MinMoney:     0,
				StockScale:   1000,
				ElectUserNum: 0,
			},
			mc.VIPConfig{
				MinMoney:     14710000,
				StockScale:   1000,
				ElectUserNum: 3,
			},
			mc.VIPConfig{
				MinMoney:     14720000,
				StockScale:   1000,
				ElectUserNum: 2,
			},
			mc.VIPConfig{
				MinMoney:     14780000,
				StockScale:   1000,
				ElectUserNum: 2,
			},
		},
		[]mc.VIPConfig{
			mc.VIPConfig{
				MinMoney:     0,
				StockScale:   1000,
				ElectUserNum: 0,
			},
			mc.VIPConfig{
				MinMoney:     14820000,
				StockScale:   1000,
				ElectUserNum: 3,
			},
			mc.VIPConfig{
				MinMoney:     14830000,
				StockScale:   1000,
				ElectUserNum: 2,
			},
			mc.VIPConfig{
				MinMoney:     14900000,
				StockScale:   1000,
				ElectUserNum: 2,
			},
		},
		[]mc.VIPConfig{
			mc.VIPConfig{
				MinMoney:     0,
				StockScale:   1000,
				ElectUserNum: 0,
			},
			mc.VIPConfig{
				MinMoney:     14640000,
				StockScale:   1000,
				ElectUserNum: 3,
			},
			mc.VIPConfig{
				MinMoney:     14650000,
				StockScale:   1000,
				ElectUserNum: 2,
			},
			mc.VIPConfig{
				MinMoney:     14700000,
				StockScale:   1000,
				ElectUserNum: 2,
			},
		},
		[]mc.VIPConfig{
			mc.VIPConfig{
				MinMoney:     0,
				StockScale:   1000,
				ElectUserNum: 0,
			},
			mc.VIPConfig{
				MinMoney:     14720000,
				StockScale:   1000,
				ElectUserNum: 2,
			},
			mc.VIPConfig{
				MinMoney:     14800000,
				StockScale:   1000,
				ElectUserNum: 2,
			},
			mc.VIPConfig{
				MinMoney:     20000000,
				StockScale:   1000,
				ElectUserNum: 2,
			},
		},
		[]mc.VIPConfig{
			mc.VIPConfig{
				MinMoney:     0,
				StockScale:   1000,
				ElectUserNum: 0,
			},
			mc.VIPConfig{
				MinMoney:     14830000,
				StockScale:   1000,
				ElectUserNum: 2,
			},
			mc.VIPConfig{
				MinMoney:     14900000,
				StockScale:   1000,
				ElectUserNum: 2,
			},
			mc.VIPConfig{
				MinMoney:     20000000,
				StockScale:   1000,
				ElectUserNum: 2,
			},
		},
		[]mc.VIPConfig{
			mc.VIPConfig{
				MinMoney:     0,
				StockScale:   1000,
				ElectUserNum: 0,
			},
			mc.VIPConfig{
				MinMoney:     14650000,
				StockScale:   1000,
				ElectUserNum: 2,
			},
			mc.VIPConfig{
				MinMoney:     14700000,
				StockScale:   1000,
				ElectUserNum: 2,
			},
			mc.VIPConfig{
				MinMoney:     20000000,
				StockScale:   1000,
				ElectUserNum: 2,
			},
		},
		[]mc.VIPConfig{
			mc.VIPConfig{
				MinMoney:     0,
				StockScale:   1000,
				ElectUserNum: 0,
			},
			mc.VIPConfig{
				MinMoney:     14800000,
				StockScale:   1000,
				ElectUserNum: 2,
			},
			mc.VIPConfig{
				MinMoney:     20000000,
				StockScale:   1000,
				ElectUserNum: 2,
			},
			mc.VIPConfig{
				MinMoney:     24000000,
				StockScale:   1000,
				ElectUserNum: 2,
			},
		},
		[]mc.VIPConfig{
			mc.VIPConfig{
				MinMoney:     0,
				StockScale:   1000,
				ElectUserNum: 0,
			},
			mc.VIPConfig{
				MinMoney:     14900000,
				StockScale:   1000,
				ElectUserNum: 2,
			},
			mc.VIPConfig{
				MinMoney:     20000000,
				StockScale:   1000,
				ElectUserNum: 2,
			},
			mc.VIPConfig{
				MinMoney:     24000000,
				StockScale:   1000,
				ElectUserNum: 2,
			},
		},

		[]mc.VIPConfig{
			mc.VIPConfig{
				MinMoney:     0,
				StockScale:   1000,
				ElectUserNum: 0,
			},
			mc.VIPConfig{
				MinMoney:     14600000,
				StockScale:   1000,
				ElectUserNum: 2,
			},
			mc.VIPConfig{
				MinMoney:     14850000,
				StockScale:   1000,
				ElectUserNum: 2,
			},
			mc.VIPConfig{
				MinMoney:     14900000,
				StockScale:   1000,
				ElectUserNum: 2,
			},
		},
		[]mc.VIPConfig{
			mc.VIPConfig{
				MinMoney:     0,
				StockScale:   1000,
				ElectUserNum: 0,
			},
			mc.VIPConfig{
				MinMoney:     10000000,
				StockScale:   1000,
				ElectUserNum: 2,
			},
			mc.VIPConfig{
				MinMoney:     14000000,
				StockScale:   1000,
				ElectUserNum: 2,
			},
			mc.VIPConfig{
				MinMoney:     14600000,
				StockScale:   1000,
				ElectUserNum: 2,
			},
		},
		[]mc.VIPConfig{
			mc.VIPConfig{
				MinMoney:     0,
				StockScale:   1000,
				ElectUserNum: 0,
			},
			mc.VIPConfig{
				MinMoney:     14600000,
				StockScale:   1000,
				ElectUserNum: 2,
			},
		},
		[]mc.VIPConfig{
			mc.VIPConfig{
				MinMoney:     0,
				StockScale:   1000,
				ElectUserNum: 0,
			},
		},
	}
)

var (
	VIPList1 = [][]mc.VIPConfig{

		[]mc.VIPConfig{
			mc.VIPConfig{
				MinMoney:     0,
				StockScale:   1000,
				ElectUserNum: 0,
			},
			mc.VIPConfig{
				MinMoney:     14700000,
				StockScale:   1000,
				ElectUserNum: 2,
			},
			mc.VIPConfig{
				MinMoney:     20000000,
				StockScale:   1000,
				ElectUserNum: 2,
			},
			mc.VIPConfig{
				MinMoney:     24000000,
				StockScale:   1000,
				ElectUserNum: 2,
			},
		},
		[]mc.VIPConfig{
			mc.VIPConfig{
				MinMoney:     0,
				StockScale:   1000,
				ElectUserNum: 0,
			},
			mc.VIPConfig{
				MinMoney:     14600000,
				StockScale:   1000,
				ElectUserNum: 2,
			},
			mc.VIPConfig{
				MinMoney:     14850000,
				StockScale:   1000,
				ElectUserNum: 2,
			},
			mc.VIPConfig{
				MinMoney:     14900000,
				StockScale:   1000,
				ElectUserNum: 2,
			},
		},
	}
)

var Black = [][]mc.VIPConfig{
	[]mc.VIPConfig{
		mc.VIPConfig{
			MinMoney:     0,
			StockScale:   1000,
			ElectUserNum: 0,
		},
		mc.VIPConfig{
			MinMoney:     14400000,
			StockScale:   1000,
			ElectUserNum: 2,
		},
		mc.VIPConfig{
			MinMoney:     14600000,
			StockScale:   1000,
			ElectUserNum: 2,
		},
		mc.VIPConfig{
			MinMoney:     14800000,
			StockScale:   1000,
			ElectUserNum: 2,
		},
	},
}

func TestN(t *testing.T) {
	//log.InitLog(3)

	for k, v := range VIPList1 {
		req := MakeValidatorReq(v)
		rspValidator := baseinterface.NewElect("layerd").ValidatorTopGen(req)
		for _, v := range rspValidator.MasterValidator {
			fmt.Println("Account:", v.Account.Big().Uint64(), "Stock:", v.Stock, "vip:", v.VIPLevel, "role:", v.Type)
		}
		for _, v := range rspValidator.BackUpValidator {
			fmt.Println("Account:", v.Account.Big().Uint64(), "Stock:", v.Stock, "vip:", v.VIPLevel, "role:", v.Type)
		}
		for _, v := range rspValidator.CandidateValidator {
			fmt.Println("Account:", v.Account.Big().Uint64(), "Stock:", v.Stock, "vip:", v.VIPLevel, "role:", v.Type)
		}
		fmt.Println("测试结束", k)
	}

}

func MakeMinerReq(vipList []mc.VIPConfig) *mc.MasterMinerReElectionReqMsg {
	blackList := []common.Address{}
	index := []int{90, 88, 86, 47, 1}
	for _, v := range index {
		blackList = append(blackList, common.BigToAddress(big.NewInt(int64(v))))
	}
	req := &mc.MasterMinerReElectionReqMsg{
		SeqNum:    1,
		RandSeed:  big.NewInt(100),
		MinerList: []vm.DepositDetail{},
		ElectConfig: mc.ElectConfigInfo_All{
			MinerNum:      21,
			ValidatorNum:  19,
			BackValidator: 5,
			ElectPlug:     "layerd",
			WhiteList:     []common.Address{},
			BlackList:     []common.Address{},
		},
	}
	for index := 10; index <= 49; index++ {
		depos := index * 10000
		req.MinerList = append(req.MinerList, vm.DepositDetail{

			Address: common.BigToAddress(big.NewInt(int64(len(req.MinerList) + 1))),
			Deposit: new(big.Int).Mul(big.NewInt(int64(depos)), common.ManValue),
		})
		mapMoney[common.BigToAddress(big.NewInt(int64(len(req.MinerList)+1)))] = depos
	}
	for index := 1000; index <= 1490; index += 10 {
		depos := index * 10000
		req.MinerList = append(req.MinerList, vm.DepositDetail{
			Address: common.BigToAddress(big.NewInt(int64(len(req.MinerList) + 1))),
			Deposit: new(big.Int).Mul(big.NewInt(int64(depos)), common.ManValue),
		})
		mapMoney[common.BigToAddress(big.NewInt(int64(len(req.MinerList)+1)))] = depos
	}
	return req
}

func TestV(t *testing.T) {
	req := MakeMinerReq(nil)
	rspMiner := baseinterface.NewElect("layerd").MinerTopGen(req)
	for _, v := range rspMiner.MasterMiner {
		fmt.Println("Account:", v.Account.Big().Uint64(), "Stock:", v.Stock, "vip:", v.VIPLevel, "role:", v.Type)
	}

}
