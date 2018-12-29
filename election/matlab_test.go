package election

import (
	"bufio"
	"fmt"
	"github.com/matrix/go-matrix/baseinterface"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/vm"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"io"
	"math"
	"math/big"
	"os"
	"strconv"
	"testing"
)

var (
	infile  = "D:\\Python\\electionv2\\in.json"
	outfile = "./out.json"
	vipList = []mc.VIPConfig{
		mc.VIPConfig{
			MinMoney:     1000000,
			InterestRate: 2000, //(分母待定为1000w)
			ElectUserNum: 3,
			StockScale:   2000, //千分比
		},
		mc.VIPConfig{
			MinMoney:     100000,
			InterestRate: 1200, //(分母待定为1000w)
			ElectUserNum: 3,
			StockScale:   1200, //千分比
		},
		mc.VIPConfig{
			MinMoney:     0,
			InterestRate: 1000, //(分母待定为1000w)
			ElectUserNum: 0,
			StockScale:   1000, //千分比
		},
	}
	electInfo = mc.ElectConfigInfo_All{
		MinerNum:      21,
		ValidatorNum:  11,
		BackValidator: 5,
	}
)

type TestM struct {
	Input  string
	Output string
	VM     [][]vm.DepositDetail
	SeqNum []*big.Int
	result []*mc.MasterValidatorReElectionRsq
}

func NewTestM(input string, output string) *TestM {
	testM := &TestM{
		Input:  input,
		Output: output,
	}
	return testM
}

func ProcessLine(line string) []int {
	ans := []int{}
	t := ""
	for index := 0; index < len(line); index++ {
		c := string(line[index])
		if c == "-" {
			cInt, err := strconv.Atoi(t)
			//fmt.Println(cInt,err)
			if err == nil {
				ans = append(ans, cInt)
			}
			t = ""
		} else {
			t += string(line[index])
		}
	}
	cInt, err := strconv.Atoi(t)
	//fmt.Println(cInt,err)
	if err == nil {
		ans = append(ans, cInt)
	}
	//fmt.Println("anssss",ans)
	return ans
}
func (self *TestM) SetSeqNum(segNum uint64) {
	self.SeqNum = append(self.SeqNum, big.NewInt(int64(segNum)))
}
func (self *TestM) SetDepArrary() {
	VM := []vm.DepositDetail{}
	f, err := os.Open(self.Input)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	rd := bufio.NewReader(f)
	ans := []int{}
	for {
		line, err := rd.ReadString('\n')

		//fmt.Println(line)
		//ProcessLine(line)
		ans = ProcessLine(line)
		//fmt.Println("ans",ans)
		if len(ans) == 4 {
			VM = append(VM, vm.DepositDetail{
				Address: common.BigToAddress(big.NewInt(int64(ans[0]))),
				//NodeID     discover.NodeID
				Deposit:    big.NewInt(int64(ans[2])),
				WithdrawH:  big.NewInt(int64(ans[1])),
				OnlineTime: big.NewInt(int64(ans[3])),
			})
		}
		if len(ans) == 1 {
			self.SetSeqNum(uint64(ans[0]))
			self.VM = append(self.VM, VM)
			VM = []vm.DepositDetail{}
		}
		if err != nil || io.EOF == err {
			break
		}
	}
	if len(ans) == 1 {
		fmt.Println("set随机数")
		self.VM = append(self.VM, VM)
		self.SetSeqNum(uint64(ans[0]))
	}
	//	fmt.Println("VM",VM)

	//fmt.Println("DepArrary",self.VM)
}

func (self *TestM) GetAns() {
	fmt.Println("self.VM len", len(self.VM))
	for k, v := range self.VM {
		data := &mc.MasterValidatorReElectionReqMsg{
			SeqNum:                  self.SeqNum[k].Uint64(),
			RandSeed:                self.SeqNum[k],
			ValidatorList:           v,
			FoundationValidatorList: []vm.DepositDetail{},
			ElectConfig:             electInfo,
			VIPList:                 vipList,
		}
		ans := baseinterface.NewElect("stock").ValidatorTopGen(data)
		self.SloveResult(ans)
	}

}
func (self *TestM) ComputerVal(start string, end string) string {
	flag := false
	f, err := os.Open(self.Output)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	rd := bufio.NewReader(f)
	ans := ""
	for {
		line, err := rd.ReadString('\n')
		if flag == true {
			ans += line
		}
		if line == start {
			flag = false
		}
		if line == end {
			break
		}
		//	ans=ProcessOutLine(line)

		if err != nil || io.EOF == err {
			break
		}
	}
	return ans
}
func (self *TestM) SloveResult(data *mc.MasterValidatorReElectionRsq) bool {
	ans := ""
	vv := data
	ans += "Master:::"
	ans += "\n"
	for _, v := range vv.MasterValidator {
		//fmt.Println("Master",v.Account.Big(),v.Stock)
		ans += v.Account.Big().String()
		ans += "-"
		stock := strconv.Itoa(int(v.Stock))
		ans += stock
		ans += "-"
		ans += "\n"
	}
	ans += "BackUp:::"
	ans += "\n"
	for _, v := range vv.BackUpValidator {
		//	fmt.Println("Back",v.Account.Big(),v.Stock)
		ans += v.Account.Big().String()
		ans += "-"
		stock := strconv.Itoa(int(v.Stock))
		ans += stock
		ans += "-"
		ans += "\n"
	}
	ans += "Candid:::"
	ans += "\n"
	for _, v := range vv.CandidateValidator {
		//fmt.Println("Can",v.Account.Big(),v.Stock)
		ans += v.Account.Big().String()
		ans += "-"
		stock := strconv.Itoa(int(v.Stock))
		ans += stock
		ans += "-"
		ans += "\n"
	}

	//fmt.Println("result",self.result.MasterValidator[0])
	file, err := os.OpenFile("./value.json", os.O_APPEND, 0644)
	if err != nil && os.IsNotExist(err) {
		file, _ = os.Create("./value.json")
		defer file.Close()
	}

	file.Write([]byte(ans))

	return true
}

func Test_Main(t *testing.T) {
	log.InitLog(3)

	testM := NewTestM(infile, outfile)
	testM.SetDepArrary()
	testM.GetAns()
}

func SaveValue() {
	//fmt.Println("+++++++++++++++++++++++++++++++++")
	//ans:=""
	//for _,v:=range CapitalMap{
	//	ans+=v.Addr.Big().String()
	//	ans+="-"
	//	ans+=fmt.Sprintf("%.15f",v.Flot)
	//	ans+="-"
	//}
	//fmt.Println(ans)
	//ans+="\n"
	//
	//
	//file,err:=os.OpenFile("./value.json",os.O_APPEND,0644)
	//if err!=nil&&os.IsNotExist(err){
	//	file,_=os.Create("./value.json")
	//	defer file.Close()
	//}
	//
	//file.Write([]byte(ans))
}
func Round1(f float64, n int) float64 {
	floatStr := fmt.Sprintf("%."+strconv.Itoa(n)+"f", f)
	inst, _ := strconv.ParseFloat(floatStr, 64)
	return inst
}

func Round(f float64, n int) float64 {
	n10 := math.Pow10(n)
	return math.Trunc((f+0.5/n10)*n10) / 10
}

func TestInt(t *testing.T) {
	a1 := 0.04939491232403063
	a2 := 0.04939491232403063

	a11 := strconv.FormatFloat(a1, 'E', -1, 64)
	a22 := strconv.FormatFloat(a2, 'E', -1, 64)
	fmt.Println("a11", a11, "a22", a22)
	if a11 == a22 {
		fmt.Println("111")
	} else {
		fmt.Println("12222")
	}
	fmt.Println(math.Max(a1, a2))
	fmt.Println(math.Dim(a1, a2))
}

func TestInt11(t *testing.T) {
	a1 := 0.04939491232403063
	a2 := 0.04939491232403061

	fmt.Println(math.Max(a1, a2))
	fmt.Println(math.Dim(a1, a2))
}
