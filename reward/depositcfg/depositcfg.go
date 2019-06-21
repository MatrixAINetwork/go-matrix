package depositcfg

type DepositCfgInterface interface {
	GetDepositPositionCfg(depositType uint64) DepositCfger
}

const (
	VersionA = "A"
)

var depositCfgMap map[string]DepositCfgInterface

func init() {
	depositCfgMap = make(map[string]DepositCfgInterface)
	depositCfgMap[VersionA] = newDepositCfgA()
}

func GetDepositCfg(version string) DepositCfgInterface {
	return depositCfgMap[version]
}
