package util

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/log"
	"math/big"
	"testing"
)

func TestCalcDepositRate(t *testing.T) {
	log.InitLog(3)
	deposit := make(map[common.Address]*big.Int, 0)
	rewards := make(map[common.Address]*big.Int, 0)
	deposit[common.HexToAddress("0x0000000000000000000000000000000000000000")] = new(big.Int).Mul(big.NewInt(19779005), big.NewInt(1e18))
	deposit[common.HexToAddress("0x0000000000000000000000000000000000000001")] = new(big.Int).Mul(big.NewInt(3053790), big.NewInt(1e18))
	deposit[common.HexToAddress("0x0000000000000000000000000000000000000002")] = new(big.Int).Mul(big.NewInt(74399986), big.NewInt(1e18))
	deposit[common.HexToAddress("0x0000000000000000000000000000000000000003")] = new(big.Int).Mul(big.NewInt(49997244), big.NewInt(1e18))
	deposit[common.HexToAddress("0x0000000000000000000000000000000000000004")] = new(big.Int).Mul(big.NewInt(47987415), big.NewInt(1e18))
	deposit[common.HexToAddress("0x0000000000000000000000000000000000000005")] = new(big.Int).Mul(big.NewInt(90463177), big.NewInt(1e18))
	deposit[common.HexToAddress("0x0000000000000000000000000000000000000006")] = new(big.Int).Mul(big.NewInt(60980567), big.NewInt(1e18))
	deposit[common.HexToAddress("0x0000000000000000000000000000000000000007")] = new(big.Int).Mul(big.NewInt(61760463), big.NewInt(1e18))
	deposit[common.HexToAddress("0x0000000000000000000000000000000000000008")] = new(big.Int).Mul(big.NewInt(85935637), big.NewInt(1e18))
	deposit[common.HexToAddress("0x0000000000000000000000000000000000000009")] = new(big.Int).Mul(big.NewInt(80540888), big.NewInt(1e18))
	deposit[common.HexToAddress("0x000000000000000000000000000000000000000a")] = new(big.Int).Mul(big.NewInt(57666385), big.NewInt(1e18))
	reward := new(big.Int).Mul(big.NewInt(698745832334794), big.NewInt(1e8))
	CalcDepositRate(reward, deposit, rewards)
}
