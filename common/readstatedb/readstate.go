package readstatedb

import (
	"github.com/MatrixAINetwork/go-matrix"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
)

const (
	ModuleReadStateDB = "读状态树"
)

func checkDataValidity(inputData interface{}) bool {
	return common.IsNil(inputData)
}

func GetElectGenTimes(stateReader matrix.StateReader, hash common.Hash) (*mc.ElectGenTimeStruct, error) {
	//if checkDataValidity(stateReader)==false{
	//	log.Error(ModuleReadStateDB,"获取选举时间点信息阶段,检查入参失败","入参为空")
	//	return nil,fmt.Errorf("获取选举时间点信息阶段,检查入参失败,入参为空")
	//}
	st, err := stateReader.StateAtBlockHash(hash)
	if err != nil {
		log.Error(ModuleReadStateDB, "获取state失败", err)
		return nil, err
	}

	electGenConfig, err := matrixstate.GetElectGenTime(st)
	if err != nil {
		log.Error(ModuleReadStateDB, "获取选举时间点信息阶段,从状态树获取失败,err", err)
		return nil, err
	}
	//if checkDataValidity(data)==false{
	//	log.Error(ModuleReadStateDB,"获取选举时间点信息阶段,检查入参失败","获取到的信息为空")
	//	return nil,fmt.Errorf("获取选举时间点信息阶段,检查入参失败,获取到的信息为空")
	//}
	if electGenConfig == nil {
		log.ERROR(ModuleReadStateDB, "获取到的选举配置信息错误", "反射后的数据为空")
		return nil, err
	}
	return electGenConfig, nil
}

func GetPreBroadcastRoot(stateReader matrix.StateReader, hash common.Hash) (*mc.PreBroadStateRoot, error) {
	//if checkDataValidity(stateReader)==false{
	//	log.Error(ModuleReadStateDB,"获取前两个广播区块root值阶段,检查入参失败","入参为空")
	//	return nil,fmt.Errorf("获取前两个广播区块root值阶段,检查入参失败,入参为空")
	//}
	st, err := stateReader.StateAtBlockHash(hash)
	if err != nil {
		log.Error(ModuleReadStateDB, "获取state失败", err)
	}
	preBroadStateRoot, err := matrixstate.GetPreBroadcastRoot(st)
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
	if preBroadStateRoot == nil {
		log.ERROR(ModuleReadStateDB, "获取前两个广播区块root值阶段", "反射后的数据为空")
		return nil, err
	}
	return preBroadStateRoot, nil
}

func GetRandomInfo(stateReader matrix.StateReader, hash common.Hash) (*mc.RandomInfoStruct, error) {
	//if checkDataValidity(stateReader)==false{
	//	log.Error(ModuleReadStateDB,"获取最小阶段,检查入参失败","入参为空")
	//	return nil,fmt.Errorf("获取最小阶段,检查入参失败,入参为空")
	//}
	st, err := stateReader.StateAtBlockHash(hash)
	if err != nil {
		log.Error(ModuleReadStateDB, "获取state失败", err)
	}
	randomInfo, err := matrixstate.GetMinHash(st)
	if err != nil {
		log.Error(ModuleReadStateDB, "获取最小阶段,从状态树获取失败,err", err)
		return nil, err
	}
	//if checkDataValidity(data)==false{
	//	log.Error(ModuleReadStateDB,"获取最小阶段,检查入参失败","获取到的信息为空")
	//	return nil,fmt.Errorf("获取最小阶段,检查入参失败,获取到的信息为空")
	//}
	if randomInfo == nil {
		log.ERROR(ModuleReadStateDB, "获取前两个广播区块root值阶段", "反射后的数据为空")
		return nil, err
	}
	return randomInfo, nil
}
