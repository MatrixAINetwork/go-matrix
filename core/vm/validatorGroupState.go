// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package vm

import (
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/vm/validatorGroup"
	"github.com/MatrixAINetwork/go-matrix/params"
	"github.com/MatrixAINetwork/go-matrix/rlp"
	"math"
	"math/big"
)

type RateOption struct {
	Rate    *big.Int
	Decimal *big.Int
}

func (ro *RateOption) Mul(a *big.Int) *big.Int {
	b := new(big.Int).Mul(a, ro.Rate)
	b.Div(b, ro.Decimal)
	return b
}

type DepositInfo struct {
	Address common.Address
	Amount  *big.Int
}
type RateInfo struct {
	Threshold *big.Int
	Rate      RateOption
}

//Rewards proportion to participants
type RewardRate struct {
	OwnerRate RateOption
	NodeRate RateOption
	//	CurrentRate RateOption
	LevelRate []RateInfo
}

type OwnerInfo struct {
	Owner           common.Address
	WithdrawAllTime uint64
	SignAddress     common.Address `rlp:"-"`
}

func NewValidatorGroupState() *ValidatorGroupState {
	return &ValidatorGroupState{
		depInfo: &MatrixDeposit002{},
	}
}

type ValidatorGroupState struct {
	depInfo      DepositInterface
	OwnerInfo    OwnerInfo
	Reward       RewardRate
	ValidatorMap validatorGroup.ValidatorInfoSlice
}

func (vc *ValidatorGroupState) SetRewardRate(OwnerRate,nodeRate *big.Int, lvlRate []*big.Int) error {
	if OwnerRate.Sign() < 0 {
		return errArguments
	}
	vc.Reward.OwnerRate = RateOption{OwnerRate, rateDecmalBig}
	if len(lvlRate) != len(depLvel) {
		return errArguments
	}
	if nodeRate.Sign()<0 || nodeRate.Cmp(rateDecmalBig)>0{
		return errArguments
	}
	vc.Reward.NodeRate = RateOption{nodeRate, rateDecmalBig}
	vc.Reward.LevelRate = make([]RateInfo, len(lvlRate))
	for i := 0; i < len(lvlRate); i++ {
		if lvlRate[i].Sign() < 0 {
			return errArguments
		}
		vc.Reward.LevelRate[i].Threshold = depLvel[i]
		vc.Reward.LevelRate[i].Rate = RateOption{lvlRate[i], rateDecmalBig}
	}
	return nil
}

/*
func (vc *ValidatorGroupState)CalCurrentWithdrawAmount(amount *big.Int)*big.Int{
	return vc.Reward.CurrentRate.Mul(amount)
}
*/
func (vc *ValidatorGroupState) SetOwner(contractAddress common.Address, state StateDBManager) error {
	data, err := rlp.EncodeToBytes(vc.OwnerInfo)
	if err != nil {
		return err
	}
	state.SetStateByteArray(params.MAN_COIN, contractAddress, EmergeKey(contractAddress, "Owner"), data)
	return nil
}
func (vc *ValidatorGroupState) GetOwner(contractAddress common.Address, state StateDBManager) error {
	data := state.GetStateByteArray(params.MAN_COIN, contractAddress, EmergeKey(contractAddress, "Owner"))
	if len(data) == 0 {
		return nil
	}
	err := rlp.DecodeBytes(data, &vc.OwnerInfo)
	return err
}
func (vc *ValidatorGroupState) SetReward(contractAddress common.Address, state StateDBManager) error {
	data, err := rlp.EncodeToBytes(vc.Reward)
	if err != nil {
		return err
	}
	state.SetStateByteArray(params.MAN_COIN, contractAddress, EmergeKey(contractAddress, "Reward"), data)
	return nil
}
func (vc *ValidatorGroupState) GetReward(contractAddress common.Address, state StateDBManager) error {
	data := state.GetStateByteArray(params.MAN_COIN, contractAddress, EmergeKey(contractAddress, "Reward"))
	if len(data) == 0 {
		return nil
	}
	err := rlp.DecodeBytes(data, &vc.Reward)
	return err
}
func (vc *ValidatorGroupState) SetValidatorMap(contractAddress common.Address, state StateDBManager) error {
	data, err := rlp.EncodeToBytes(vc.ValidatorMap)
	if err != nil {
		return err
	}
	state.SetStateByteArray(params.MAN_COIN, contractAddress, EmergeKey(contractAddress, "ValiMap"), data)
	return nil
}
func (vc *ValidatorGroupState) GetAllDepositInfo(contractAddress common.Address, state StateDBManager) *common.DepositBase {
	newCon := NewContract(AccountRef(contractAddress), AccountRef(developContractAddress), big.NewInt(0), uint64(0), params.MAN_COIN)
	return vc.depInfo.GetDepositBase(newCon, state, contractAddress)

}
func (vc *ValidatorGroupState) GetValidatorMap(contractAddress common.Address, time uint64, state StateDBManager) error {
	data := state.GetStateByteArray(params.MAN_COIN, contractAddress, EmergeKey(contractAddress, "ValiMap"))
	if len(data) == 0 {
		return nil
	}
	err := rlp.DecodeBytes(data, &vc.ValidatorMap)
	if err != nil {
		return err
	}
	allDepInfo := vc.GetAllDepositInfo(contractAddress, state)
	reFundTime := uint64(math.MaxUint64)
	depMap := make(map[uint64]*common.DepositMsg)
	if allDepInfo != nil {
		vc.OwnerInfo.SignAddress = allDepInfo.AddressA1
		for i := 0; i < len(allDepInfo.Dpstmsg); i++ {
			msg := &allDepInfo.Dpstmsg[i]
			depMap[msg.Position] = msg
		}
		if len(allDepInfo.Dpstmsg) > 0 && len(allDepInfo.Dpstmsg[0].WithDrawInfolist) > 0 {
			reFundTime = allDepInfo.Dpstmsg[0].WithDrawInfolist[0].WithDrawTime
		}
	}
	for i := 0; i < len(vc.ValidatorMap); i++ {
		valiInfo := &vc.ValidatorMap[i]
		valiInfo.AllAmount = new(big.Int).Set(valiInfo.Current.Amount)
		for j := 0; j < len(valiInfo.Current.WithdrawList); j++ {
			if valiInfo.Current.WithdrawList[j].WithDrawTime < reFundTime {
				valiInfo.Current.WithdrawList = append(valiInfo.Current.WithdrawList[:j], valiInfo.Current.WithdrawList[j+1:]...)
				j--
			}
		}
		for j := 0; j < len(valiInfo.Positions); j++ {
			depPos := &valiInfo.Positions[j]
			if msg, exist := depMap[depPos.Position]; exist {
				depPos.Amount = msg.DepositAmount
				if len(msg.WithDrawInfolist) > 0 && msg.WithDrawInfolist[0].WithDrawTime > 0 {
					depPos.EndTime = msg.EndTime
				} else {
					depPos.EndTime = 0
				}
				depPos.DType = msg.DepositType
				if depPos.EndTime == 0 || depPos.EndTime > time {
					valiInfo.AllAmount.Add(valiInfo.AllAmount, depPos.Amount)
				}
			} else {
				valiInfo.Positions = append(valiInfo.Positions[:j], valiInfo.Positions[j+1:]...)
				j--
			}

		}
	}
	return err
}
func (vc *ValidatorGroupState) SuicideContract(contractAddress common.Address, state StateDBManager) (bool, error) {
	if vc.CheckEmptyValidatorGroupState(contractAddress, state) {
		state.Suicide(params.MAN_COIN, contractAddress)
		parent := &ValidatorContractState{}
		err := parent.RemoveEmptyValidatorGroup(contractAddress, state)
		if err == nil {
			return true, err
		}
		return false, err
	}
	return false, nil
}
func (vc *ValidatorGroupState) CheckValidatorInfo(valiInfo *validatorGroup.ValidatorInfo, contractAddress common.Address, state StateDBManager) error {
	if valiInfo != nil && len(valiInfo.Positions) == 0 {
		if valiInfo.Reward.Sign() == 0 && valiInfo.Current.Amount.Sign() == 0 && len(valiInfo.Current.WithdrawList) == 0 && valiInfo.AllAmount.Sign() == 0 {
			vc.ValidatorMap.Remove(valiInfo.Address)

			//			_,err := vc.SuicideContract(contractAddress,state)
			//			if err != nil{
			//				return err
			//			}
		}
	}
	return nil
}
func (vc *ValidatorGroupState) CheckEmptyValidatorGroupState(contractAddress common.Address, state StateDBManager) bool {
	if len(vc.ValidatorMap) == 0 {
		balance := state.GetBalance(params.MAN_COIN, contractAddress)
		if balance[common.MainAccount].Balance.Sign() == 0 {
			return true
		}
	}
	return false
}
func (vc *ValidatorGroupState) SetState(contractAddress common.Address, state StateDBManager) error {
	suicide, err := vc.SuicideContract(contractAddress, state)
	if err != nil {
		return err
	}
	if suicide {
		return nil
	}
	err = vc.SetOwner(contractAddress, state)
	if err != nil {
		return err
	}
	err = vc.SetReward(contractAddress, state)
	if err != nil {
		return err
	}
	err = vc.SetValidatorMap(contractAddress, state)
	if err != nil {
		return err
	}
	return nil
}
func (vc *ValidatorGroupState) GetState(contractAddress common.Address, time uint64, state StateDBManager) error {
	err := vc.GetOwner(contractAddress, state)
	if err != nil {
		return err
	}
	err = vc.GetReward(contractAddress, state)
	if err != nil {
		return err
	}
	err = vc.GetValidatorMap(contractAddress, time, state)
	if err != nil {
		return err
	}
	return nil
}
func (vc *ValidatorGroupState) CalDepositWeight(address common.Address, amount *big.Int) *big.Int {
	if address == vc.OwnerInfo.Owner {
		return vc.Reward.OwnerRate.Mul(amount)
	} else {
		for i := len(vc.Reward.LevelRate) - 1; i >= 0; i-- {
			if amount.Cmp(vc.Reward.LevelRate[i].Threshold) >= 0 {
				return vc.Reward.LevelRate[i].Rate.Mul(amount)
			}
		}
	}
	return big.NewInt(0)
}
func (vc *ValidatorGroupState) DistributeAmount(amount *big.Int,getDepoist func(*validatorGroup.ValidatorInfo) *big.Int,addFunc func(*validatorGroup.ValidatorInfo,*big.Int)) error{
	if amount.Sign() == 0 {
		return nil
	}
	nodeAmount := vc.Reward.NodeRate.Mul(amount)
	rewards := new(big.Int).Sub(amount,nodeAmount)
	weightInfo := make([]DepositInfo, 0, len(vc.ValidatorMap))
	allWeight := big.NewInt(0)
	for i := 0; i < len(vc.ValidatorMap); i++ {
		valiInfo := &vc.ValidatorMap[i]
		weight := vc.CalDepositWeight(valiInfo.Address, getDepoist(valiInfo))
		allWeight.Add(allWeight, weight)
		weightInfo = append(weightInfo, DepositInfo{valiInfo.Address, weight})
	}
	leftAmount := new(big.Int).Set(rewards)
	if allWeight.Sign() > 0 {
		for i, info := range weightInfo {
			reward := new(big.Int).Mul(rewards, info.Amount)
			reward.Div(reward, allWeight)
			leftAmount.Sub(leftAmount,reward)
			addFunc(&vc.ValidatorMap[i],reward)
		}
	}
	if index, exist := vc.ValidatorMap.Find(vc.OwnerInfo.Owner);exist {
		addFunc(&vc.ValidatorMap[index],nodeAmount)
		if leftAmount.Sign()>0 {
			addFunc(&vc.ValidatorMap[index],leftAmount)
		}
	}
	return nil
}
func (vc *ValidatorGroupState) DistributeRewards(amount *big.Int) error {
	getDepoist := func(info *validatorGroup.ValidatorInfo) *big.Int {
		return info.AllAmount
	}
	addFunc := func(info *validatorGroup.ValidatorInfo, amount *big.Int){
		info.Reward.Add(info.Reward,amount)
	}
	return vc.DistributeAmount(amount,getDepoist,addFunc)
	/*
	nodeAmount := vc.Reward.NodeRate.Mul(amount)
	rewards := new(big.Int).Sub(amount,nodeAmount)
	weightInfo := make([]DepositInfo, 0, len(vc.ValidatorMap))
	allWeight := big.NewInt(0)
	for i := 0; i < len(vc.ValidatorMap); i++ {
		valiInfo := &vc.ValidatorMap[i]
		weight := vc.CalDepositWeight(valiInfo.Address, valiInfo.AllAmount)
		allWeight.Add(allWeight, weight)
		weightInfo = append(weightInfo, DepositInfo{valiInfo.Address, weight})
	}
	leftAmount := new(big.Int).Set(rewards)
	if allWeight.Sign() > 0 {
		for i, info := range weightInfo {
			reward := new(big.Int).Mul(rewards, info.Amount)
			reward.Div(reward, allWeight)
			leftAmount.Sub(leftAmount,reward)
			vc.ValidatorMap[i].Reward.Add(vc.ValidatorMap[i].Reward, reward)
		}
	}
	if index, exist := vc.ValidatorMap.Find(vc.OwnerInfo.Owner);exist {
		vc.ValidatorMap[index].Reward.Add(vc.ValidatorMap[index].Reward,leftAmount)
	}
	return nil
*/
}
func (vc *ValidatorGroupState) DistributeCurrentInterests(amount *big.Int) error {
	getDepoist := func(info *validatorGroup.ValidatorInfo) *big.Int {
		return info.Current.Amount
	}
	addFunc := func(info *validatorGroup.ValidatorInfo, amount *big.Int){
		info.Current.Interest.Add(info.Current.Interest,amount)
	}
	return vc.DistributeAmount(amount,getDepoist,addFunc)
	/*
	if amount.Sign() == 0 {
		return nil
	}
	weightInfo := make([]DepositInfo, 0, len(vc.ValidatorMap))
	allWeight := big.NewInt(0)
	for i := 0; i < len(vc.ValidatorMap); i++ {
		valiInfo := &vc.ValidatorMap[i]
		amount := valiInfo.Current.Amount
		if valiInfo.Current.PreAmount.Sign() > 0 && valiInfo.Current.PreAmount.Cmp(valiInfo.Current.Amount) < 0 {
			amount = valiInfo.Current.PreAmount
		}
		weight := vc.CalDepositWeight(valiInfo.Address, amount)
		allWeight.Add(allWeight, weight)
		weightInfo = append(weightInfo, DepositInfo{valiInfo.Address, weight})
		valiInfo.Current.PreAmount.Set(valiInfo.Current.Amount)
	}
	leftAmount := new(big.Int).Set(amount)
	if allWeight.Sign() > 0 {
		for i, info := range weightInfo {
			reward := new(big.Int).Mul(amount, info.Amount)
			reward.Div(reward, allWeight)
			leftAmount.Sub(leftAmount,reward)
			vc.ValidatorMap[i].Current.Interest.Add(vc.ValidatorMap[i].Current.Interest, reward)
		}
	}
	if leftAmount.Sign()>0{
		if index, exist := vc.ValidatorMap.Find(vc.OwnerInfo.Owner);exist {
			vc.ValidatorMap[index].Current.Interest.Add(vc.ValidatorMap[index].Current.Interest,leftAmount)
		}
	}
	return nil
	*/

}
