// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package support

import (
	"github.com/matrix/go-matrix/core/vm"
	"github.com/matrix/go-matrix/election/support/mt19937"
)

//func (Ele *Elector) ValNodesSelected(probVal []Stf, seed int64) ([]Strallyint, []Strallyint, []Strallyint) {
func ValNodesSelected(probVal []Stf, seed int64, M int, P int, J int) ([]Strallyint, []Strallyint, []Strallyint) {

	//归一化价值函数 成为 采样概率
	probnormalized := Normalize(probVal)
	//	fmt.Println(probnormalized)
	// 选出M+P-J个节点 或者 进行1000次采样
	PricipalValNodes, BakValNodes, RemainingProbNormalizedNodes := SampleMPlusPNodes(probnormalized, seed, M, J) //SampleMPlusPNodes(probnormalized=probnormalized,seed=seed,M=M,J=J,P=5,MaxSample=MaxSample)
	// 计算所有剩余节点的股权RemainingProbNormalizedNodes
	//fmt.Println("PricipalValNodes", len(PricipalValNodes))
	RemainingValNodes := CalcRemainingNodesVotes(RemainingProbNormalizedNodes)

	// 基金会节点加入验证主节点列表
	// 如果验证主节点不足N个,使用剩余节点列表补足M-J个
	for len(PricipalValNodes) < M-J && len(RemainingValNodes) > 0 {
		PricipalValNodes = append(PricipalValNodes, RemainingValNodes[0])
		RemainingValNodes = RemainingValNodes[1:]
	}
	// 如果备份主节点不足P个,使用剩余节点列表补足P个
	for len(BakValNodes) < P && len(RemainingValNodes) > 0 {
		BakValNodes = append(BakValNodes, RemainingValNodes[0])
		RemainingValNodes = RemainingValNodes[1:]
	}
	return PricipalValNodes, BakValNodes, RemainingValNodes
}

func MinerNodesSelected(probVal []Stf, seed int64, Ms int) ([]Strallyint, []Strallyint) {
	probnormalized := Normalize(probVal)

	//fmt.Println(probnormalized)
	PricipalMinerNodes, BakMinerNodes := SampleMinerNodes(probnormalized, seed, Ms)

	//计算所有剩余节点的股权
	BakMinerNodes = CalcRemainingNodesVotes(BakMinerNodes)
	return PricipalMinerNodes, BakMinerNodes
}

type Stf struct {
	Str  string
	Flot float64
}

type pnormalized struct {
	Value  float64
	Nodeid string
}

type Strallyint struct {
	Value  int
	Nodeid string
}

func CalcAllValueFunction(nodelist []vm.DepositDetail) []Stf { //nodelist []Mynode) map[string]float32 {
	//	CapitalMap := make(map[string]float64)
	//	CapitalMap := make(map[string]float32)
	var CapitalMap []Stf

	for _, item := range nodelist {
		self := SelfNodeInfo{nodeid: item.NodeID.String(), stk: float64(item.Deposit.Uint64()), uptime: int(item.OnlineTime.Uint64()), tps: 1000, Coef_tps: 0.2, Coef_stk: 0.25}
		value := self.Last_Time() * (self.TPS_POWER()*self.Coef_tps + self.deposit_stake()*self.Coef_stk)
		//		//		CapitalMap[self.nodeid] = float32(value)
		CapitalMap = append(CapitalMap, Stf{Str: self.nodeid, Flot: float64(value)})
	}
	return CapitalMap
}

func Normalize(probVal []Stf) []pnormalized {

	//fmt.Println(probVal)
	var total float64
	for _, item := range probVal {
		total += item.Flot
	}
	var pnormalizedlist []pnormalized
	for _, item := range probVal {
		var tmp pnormalized
		tmp.Value = item.Flot / total
		tmp.Nodeid = item.Str
		pnormalizedlist = append(pnormalizedlist, tmp)
		//		fmt.Println("There are", views, "views for", key)
	}
	return pnormalizedlist
}

func Sample1NodesInValNodes(probnormalized []pnormalized, rand01 float64) string {

	for _, iterm := range probnormalized {
		rand01 -= iterm.Value
		if rand01 < 0 {
			return iterm.Nodeid
		}
	}
	return probnormalized[0].Nodeid
}

func CommbineFundNodesAndPricipal(probVal []Stf, probFund []Stf, PricipalValNodes []Strallyint, ratiodnlimit float64, ratiouplimit float64) []Strallyint {

	if (J == 0 || len(probFund) == 0) && len(PricipalValNodes) == 0 {
		var empty []Strallyint
		return empty
	}

	if J == 0 || len(probFund) == 0 {
		return PricipalValNodes
	}
	if len(PricipalValNodes) == 0 {
		var probFundnormalized []Strallyint
		temp := Normalize(probFund)
		for _, item := range temp {
			var elem *Strallyint
			elem = new(Strallyint)
			elem.Nodeid = item.Nodeid
			elem.Value = int(item.Value * 100)
			probFundnormalized = append(probFundnormalized, *elem)
		}
		return probFundnormalized
	}

	var PricipalVoteSum int
	var probPricipalSum float64
	var probFundSum float64
	var ratio float64

	for _, item := range PricipalValNodes {
		//		probFundnormalized[index].Value *= 100
		PricipalVoteSum += item.Value
	}

	//	var pricipalkeys []string
	var probValMap map[string]float64
	probValMap = make(map[string]float64)
	for _, item := range probVal {
		probValMap[item.Str] = item.Flot
	}

	//根据主验证节点找到对应的价值 计算选举出的主验证价值和
	for _, item := range PricipalValNodes {
		v := probValMap[item.Nodeid]
		probPricipalSum += v
	}

	//计算基金会节点价值和
	for _, item := range probFund {
		probFundSum += item.Flot
	}

	//计算比率
	ratio = updownlimit(probFundSum/probPricipalSum, 2.5, 4.0)

	var FundValNodes []Strallyint
	var temp *Strallyint
	temp = new(Strallyint)
	for _, item := range probFund {
		temp.Nodeid = item.Str
		//基金会节点价值 / 基金会总价值 * 比率 * 竞选到的主验证节点的投票总数
		temp.Value = int(item.Flot / probFundSum * ratio * float64(PricipalVoteSum))
		FundValNodes = append(FundValNodes, *temp)
	}

	PricipalValNodes = append(PricipalValNodes, FundValNodes...)
	//	FundValNodes = append(FundValNodes, PricipalValNodes...)
	return PricipalValNodes
}
func updownlimit(a float64, ratiouplimit float64, ratiodnlimit float64) float64 {
	if a < ratiodnlimit {
		a = ratiodnlimit
	}
	if a > ratiouplimit {
		a = ratiouplimit
	}
	return a
}

func SampleMinerNodes(probnormalized []pnormalized, seed int64, Ms int) ([]Strallyint, []Strallyint) {

	var PricipalMinerNodes []Strallyint
	var BakMinerNodes []Strallyint
	/*
		sort := func(probnormalized []pnormalized, PricipalMinerNodes []Strallyint, BakMinerNodes []Strallyint) ([]Strallyint, []Strallyint) {
			Pricipal := make(map[string]int)
			BakMin := make(map[string]int)

			var RPricipalMinerNodes []Strallyint
			var RBakMinerNodes []Strallyint

			for _, item := range PricipalMinerNodes {
				Pricipal[item.Nodeid] = item.Value
			}
			for _, item := range BakMinerNodes {
				BakMin[item.Nodeid] = item.Value
			}
			for _, item := range probnormalized {
				var ok bool
				_, ok = Pricipal[item.Nodeid]
				if ok == true {
					RPricipalMinerNodes = append(RPricipalMinerNodes, Strallyint{Nodeid: item.Nodeid, Value: Pricipal[item.Nodeid]})
					continue
				}
				_, ok = BakMin[item.Nodeid]
				if ok == true {
					RBakMinerNodes = append(RBakMinerNodes, Strallyint{Nodeid: item.Nodeid, Value: BakMin[item.Nodeid]})
				}
			}
			return RPricipalMinerNodes, RBakMinerNodes
		}
	*/
	// 如果当选节点不到N个,其他列表为空
	dict := make(map[string]int)
	//Ele.N = Ms
	if len(probnormalized) <= N { //加判断 定义为func
		for _, item := range probnormalized {
			//			probnormalized[index].value = 100 * iterm.value
			temp := Strallyint{Value: int(100 * item.Value), Nodeid: item.Nodeid}
			PricipalMinerNodes = append(PricipalMinerNodes, temp)
		}
		//		return [(e[0],int(100*e[1])) for e in probnormalized],[],[]
		return PricipalMinerNodes, BakMinerNodes
		//return sort(probnormalized, PricipalMinerNodes, BakMinerNodes)
	}

	// 如果当选节点超过N,最多连续进行1000次采样或者选出N个节点
	rand := mt19937.RandUniformInit(seed)
	for i := 0; i < MaxSample; i++ {
		node := Sample1NodesInValNodes(probnormalized, float64(rand.Uniform(0.0, 1.0)))
		_, ok := dict[node]
		if ok == true {
			dict[node] = dict[node] + 1
		} else {
			dict[node] = 1
		}
		if len(dict) == N {
			break
		}
	}

	// 如果没有选够N个
	for _, item := range probnormalized {
		vint, ok := dict[item.Nodeid]

		if ok == true {
			var tmp Strallyint
			tmp.Nodeid = item.Nodeid
			tmp.Value = vint
			PricipalMinerNodes = append(PricipalMinerNodes, tmp)
		} else {
			BakMinerNodes = append(BakMinerNodes, Strallyint{Value: int(item.Value), Nodeid: item.Nodeid})
		}
	}
	lenPM := len(PricipalMinerNodes)
	if N > lenPM {
		PricipalMinerNodes = append(PricipalMinerNodes, BakMinerNodes[:N-lenPM]...)
		BakMinerNodes = BakMinerNodes[N-lenPM:]
	}
	return PricipalMinerNodes, BakMinerNodes
	///return sort(probnormalized, PricipalMinerNodes, BakMinerNodes)
}

func CalcRemainingNodesVotes(RemainingProbNormalizedNodes []Strallyint) []Strallyint {
	for index, _ := range RemainingProbNormalizedNodes {
		RemainingProbNormalizedNodes[index].Value = 1
	}
	return RemainingProbNormalizedNodes
}

//做异常判断
func SampleMPlusPNodes(probnormalized []pnormalized, seed int64, M int, J int) ([]Strallyint, []Strallyint, []Strallyint) {
	var PricipalValNodes []Strallyint
	var RemainingProbNormalizedNodes []Strallyint //[]pnormalized
	var BakValNodes []Strallyint

	// 如果当选节点不到M-J个(加上基金会节点不足M个),则全部当选,其他列表为空
	dict := make(map[string]int)
	if len(probnormalized) <= M-J { //加判断 定义为func
		for _, item := range probnormalized {
			temp := Strallyint{Value: int(100 * item.Value), Nodeid: item.Nodeid}
			PricipalValNodes = append(PricipalValNodes, temp)
		}
		//		return sortfunc(probnormalized, PricipalValNodes, BakValNodes, RemainingProbNormalizedNodes)
		return PricipalValNodes, BakValNodes, RemainingProbNormalizedNodes
	}

	// 如果当选节点超过M-J,最多连续进行1000次采样或者选出M+P-J个节点
	rand := mt19937.RandUniformInit(seed)

	for i := 0; i < MaxSample; i++ {
		node := Sample1NodesInValNodes(probnormalized, float64(rand.Uniform(0.0, 1.0)))

		_, ok := dict[node]
		if ok == true {
			dict[node] = dict[node] + 1
		} else {
			dict[node] = 1
		}

		if len(dict) == (M - J) {
			break
		}
	}
	for _, item := range probnormalized {
		_, ok := dict[item.Nodeid]
		if ok == false {
			RemainingProbNormalizedNodes = append(RemainingProbNormalizedNodes, Strallyint{Nodeid: item.Nodeid, Value: dict[item.Nodeid]})
		} else {
			PricipalValNodes = append(PricipalValNodes, Strallyint{Nodeid: item.Nodeid, Value: dict[item.Nodeid]})
		}
	}
	return PricipalValNodes, BakValNodes, RemainingProbNormalizedNodes
}

type SelfNodeInfo struct {
	nodeid   string
	stk      float64
	uptime   int
	tps      int
	Coef_tps float64
	Coef_stk float64
}

func (self *SelfNodeInfo) TPS_POWER() float64 {
	tps_weight := 1.0
	if self.tps >= 16000 {
		tps_weight = 5.0
	} else if self.tps >= 8000 {
		tps_weight = 4.0
	} else if self.tps >= 4000 {
		tps_weight = 3.0
	} else if self.tps >= 2000 {
		tps_weight = 2.0
	} else if self.tps >= 1000 {
		tps_weight = 1.0
	} else {
		tps_weight = 0.0
	}
	return tps_weight
}

func (self *SelfNodeInfo) Last_Time() float64 {
	CandidateTime_weight := 4.0
	if self.uptime <= 64 {
		CandidateTime_weight = 0.25
	} else if self.uptime <= 128 {
		CandidateTime_weight = 0.5
	} else if self.uptime <= 256 {
		CandidateTime_weight = 1
	} else if self.uptime <= 512 {
		CandidateTime_weight = 2
	} else {
		CandidateTime_weight = 4
	}
	return CandidateTime_weight
}

func (self *SelfNodeInfo) deposit_stake() float64 {
	stake_weight := 1.0
	if self.stk >= 40000 {
		stake_weight = 4.5
	} else if self.stk >= 20000 {
		stake_weight = 2.15
	} else if self.stk >= 10000 {
		stake_weight = 1.0
	} else {
		stake_weight = 0.0
	}
	return stake_weight
}
