// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package support

import (
	"fmt"
	"math"
	"math/big"
	"math/rand"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/common/mt19937"
	"github.com/MatrixAINetwork/go-matrix/core/vm"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
)

const (
	DefaultMinerStock = 1
)

type RatioList struct {
	MinNum uint64
	Ratio  float64
}
type Pnormalized struct {
	Value float64
	Addr  common.Address
}
type AllNative struct {
	Master    []mc.ElectNodeInfo //验证者主节点
	BackUp    []mc.ElectNodeInfo //验证者备份
	Candidate []mc.ElectNodeInfo //验证者候选

	MasterQ    []common.Address //第一梯队候选
	BackUpQ    []common.Address //第二梯队候选
	CandidateQ []common.Address //第三梯队候选
	ElectInfo  *mc.ElectConfigInfo
}

type Strallyint struct {
	Value    int
	Addr     common.Address
	VIPLevel common.VIPRoleType
}

type Node struct {
	Address     common.Address
	SignAddress common.Address
	Deposit     *big.Int
	WithdrawH   *big.Int
	OnlineTime  *big.Int
	Ratio       uint16
	vipLevel    common.VIPRoleType
	index       int
	Usable      bool
}

type Electoion struct {
	SeqNum      uint64
	RandSeed    *mt19937.RandUniform
	VipLevelCfg []mc.VIPConfig
	NodeList    []Node

	EleCfg mc.ElectConfigInfo_All

	ChosedNum                  int
	NeedNum                    int
	HasChosedNode              [][]Strallyint
	MapMoney                   map[common.Address]uint64
	BlockProduceSlashBlackList mc.BlockProduceSlashBlackList
	BlockBlackProc             *BlockProduceProc
}

type ElectDP struct {
	RandSeed    *mt19937.RandUniform
	EleCfg      *mc.ElectConfigInfo_All
	ChosedNum   uint64
	DepositNode []vm.DepositDetail
	DepostiA0   map[common.Address]bool
	DepostiA1   map[common.Address]common.Address
	BpBlackList *mc.BasePowerSlashBlackList
	Num         uint64
	UpdateSeq   bool
}

func getAccountListInDeposit(minerList []vm.DepositDetail, candidateList []common.Address) []common.Address {
	newCandidateList := make([]common.Address, 0, 1024)
	for i := 0; i < len(minerList); i++ {
		if FindAddress(minerList[i].Address, candidateList) {
			newCandidateList = append(newCandidateList, minerList[i].Address)
		} else {
			//log.Trace("动态选举方案", "节点不存在", minerList[i].Address)
		}

	}
	return newCandidateList
}

func (self *ElectDP) getAccountListInDepositMap(candidateList []common.Address) []common.Address {
	newCandidateList := make([]common.Address, 0, 1024)
	for _, v := range candidateList {
		if _, ok := self.DepostiA0[v]; ok {
			newCandidateList = append(newCandidateList, v)
		}
	}
	return newCandidateList
}

func (self *ElectDP) GetUsableNodeList(candidateList, excludeNode []common.Address) []common.Address {
	var newCandidateList []common.Address
	if len(candidateList) == 0 {
		for _, v := range self.DepositNode {
			newCandidateList = append(newCandidateList, v.Address)
		}
	} else {
		newCandidateList = self.getAccountListInDepositMap(candidateList)
	}

	canUseNodeList := self.NewFilterWhiteBlackList(newCandidateList, excludeNode)
	log.Info("动态选举方案", "黑白名单处理输入节点个数", len(candidateList), "输出节点个数", len(canUseNodeList))
	return canUseNodeList
}
func (self *ElectDP) NewFilterWhiteBlackList(CandidateList, excludeNode []common.Address) []common.Address {
	nodeUserMap := make(map[common.Address]bool, 1024)

	for _, v := range CandidateList {
		nodeUserMap[v] = true
	}
	//白名单
	if self.EleCfg.WhiteListSwitcher {
		for _, v := range CandidateList {
			if !FindAddress(v, self.EleCfg.WhiteList) {
				nodeUserMap[v] = false
			}
		}
	}
	//黑名单A0
	for _, v := range self.EleCfg.BlackList {
		nodeUserMap[v] = false
	}

	//黑名单,过滤A1
	for _, v := range self.EleCfg.BlackList {
		if A0, ok := self.DepostiA1[v]; ok {
			nodeUserMap[A0] = false
		}
	}

	//extra
	for _, v := range excludeNode {
		nodeUserMap[v] = false
	}
	//算力黑名单
	for _, v := range self.BpBlackList.BlackList {
		if v.ProhibitCycleCounter > 0 {
			nodeUserMap[v.Address] = false
		}
	}
	//转换为节点列表
	canUseNodeList := make([]common.Address, 0)
	for _, v := range CandidateList {
		if nodeUserMap[v] {
			canUseNodeList = append(canUseNodeList, v)
		}
	}
	return canUseNodeList
}

func (self *ElectDP) OldFilterWhiteBlackList(CandidateList, excludeNode []common.Address) []common.Address {
	nodeUserMap := make(map[common.Address]bool, 1024)

	for _, v := range CandidateList {
		nodeUserMap[v] = true
	}
	//白名单
	if self.EleCfg.WhiteListSwitcher {
		for _, v := range CandidateList {
			if !FindAddress(v, self.EleCfg.WhiteList) {
				nodeUserMap[v] = false
			}
		}
	}
	//黑名单A0
	for _, v := range CandidateList {
		if FindAddress(v, self.EleCfg.BlackList) {
			nodeUserMap[v] = false
		}
	}
	//黑名单,过滤A1
	for _, v := range self.DepositNode {
		if FindAddress(v.SignAddress, self.EleCfg.BlackList) {
			nodeUserMap[v.Address] = false
			continue
		}
	}
	//extra
	for _, v := range CandidateList {
		if FindAddress(v, excludeNode) {
			nodeUserMap[v] = false
		}
	}
	//算力黑名单
	for _, v := range self.BpBlackList.BlackList {
		if v.ProhibitCycleCounter > 0 && FindAddress(v.Address, CandidateList) {
			nodeUserMap[v.Address] = false
			continue
		}
	}
	//转换为节点列表
	canUseNodeList := make([]common.Address, 0)
	for _, v := range CandidateList {
		if nodeUserMap[v] {
			canUseNodeList = append(canUseNodeList, v)
		}
	}
	return canUseNodeList
}

func (self *ElectDP) AddChosedNum(hasChosedNum uint64) {
	self.ChosedNum = self.ChosedNum + hasChosedNum
}

//todo:按照轮次减
func (self *ElectDP) DecrementBpSlashCount() {
	if self.UpdateSeq == false {
		return
	}
	newBlackList := make([]mc.BasePowerSlash, 0)
	for i := 0; i < len(self.BpBlackList.BlackList); i++ {
		if self.BpBlackList.BlackList[i].ProhibitCycleCounter > 0 {
			self.BpBlackList.BlackList[i].ProhibitCycleCounter = self.BpBlackList.BlackList[i].ProhibitCycleCounter - 1
			newBlackList = append(newBlackList, self.BpBlackList.BlackList[i])
		}
	}
	self.BpBlackList.BlackList = newBlackList
}

func (node *Node) SetUsable(status bool) {

	node.Usable = status
}

func (node *Node) SetIndex(index int) {
	node.index = index
}
func (node *Node) SetVipLevelInfo(VipLevelCfg []mc.VIPConfig) uint64 {
	temp := big.NewInt(0).Set(node.Deposit)
	deposMan := temp.Div(temp, common.ManValue).Uint64()

	for index := len(VipLevelCfg) - 1; index >= 0; index-- {
		if deposMan >= VipLevelCfg[index].MinMoney {
			node.vipLevel = common.GetVIPLevel(index)
			node.Ratio = VipLevelCfg[index].StockScale
			return deposMan
		}
	}
	node.Ratio = DefaultRatio
	node.vipLevel = common.VIP_Nil
	return deposMan
}

func (node *Node) SetDepositInfo(depsit vm.DepositDetail) {
	node.Address = depsit.Address
	node.SignAddress = depsit.SignAddress
	node.OnlineTime = depsit.OnlineTime
	node.WithdrawH = depsit.WithdrawH
	node.Deposit = depsit.Deposit
	//todo:地址为空地址 ，WithdrawH，OnlineTime负值，抵押负值
	if nil == depsit.Deposit {
		node.Deposit = big.NewInt(DefaultNodeConfig)
	}
	if nil == depsit.WithdrawH {
		node.WithdrawH = big.NewInt(DefaultNodeConfig)
	}
	if nil == depsit.OnlineTime {
		node.OnlineTime = big.NewInt(DefaultNodeConfig)
	}
}

func NewElelection(VipLevelCfg []mc.VIPConfig, vm []vm.DepositDetail, EleCfg mc.ElectConfigInfo_All, randseed *big.Int, seqNum uint64, types common.RoleType) *Electoion {
	var vip Electoion
	vip.SeqNum = seqNum
	vip.RandSeed = mt19937.RandUniformInit(randseed.Int64())
	vip.EleCfg = EleCfg

	vip.VipLevelCfg = VipLevelCfg
	vip.ChosedNum = 0
	switch types {
	case common.RoleValidator:
		vip.NeedNum = int(EleCfg.BackValidator + EleCfg.ValidatorNum)
	default:
		vip.NeedNum = int(EleCfg.MinerNum)
	}
	vip.MapMoney = make(map[common.Address]uint64)

	for i := 0; i < len(vm); i++ {
		vip.NodeList = append(vip.NodeList, Node{})
	}
	for i := 0; i < len(vm); i++ {
		vip.NodeList[i].SetDepositInfo(vm[i])
		manValue := vip.NodeList[i].SetVipLevelInfo(VipLevelCfg)
		vip.NodeList[i].SetIndex(i)
		vip.NodeList[i].SetUsable(true)

		vip.MapMoney[vip.NodeList[i].Address] = manValue

	}
	return &vip
}
func NewMEPElection(VipLevelCfg []mc.VIPConfig, vm []vm.DepositDetail, EleCfg mc.ElectConfigInfo_All, randseed *big.Int, seqNum uint64, types common.RoleType) *Electoion {
	var vip Electoion
	vip.SeqNum = seqNum
	vip.RandSeed = mt19937.RandUniformInit(randseed.Int64())
	vip.EleCfg = EleCfg
	vip.VipLevelCfg = VipLevelCfg
	vip.ChosedNum = 0
	switch types {
	case common.RoleValidator:
		vip.NeedNum = int(EleCfg.BackValidator + EleCfg.ValidatorNum)
	default:
		vip.NeedNum = int(EleCfg.MinerNum)
	}
	vip.MapMoney = make(map[common.Address]uint64)
	for i := 0; i < len(vm); i++ {
		vip.NodeList = append(vip.NodeList, Node{})
	}
	for i := 0; i < len(vm); i++ {
		vip.NodeList[i].SetDepositInfo(vm[i])
		vip.NodeList[i].SetIndex(i)
		vip.NodeList[i].SetUsable(true)
		vip.MapMoney[vip.NodeList[i].Address] = DefaultMinerDeposit
	}
	return &vip
}
func NewDpElection(mmrerm *mc.MasterMinerReElectionReqMsg, bpBlackList *mc.BasePowerSlashBlackList) *ElectDP {
	var edp ElectDP
	edp.RandSeed = mt19937.RandUniformInit(mmrerm.RandSeed.Int64())
	edp.EleCfg = &mmrerm.ElectConfig
	edp.ChosedNum = 0
	edp.DepositNode = mmrerm.MinerList
	edp.DepostiA0 = make(map[common.Address]bool)
	edp.DepostiA1 = make(map[common.Address]common.Address)
	for _, v := range edp.DepositNode {
		edp.DepostiA0[v.Address] = false
		edp.DepostiA1[v.SignAddress] = v.Address
	}
	edp.BpBlackList = bpBlackList
	edp.Num = mmrerm.SeqNum
	edp.UpdateSeq = false
	return &edp
}
func (vip *Electoion) GetAvailableNodeNum() int {
	var availableNodeNum = 0

	for i := 0; i < len(vip.NodeList); i++ {
		if vip.NodeList[i].Usable {
			availableNodeNum++
		}
	}
	return availableNodeNum
}
func FindAddress(addr common.Address, addrList []common.Address) bool {
	for _, v := range addrList {
		if v.Equal(addr) == true {
			return true
		}
	}
	return false
}
func (vip *Electoion) DisPlayNode() {
	//for _,v:=range vip.NodeList{
	//	fmt.Println(v.Address.String(),v.Deposit.String(),vip.MapMoney[v.Address])
	//}
	for _, v := range vip.NodeList {
		fmt.Println(v.Address.String(), v.Deposit, v.WithdrawH, v.OnlineTime, v.vipLevel, v.index, "Ratio", v.Ratio, v.Usable)
	}
}
func (vip *Electoion) SetChosed(node []Strallyint) {

	ChoseNode := []common.Address{}
	for _, v := range node {
		ChoseNode = append(ChoseNode, v.Addr)
	}
	for k, v := range vip.NodeList {
		if FindAddress(v.Address, ChoseNode) {
			vip.NodeList[k].SetUsable(false)
		}
	}
	vip.ChosedNum += len(node)
	vip.HasChosedNode = append(vip.HasChosedNode, node)
}
func (vip *Electoion) ProcessBlackNode() {
	for k, v := range vip.NodeList {
		if FindAddress(v.Address, vip.EleCfg.BlackList) {
			vip.NodeList[k].SetUsable(false)
			continue
		}
		if FindAddress(v.SignAddress, vip.EleCfg.BlackList) {
			vip.NodeList[k].SetUsable(false)
		}
	}
}

func (vip *Electoion) GetMinSuperNodeAmount() (bool, uint64) {
	L := len(vip.HasChosedNode[0])

	if L == 0 {
		return false, 1
	}

	minv, _ := vip.MapMoney[vip.HasChosedNode[0][0].Addr]
	for i := 1; i < L; i++ {
		amount, _ := vip.MapMoney[vip.HasChosedNode[0][i].Addr]
		if amount < minv {
			minv = amount
		}
	}
	return true, minv
}

func stockProtect(stock int) int {
	if stock == 0 {
		stock = 1
	}
	if stock > 0xFFFF {
		stock = 0xFFFF
	}
	return stock
}
func (vip *Electoion) SuperNodeStockProc(randSuperNodeStock map[common.Address]int, stockExp float64) {
	//no supernode , ignore
	if len(vip.HasChosedNode[0]) == 0 {
		return
	}

	//no rand sample case : All mortgages are the same. fix stock 1000;
	if len(vip.HasChosedNode[1]) == 0 {
		_, minSuperNodeAmount := vip.GetMinSuperNodeAmount()
		for i := 0; i < len(vip.HasChosedNode[0]); i++ {
			factor := math.Pow(float64(vip.MapMoney[vip.HasChosedNode[0][i].Addr])/float64(minSuperNodeAmount), stockExp)
			stock := int(factor*100 + 0.5)
			stock = stockProtect(stock)
			vip.HasChosedNode[0][i].Value = stock
		}
	} else {
		for i := 0; i < len(vip.HasChosedNode[0]); i++ {
			address := vip.HasChosedNode[0][i].Addr
			stock := 1
			if value, ok := randSuperNodeStock[address]; ok {
				stock = value
			}
			stock = stockProtect(stock)
			vip.HasChosedNode[0][i].Value = stock
		}
	}
}

func (vip *Electoion) GetVipStock(addr common.Address) int {
	stockSum := int(0)
	stockDespoit := uint64(0)
	for k, v := range vip.HasChosedNode {
		if k != len(vip.HasChosedNode)-1 {
			continue
		}
		for _, vv := range v {
			stockSum += vv.Value
			stockDespoit += vip.MapMoney[vv.Addr]
		}
	}
	ratio := int(0.0)
	if stockDespoit == 0 {
		ratio = int(vip.MapMoney[addr] / vip.VipLevelCfg[1].MinMoney)
	} else {
		ratio = int(float64(stockSum)/float64(stockDespoit)*float64(vip.MapMoney[addr]) + 0.5)
	}
	if ratio > 0xffff {
		return 0xffff
	}
	if ratio == 0 {
		return 1
	}

	return ratio

}

type BlockProduceProc struct {
	addressMap map[common.Address]int
	List       []mc.UserBlockProduceSlash
}

func NewBlockProduceProc(blackList mc.BlockProduceSlashBlackList) *BlockProduceProc {
	s := &BlockProduceProc{make(map[common.Address]int), make([]mc.UserBlockProduceSlash, 0, 0)}
	if blackList.BlackList != nil {
		for k, v := range blackList.BlackList {
			s.addressMap[v.Address] = k
			s.List = append(s.List, v)
		}
	}
	return s
}
func (s *BlockProduceProc) IsBlackList(address common.Address) (int, bool) {
	k, ok := s.addressMap[address]
	return k, ok
}

func (s *BlockProduceProc) DecrementCount(address common.Address) {
	if k, ok := s.addressMap[address]; ok {
		if s.List[k].ProhibitCycleCounter > 0 {
			s.List[k].ProhibitCycleCounter = s.List[k].ProhibitCycleCounter - 1
		}
	}
}

func (vip *Electoion) ProcessWhiteNode() {
	for k, v := range vip.NodeList {
		if !FindAddress(v.Address, vip.EleCfg.WhiteList) {
			vip.NodeList[k].SetUsable(false)
		}
	}
}
func (vip *Electoion) GetNodeByAccount(address common.Address) (int, bool) {
	for k, v := range vip.NodeList {
		if v.Address.Equal(address) {
			return k, true
		}
	}
	return 0, false
}
func (vip *Electoion) GetNodeByLevel(level common.VIPRoleType) []Node {
	specialNode := make([]Node, 0)
	for i := 0; i < len(vip.NodeList); i++ {
		if vip.NodeList[i].Usable == false {
			continue
		}
		if vip.NodeList[i].vipLevel >= level {
			specialNode = append(specialNode, vip.NodeList[i])
		}
	}
	return specialNode
}
func (vip *Electoion) FilterBlockSlashList() {
	for i := 0; i < len(vip.NodeList); i++ {
		if _, ok := vip.BlockBlackProc.IsBlackList(vip.NodeList[i].Address); ok {
			vip.NodeList[i].Usable = false
		}
	}
}
func (vip *Electoion) GetUsableNode() []Node {
	usableNodeList := make([]Node, 0)
	for i := 0; i < len(vip.NodeList); i++ {
		if vip.NodeList[i].Usable == false {
			continue
		}
		usableNodeList = append(usableNodeList, vip.NodeList[i])
	}
	return usableNodeList
}
func superNodePrePorc(nodeList []Node, superThreshold int64) []Node {
	superNode := make([]Node, 0)

	if nodeList == nil {
		return superNode
	}
	acc := big.NewInt(0)
	for _, v := range nodeList {
		if v.Usable {
			acc.Add(acc, v.Deposit)
		}
	}

	fac := big.NewInt(superThreshold)
	for _, v := range nodeList {
		if v.Usable == false {
			continue
		}

		if big.NewInt(0).Mul(v.Deposit, fac).Cmp(acc) >= 0 {
			superNode = append(superNode, v)
		}
	}
	return superNode
}
func (vip *Electoion) GenSuperNode(superThreshold int64) ([]Strallyint, []Node) {
	vipNodeMap := make(map[common.Address]int)

	for k, v := range vip.NodeList {
		vipNodeMap[v.Address] = k
	}
	//pre get super node
	preSuperNode := superNodePrePorc(vip.NodeList, superThreshold)

	//set all nodes usable
	for _, v := range preSuperNode {
		if index, ok := vipNodeMap[v.Address]; ok {
			vip.NodeList[index].SetUsable(false)
		} else {
			log.Error("Election Module", "Pre SuperNode Invalid", v.Address.String())
		}
	}

	//Exclude BlockSlashBlackList From pre-superNode
	superNodeS := make([]Strallyint, 0, len(preSuperNode))
	superNodeN := make([]Node, 0, len(preSuperNode))
	for _, v := range preSuperNode {
		if _, ok := vip.BlockBlackProc.IsBlackList(v.Address); ok {
			vip.BlockBlackProc.DecrementCount(v.Address)
			log.Trace("Layered_BSS", "SuperNode", v.Address.String(), "Elect Slash", true)
		} else {
			log.Trace("Layered_BSS", "SuperNode", v.Address.String(), "Elect Slash", false)
			superNodeS = append(superNodeS, Strallyint{Value: 1, Addr: v.Address, VIPLevel: common.VIP_1})
			superNodeN = append(superNodeN, v)
		}
	}
	return superNodeS, superNodeN
}

func (vip *Electoion) GetNodeIndexByLevel(level common.VIPRoleType) []int {
	specialNode := make([]int, 0)
	for i := 0; i < len(vip.NodeList); i++ {
		if level == vip.NodeList[i].vipLevel {
			specialNode = append(specialNode, i)
		}
	}
	return specialNode
}

func (vip *Electoion) GetLastNode() []Node {
	var remainNodeList = make([]Node, 0)
	for i := 0; i < len(vip.NodeList); i++ {
		if vip.NodeList[i].Usable == false {
			continue
		}
		remainNodeList = append(remainNodeList, vip.NodeList[i])
	}
	return remainNodeList
}

func (vip *Electoion) GetWeight(role common.RoleType) []Pnormalized {
	lastnode := vip.GetLastNode()
	return CalcValue(lastnode, role)
}

func Knuth_Fisher_Yates_Algorithm(nodeList []Node, randSeed *big.Int) []Node {
	//高纳德置乱算法
	rand.Seed(randSeed.Int64())
	for index := len(nodeList) - 1; index > 0; index-- {
		aimIndex := rand.Intn(index + 1)
		t := nodeList[index]
		nodeList[index] = nodeList[aimIndex]
		nodeList[aimIndex] = t
	}
	return nodeList
}
func (vip *Electoion) GetIndex(addr common.Address) (int, bool) {
	for k, v := range vip.NodeList {
		if v.Address.Equal(addr) {
			return k, true
		}
	}
	return 0, false
}

func (vip *Electoion) SetBlockBlackList(list mc.BlockProduceSlashBlackList) {
	vip.BlockBlackProc = NewBlockProduceProc(list)
}

type SortNodeList []Node

func (self SortNodeList) Len() int {
	return len(self)
}
func (self SortNodeList) Less(i, j int) bool {
	if self[i].Deposit.Cmp(self[j].Deposit) == 0 {
		return self[i].OnlineTime.Cmp(self[j].OnlineTime) > 0
	}
	return self[i].Deposit.Cmp(self[j].Deposit) > 0
}
func (self SortNodeList) Swap(i, j int) {
	temp := self[i]
	self[i] = self[j]
	self[j] = temp
}
