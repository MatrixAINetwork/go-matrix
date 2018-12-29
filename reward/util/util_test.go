package util

import (
	"math/big"
	"testing"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/log"
)

func TestCalcDepositRate(t *testing.T) {
	log.InitLog(5)
	deposit := make(map[common.Address]DepositInfo, 0)
	deposit[common.HexToAddress("0x0000000000000000000000000000000000000000")] = DepositInfo{new(big.Int).Mul(big.NewInt(19779005), big.NewInt(1e18)), 9000}
	deposit[common.HexToAddress("0x0000000000000000000000000000000000000001")] = DepositInfo{new(big.Int).Mul(big.NewInt(19779005), big.NewInt(1e18)), 120000}
	deposit[common.HexToAddress("0x0000000000000000000000000000000000000002")] = DepositInfo{new(big.Int).Mul(big.NewInt(19779005), big.NewInt(1e18)), 30000}
	deposit[common.HexToAddress("0x0000000000000000000000000000000000000003")] = DepositInfo{new(big.Int).Mul(big.NewInt(19779005), big.NewInt(1e18)), 30000}
	deposit[common.HexToAddress("0x0000000000000000000000000000000000000004")] = DepositInfo{new(big.Int).Mul(big.NewInt(19779005), big.NewInt(1e18)), 30000}
	deposit[common.HexToAddress("0x0000000000000000000000000000000000000005")] = DepositInfo{new(big.Int).Mul(big.NewInt(19779005), big.NewInt(1e18)), 30000}
	deposit[common.HexToAddress("0x0000000000000000000000000000000000000006")] = DepositInfo{new(big.Int).Mul(big.NewInt(19779005), big.NewInt(1e18)), 30000}
	deposit[common.HexToAddress("0x0000000000000000000000000000000000000007")] = DepositInfo{new(big.Int).Mul(big.NewInt(19779005), big.NewInt(1e18)), 30000}
	deposit[common.HexToAddress("0x0000000000000000000000000000000000000008")] = DepositInfo{new(big.Int).Mul(big.NewInt(19779005), big.NewInt(1e18)), 30000}
	deposit[common.HexToAddress("0x0000000000000000000000000000000000000009")] = DepositInfo{new(big.Int).Mul(big.NewInt(19779005), big.NewInt(1e18)), 30000}
	deposit[common.HexToAddress("0x000000000000000000000000000000000000000a")] = DepositInfo{new(big.Int).Mul(big.NewInt(19779005), big.NewInt(1e18)), 30000}
	reward := new(big.Int).Mul(big.NewInt(698745832334794), big.NewInt(1e18))
	CalcStockRate(reward, deposit)
}

//func TestCalcDepositRate1(t *testing.T) {
//	log.InitLog(3)
//	deposit := make(map[common.Address]*big.Int, 0)
//
//	deposit[common.HexToAddress("0x0000000000000000000000000000000000000000")] = new(big.Int).Mul(big.NewInt(19779005), big.NewInt(-1e18))
//	deposit[common.HexToAddress("0x0000000000000000000000000000000000000001")] = new(big.Int).Mul(big.NewInt(3053790), big.NewInt(-1e18))
//	deposit[common.HexToAddress("0x0000000000000000000000000000000000000002")] = new(big.Int).Mul(big.NewInt(74399986), big.NewInt(-1e18))
//	deposit[common.HexToAddress("0x0000000000000000000000000000000000000003")] = new(big.Int).Mul(big.NewInt(49997244), big.NewInt(-1e18))
//	deposit[common.HexToAddress("0x0000000000000000000000000000000000000004")] = new(big.Int).Mul(big.NewInt(47987415), big.NewInt(-1e18))
//	deposit[common.HexToAddress("0x0000000000000000000000000000000000000005")] = new(big.Int).Mul(big.NewInt(90463177), big.NewInt(-1e18))
//	deposit[common.HexToAddress("0x0000000000000000000000000000000000000006")] = new(big.Int).Mul(big.NewInt(60980567), big.NewInt(-1e18))
//	deposit[common.HexToAddress("0x0000000000000000000000000000000000000007")] = new(big.Int).Mul(big.NewInt(61760463), big.NewInt(-1e18))
//	deposit[common.HexToAddress("0x0000000000000000000000000000000000000008")] = new(big.Int).Mul(big.NewInt(85935637), big.NewInt(-1e18))
//	deposit[common.HexToAddress("0x0000000000000000000000000000000000000009")] = new(big.Int).Mul(big.NewInt(80540888), big.NewInt(1e18))
//	deposit[common.HexToAddress("0x000000000000000000000000000000000000000a")] = new(big.Int).Mul(big.NewInt(57666385), big.NewInt(1e18))
//	reward := new(big.Int).Mul(big.NewInt(698745832334794), big.NewInt(1e8))
//	CalcDepositRate(reward, deposit)
//}
//
//func TestCalcDepositRate2(t *testing.T) {
//	log.InitLog(3)
//	deposit := make(map[common.Address]*big.Int, 0)
//
//	deposit[common.HexToAddress("0x0000000000000000000000000000000000000000")] = new(big.Int).Mul(big.NewInt(1), big.NewInt(1e17))
//	deposit[common.HexToAddress("0x0000000000000000000000000000000000000001")] = new(big.Int).Mul(big.NewInt(1), big.NewInt(1e17))
//	deposit[common.HexToAddress("0x0000000000000000000000000000000000000002")] = new(big.Int).Mul(big.NewInt(1), big.NewInt(1e17))
//	deposit[common.HexToAddress("0x0000000000000000000000000000000000000003")] = new(big.Int).Mul(big.NewInt(1), big.NewInt(1e17))
//	deposit[common.HexToAddress("0x0000000000000000000000000000000000000004")] = new(big.Int).Mul(big.NewInt(1), big.NewInt(1e17))
//	deposit[common.HexToAddress("0x0000000000000000000000000000000000000005")] = new(big.Int).Mul(big.NewInt(1), big.NewInt(1e17))
//	deposit[common.HexToAddress("0x0000000000000000000000000000000000000006")] = new(big.Int).Mul(big.NewInt(1), big.NewInt(1e17))
//	deposit[common.HexToAddress("0x0000000000000000000000000000000000000007")] = new(big.Int).Mul(big.NewInt(1), big.NewInt(1e17))
//	deposit[common.HexToAddress("0x0000000000000000000000000000000000000008")] = new(big.Int).Mul(big.NewInt(1), big.NewInt(1e17))
//	deposit[common.HexToAddress("0x0000000000000000000000000000000000000009")] = new(big.Int).Mul(big.NewInt(1), big.NewInt(1e17))
//	deposit[common.HexToAddress("0x000000000000000000000000000000000000000a")] = new(big.Int).Mul(big.NewInt(1), big.NewInt(1e17))
//	reward := new(big.Int).Mul(big.NewInt(698745832334794), big.NewInt(1e8))
//	CalcDepositRate(reward, deposit)
//}
//
//func TestCalcDepositRate3(t *testing.T) {
//	log.InitLog(3)
//	deposit := make(map[common.Address]*big.Int, 0)
//
//	deposit[common.HexToAddress("0x0000000000000000000000000000000000000000")] = big.NewInt(0)
//	deposit[common.HexToAddress("0x0000000000000000000000000000000000000001")] = big.NewInt(0)
//	deposit[common.HexToAddress("0x0000000000000000000000000000000000000002")] = big.NewInt(0)
//	deposit[common.HexToAddress("0x0000000000000000000000000000000000000003")] = big.NewInt(0)
//	deposit[common.HexToAddress("0x0000000000000000000000000000000000000004")] = big.NewInt(0)
//	deposit[common.HexToAddress("0x0000000000000000000000000000000000000005")] = big.NewInt(0)
//	deposit[common.HexToAddress("0x0000000000000000000000000000000000000006")] = big.NewInt(0)
//	deposit[common.HexToAddress("0x0000000000000000000000000000000000000007")] = big.NewInt(0)
//	deposit[common.HexToAddress("0x0000000000000000000000000000000000000008")] = big.NewInt(0)
//	deposit[common.HexToAddress("0x0000000000000000000000000000000000000009")] = big.NewInt(0)
//	deposit[common.HexToAddress("0x000000000000000000000000000000000000000a")] = big.NewInt(0)
//	reward := new(big.Int).Mul(big.NewInt(698745832334794), big.NewInt(1e8))
//	CalcDepositRate(reward, deposit)
//}
//
//func TestCalcDepositRate4(t *testing.T) {
//	log.InitLog(3)
//	deposit := make(map[common.Address]*big.Int, 0)
//
//	deposit[common.HexToAddress("0x0000000000000000000000000000000000000000")] = new(big.Int).Mul(big.NewInt(19779005), big.NewInt(1e18))
//	deposit[common.HexToAddress("0x0000000000000000000000000000000000000001")] = new(big.Int).Mul(big.NewInt(3053790), big.NewInt(1e18))
//	deposit[common.HexToAddress("0x0000000000000000000000000000000000000002")] = new(big.Int).Mul(big.NewInt(74399986), big.NewInt(1e18))
//	deposit[common.HexToAddress("0x0000000000000000000000000000000000000003")] = new(big.Int).Mul(big.NewInt(49997244), big.NewInt(1e18))
//	deposit[common.HexToAddress("0x0000000000000000000000000000000000000004")] = new(big.Int).Mul(big.NewInt(47987415), big.NewInt(1e18))
//	deposit[common.HexToAddress("0x0000000000000000000000000000000000000005")] = new(big.Int).Mul(big.NewInt(90463177), big.NewInt(1e18))
//	deposit[common.HexToAddress("0x0000000000000000000000000000000000000006")] = new(big.Int).Mul(big.NewInt(60980567), big.NewInt(1e18))
//	deposit[common.HexToAddress("0x0000000000000000000000000000000000000007")] = new(big.Int).Mul(big.NewInt(61760463), big.NewInt(1e18))
//	deposit[common.HexToAddress("0x0000000000000000000000000000000000000008")] = new(big.Int).Mul(big.NewInt(85935637), big.NewInt(1e18))
//	deposit[common.HexToAddress("0x0000000000000000000000000000000000000009")] = new(big.Int).Mul(big.NewInt(80540888), big.NewInt(1e18))
//	deposit[common.HexToAddress("0x000000000000000000000000000000000000000a")] = new(big.Int).Mul(big.NewInt(57666385), big.NewInt(1e18))
//	reward := new(big.Int).Mul(big.NewInt(1), big.NewInt(1e7))
//	CalcDepositRate(reward, deposit)
//}
//
//func TestCalcDepositRate5(t *testing.T) {
//	log.InitLog(3)
//	deposit := make(map[common.Address]*big.Int, 0)
//
//	deposit[common.HexToAddress("0x000000000000000000000000000000000000000b")] = new(big.Int).Mul(big.NewInt(100000), big.NewInt(1e18))
//	deposit[common.HexToAddress("0x0000000000000000000000000000000000000001")] = new(big.Int).Mul(big.NewInt(100000), big.NewInt(1e18))
//	deposit[common.HexToAddress("0x0000000000000000000000000000000000000002")] = new(big.Int).Mul(big.NewInt(100000), big.NewInt(1e18))
//	deposit[common.HexToAddress("0x0000000000000000000000000000000000000003")] = new(big.Int).Mul(big.NewInt(100000), big.NewInt(1e18))
//	deposit[common.HexToAddress("0x0000000000000000000000000000000000000004")] = new(big.Int).Mul(big.NewInt(100000), big.NewInt(1e18))
//	deposit[common.HexToAddress("0x0000000000000000000000000000000000000005")] = new(big.Int).Mul(big.NewInt(100000), big.NewInt(1e18))
//	deposit[common.HexToAddress("0x0000000000000000000000000000000000000006")] = new(big.Int).Mul(big.NewInt(100000), big.NewInt(1e18))
//	deposit[common.HexToAddress("0x0000000000000000000000000000000000000007")] = new(big.Int).Mul(big.NewInt(100000), big.NewInt(1e18))
//	deposit[common.HexToAddress("0x0000000000000000000000000000000000000008")] = new(big.Int).Mul(big.NewInt(100000), big.NewInt(1e18))
//	deposit[common.HexToAddress("0x0000000000000000000000000000000000000009")] = new(big.Int).Mul(big.NewInt(100000), big.NewInt(1e18))
//	deposit[common.HexToAddress("0x000000000000000000000000000000000000000a")] = new(big.Int).Mul(big.NewInt(100000), big.NewInt(1e18))
//	reward := new(big.Int).Mul(big.NewInt(1), big.NewInt(13986000000000000*0.6))
//	CalcDepositRate(reward, deposit)
//}
