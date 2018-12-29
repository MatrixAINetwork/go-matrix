package readstatedb

import (
	"github.com/matrix/go-matrix"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
)

const (
	ModuleReadStateDB = "读状态树"
)

func checkDataValidity(inputData interface{}) bool {
	return common.IsNil(inputData)
}

func GetElectGenTimes(stateReader matrix.StateReader, height uint64) (*mc.ElectGenTimeStruct, error) {
	//if checkDataValidity(stateReader)==false{
	//	log.Error(ModuleReadStateDB,"获取选举时间点信息阶段,检查入参失败","入参为空")
	//	return nil,fmt.Errorf("获取选举时间点信息阶段,检查入参失败,入参为空")
	//}
	data, err := stateReader.GetMatrixStateDataByNumber(mc.MSKeyElectGenTime, height)
	if err != nil {
		log.Error(ModuleReadStateDB, "获取选举时间点信息阶段,从状态树获取失败,err", err)
		return nil, err
	}
	//if checkDataValidity(data)==false{
	//	log.Error(ModuleReadStateDB,"获取选举时间点信息阶段,检查入参失败","获取到的信息为空")
	//	return nil,fmt.Errorf("获取选举时间点信息阶段,检查入参失败,获取到的信息为空")
	//}
	electGenConfig, OK := data.(*mc.ElectGenTimeStruct)
	if OK == false {
		log.ERROR(ModuleReadStateDB, "获取到的选举配置信息错误", "反射失败")
		return nil, err
	}
	if electGenConfig == nil {
		log.ERROR(ModuleReadStateDB, "获取到的选举配置信息错误", "反射后的数据为空")
		return nil, err
	}
	return electGenConfig, nil
}

func GetPreBroadcastRoot(stateReader matrix.StateReader, height uint64) (*mc.PreBroadStateRoot, error) {
	//if checkDataValidity(stateReader)==false{
	//	log.Error(ModuleReadStateDB,"获取前两个广播区块root值阶段,检查入参失败","入参为空")
	//	return nil,fmt.Errorf("获取前两个广播区块root值阶段,检查入参失败,入参为空")
	//}
	data, err := stateReader.GetMatrixStateDataByNumber(mc.MSKeyPreBroadcastRoot, height)
	if err != nil {
		log.Error(ModuleReadStateDB, "获取前两个广播区块root值阶段,从状态树获取失败,err", err)
		return nil, err
	}
	//if checkDataValidity(data)==false{
	//	log.Error(ModuleReadStateDB,"获取前两个广播区块root值阶段,检查入参失败","获取到的信息为空")
	//	return nil,fmt.Errorf("获取前两个广播区块root值阶段,检查入参失败,获取到的信息为空")
	//}	//if checkDataValidity(data)==false{
	//	log.Error(ModuleReadStateDB,"获取前两个广播区块root值阶段,检查入参失败","获取到的信息为空")
	//	return nil,fmt.Errorf("获取前两个广播区块root值阶段,检查入参失败,获取到的信息为空")
	//}
	preBroadStateRoot, OK := data.(*mc.PreBroadStateRoot)
	if OK == false {
		log.ERROR(ModuleReadStateDB, "获取前两个广播区块root值阶段", "反射失败")
		return nil, err
	}
	if preBroadStateRoot == nil {
		log.ERROR(ModuleReadStateDB, "获取前两个广播区块root值阶段", "反射后的数据为空")
		return nil, err
	}
	return preBroadStateRoot, nil
}

func GetRandomInfo(stateReader matrix.StateReader, height uint64) (*mc.RandomInfoStruct, error) {
	//if checkDataValidity(stateReader)==false{
	//	log.Error(ModuleReadStateDB,"获取最小阶段,检查入参失败","入参为空")
	//	return nil,fmt.Errorf("获取最小阶段,检查入参失败,入参为空")
	//}
	data, err := stateReader.GetMatrixStateDataByNumber(mc.MSKeyMinHash, height)
	if err != nil {
		log.Error(ModuleReadStateDB, "获取最小阶段,从状态树获取失败,err", err)
		return nil, err
	}
	//if checkDataValidity(data)==false{
	//	log.Error(ModuleReadStateDB,"获取最小阶段,检查入参失败","获取到的信息为空")
	//	return nil,fmt.Errorf("获取最小阶段,检查入参失败,获取到的信息为空")
	//}
	randomInfo, OK := data.(*mc.RandomInfoStruct)
	if OK == false {
		log.ERROR(ModuleReadStateDB, "获取前两个广播区块root值阶段", "反射失败")
		return nil, err
	}
	if randomInfo == nil {
		log.ERROR(ModuleReadStateDB, "获取前两个广播区块root值阶段", "反射后的数据为空")
		return nil, err
	}
	return randomInfo, nil
}
