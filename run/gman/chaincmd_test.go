package main

import (
	"encoding/json"
	"fmt"
	"github.com/MatrixAINetwork/go-matrix/base58"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core"
	"github.com/MatrixAINetwork/go-matrix/log"
	"math/big"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type TestStruct1 struct {
	T1 common.Signature
	T2 []common.Signature
	T3 []*common.Signature
	T4 *common.Signature
}

func TestSignatureMarshal(t *testing.T) {
	testa := TestStruct1{
		T1: common.Signature{12, 13, 14},
		T2: []common.Signature{
			common.Signature{12, 13, 15},
			common.Signature{12, 13, 16},
			common.Signature{12, 13, 17},
		},
		T3: []*common.Signature{
			&common.Signature{
				181,
				8,
				246,
				28,
				118,
				103,
				127,
				70,
				144,
				31,
				187,
				28,
				71,
				14,
				164,
				113,
				133,
				96,
				141,
				160,
				117,
				234,
				127,
				5,
				254,
				240,
				146,
				127,
				39,
				247,
				161,
				150,
				75,
				243,
				248,
				192,
				32,
				110,
				149,
				242,
				151,
				195,
				226,
				167,
				74,
				223,
				135,
				250,
				233,
				174,
				109,
				239,
				101,
				177,
				155,
				129,
				68,
				92,
				218,
				222,
				45,
				207,
				165,
				112,
				0,
			},
			&common.Signature{12, 13, 26},
			&common.Signature{12, 13, 27},
		},
		T4: &common.Signature{12, 13, 35},
	}
	buff, err := json.Marshal(testa)
	if err != nil {
		t.Error(err)
	}
	stringA := string(buff)
	t.Log(string(buff))
	testb := &TestStruct1{}
	err = json.Unmarshal(buff, testb)
	if err != nil {
		t.Error(err)
	}
	buff, err = json.Marshal(testb)
	stringB := string(buff)
	t.Log(string(buff))
	if stringA != stringB {
		t.Error("json Marshal and unmarshal error")
	}
}
func TestInitGenesis(t *testing.T) {
	log.InitLog(3)
	genesisPath := filepath.Join(".", "MANGenesis.json")
	defGen, err := core.DefaultGenesis(genesisPath)
	if err != nil {
		t.Errorf("Failed to read genesis file: %v", err)
	}
	dec := make(map[string]interface{})
	file, err := os.Open(genesisPath)
	if err != nil {
		t.Errorf("Failed to read genesis file: %v", err)
	}
	defer file.Close()
	err = json.NewDecoder(file).Decode(&dec)
	if err != nil {
		t.Log(err)
	}
	checkAlloc(dec, defGen, t)
	checkNetTopology(dec, defGen, t)
	checkConfig(dec, defGen, t)

	//	str,_ := json.Marshal(defGen)
	//	t.Log(string(str))

}

var allocGenesis = `{
	"alloc"      : {"MAN.2nRsUetjWAaYUizRkgBxGETimfUTz":{
            "balance":"10000000000000000000000000"
        },
        "MAN.2nRsUetjWAaYUizRkgBxGETimfUUs":{
            "balance":"25000000000000000000000000"
        },
        "MAN.2nRsUetjWAaYUizRkgBxGETimfUV2":{
            "balance":"10000000000000000000000000"
        },
        "MAN.2nRsUetjWAaYUizRkgBxGETimfUW7":{
            "balance":"5000000000000000000000000"
        },
        "MAN.2nRsUetjWAaYUizRkgBxGETimfUXN":{
            "balance":"10000000000000000000000000"
        },
        "MAN.4L95KmR3e8eUJvzwK2thft1eKaFYa":{
            "balance":"300000000000000000000000000"
        },
        "MAN.4739r322TyL3xCpbbdohS8NhBgGwi":{
            "balance":"200000000000000000000000000"
        },
        "MAN.2zXWsDtyt7vhVADGTz2yXD6h7WJnF":{
            "balance":"87650000000000000000000000"
        }
    },
	"coinbase"   : "0x0000000000000000000000000000000000000000",
	"difficulty" : "0x20000",
	"extraData"  : "",
	"gasLimit"   : "0x2fefd8",
	"nonce"      : "0x0000000000000042",
	"mixhash"    : "0x0000000000000000000000000000000000000000000000000000000000000000",
	"parentHash" : "0x0000000000000000000000000000000000000000000000000000000000000000",
	"timestamp"  : "0x00",
	"config"     : {
		"daoForkBlock"   : 314,
		"daoForkSupport" : false
	}
}`

type AllocAccount struct {
	Alloc map[core.GenesisAddress]core.GenesisAccount `json:"alloc"`
}

func TestUnMarshalGenesisAccount(t *testing.T) {
	log.InitLog(3)
	alloc := AllocAccount{
		make(map[core.GenesisAddress]core.GenesisAccount),
	}
	//	temp := common.Signature{}
	//	mar,_:=temp.MarshalJSON()
	//	t.Log(string(mar))
	alloc.Alloc[core.GenesisAddress{12, 13, 14, 15, 16, 17}] = core.GenesisAccount{Balance: big.NewInt(200)}
	alloc.Alloc[core.GenesisAddress{22, 23, 24, 25, 26, 27}] = core.GenesisAccount{Balance: big.NewInt(300)}
	alloc.Alloc[core.GenesisAddress{32, 33, 24, 35, 26, 27}] = core.GenesisAccount{Balance: big.NewInt(400)}
	for key, _ := range alloc.Alloc {
		t.Log(common.Bytes2Hex(key[:]))
	}
	buff, err := json.Marshal(alloc)
	if err != nil {
		t.Log(err)
	}
	t.Log(string(buff))
	alloc1 := AllocAccount{
		make(map[core.GenesisAddress]core.GenesisAccount),
	}
	err = json.Unmarshal(buff, &alloc1)
	if err != nil {
		t.Log(err)
	}
}
func TestUnMarshalGenesisAlloc(t *testing.T) {
	genesis := new(core.Genesis)
	err := json.Unmarshal([]byte(allocGenesis), genesis)
	if err != nil {
		t.Log(err)
	}
	dec := make(map[string]interface{})
	err = json.Unmarshal([]byte(allocGenesis), &dec)
	if err != nil {
		t.Log(err)
	}
	checkAlloc(dec, genesis, t)

}

func TestMarshalGenesis(t *testing.T) {
	genesis := new(core.Genesis)
	err := json.Unmarshal([]byte(allocGenesis), genesis)
	if err != nil {
		t.Log(err)
	}
	dec := make(map[string]interface{})
	err = json.Unmarshal([]byte(allocGenesis), &dec)
	if err != nil {
		t.Log(err)
	}
	checkAlloc(dec, genesis, t)
	if _, err := json.Marshal(genesis); err != nil {
		fmt.Println(err)
	}
	/*	if out,err := json.Marshal(genesis                   ); err != nil{fmt.Println( out)}
		if _,err := json.Marshal(genesis.Config            ); err != nil{fmt.Println( "  genesisConfig             ")}
		if _,err := json.Marshal(genesis.Nonce             ); err != nil{fmt.Println( "  genesisNonce              ")}
		if _,err := json.Marshal(genesis.Timestamp         ); err != nil{fmt.Println( "  genesisTimestamp          ")}
		if _,err := json.Marshal(genesis.ExtraData         ); err != nil{fmt.Println( "  genesisExtraData          ")}
		if _,err := json.Marshal(genesis.Version           ); err != nil{fmt.Println( "  genesisVersion            ")}
		if data,err := json.Marshal(genesis.VersionSignatures ); err == nil{fmt.Println("%s", data)}
		if _,err := json.Marshal(genesis.VrfValue          ); err != nil{fmt.Println( "  genesisVrfValue           ")}
		if _,err := json.Marshal(genesis.Leader            ); err != nil{fmt.Println( "  genesisLeader             ")}
		if _,err := json.Marshal(genesis.NextElect         ); err != nil{fmt.Println( "  genesisNextElect          ")}
		if _,err := json.Marshal(genesis.NetTopology       ); err != nil{fmt.Println( "  genesisNetTopology        ")}
		if _,err := json.Marshal(genesis.Signatures        ); err != nil{fmt.Println( "  genesisSignatures         ")}
		if _,err := json.Marshal(genesis.GasLimit          ); err != nil{fmt.Println( "  genesisGasLimit           ")}
		if _,err := json.Marshal(genesis.Difficulty        ); err != nil{fmt.Println( "  genesisDifficulty         ")}
		if _,err := json.Marshal(genesis.Mixhash           ); err != nil{fmt.Println( "  genesisMixhash            ")}
		if _,err := json.Marshal(genesis.Coinbase          ); err != nil{fmt.Println( "  genesisCoinbase           ")}
		if _,err := json.Marshal(genesis.Alloc             ); err != nil{fmt.Println( "  genesisAlloc              ")}
		if _,err := json.Marshal(genesis.MState            ); err != nil{fmt.Println( "  genesisMState             ")}
		if _,err := json.Marshal(genesis.Number            ); err != nil{fmt.Println( "  genesisNumber             ")}
		if _,err := json.Marshal(genesis.GasUsed           ); err != nil{fmt.Println( "  genesisGasUsed            ")}
		if _,err := json.Marshal(genesis.ParentHash        ); err != nil{fmt.Println( "  genesisParentHash         ")}
		if _,err := json.Marshal(genesis.Root              ); err != nil{fmt.Println( "  genesisRoot               ")}
		if _,err := json.Marshal(genesis.TxHash            ); err != nil{fmt.Println( "  genesisTxHash             ")}*/
}
func TestDefaultGenesisAlloc(t *testing.T) {
	genesis := new(core.Genesis)
	err := json.Unmarshal([]byte(core.DefaultGenesisJson), genesis)
	if err != nil {
		t.Log(err)
	}
	dec := make(map[string]interface{})
	err = json.Unmarshal([]byte(core.DefaultGenesisJson), &dec)
	if err != nil {
		t.Log(err)
	}
	checkAlloc(dec, genesis, t)
	checkNetTopology(dec, genesis, t)
	checkConfig(dec, genesis, t)
}
func TestAllGenesisAlloc(t *testing.T) {
	genesis := new(core.Genesis)
	err := json.Unmarshal([]byte(core.AllGenesisJson), genesis)
	if err != nil {
		t.Log(err)
	}
	dec := make(map[string]interface{})
	err = json.Unmarshal([]byte(core.AllGenesisJson), &dec)
	if err != nil {
		t.Log(err)
	}
	checkAlloc(dec, genesis, t)
	checkNetTopology(dec, genesis, t)
	checkConfig(dec, genesis, t)
}
func checkAlloc(dec map[string]interface{}, genesis *core.Genesis, t *testing.T) bool {
	alloc := dec["alloc"]
	if len(alloc.(map[string]interface{})) != len(genesis.Alloc) {
		t.Error("Alloc Length is error")
		return false
	}
	allMap := alloc.(map[string]interface{})
	for key, item := range genesis.Alloc {
		key1 := base58.Base58EncodeToString("MAN", key)
		value, exist := allMap[key1]
		if !exist {
			t.Error("Address error : " + key1)
			continue
		}
		account := value.(map[string]interface{})
		balance, exist := account["balance"]
		if !exist {
			if item.Balance.Sign() != 0 {
				t.Error("Address Balance is not zero ! error : " + key1)
				continue
			}
		} else {
			bal := new(big.Int)
			json.Unmarshal([]byte(balance.(string)), bal)
			if bal.Cmp(item.Balance) != 0 {
				t.Error("Address Balance is not Equal ! error : " + key1)
			}
		}

	}
	return true
}
func checkNetTopology(dec map[string]interface{}, genesis *core.Genesis, t *testing.T) bool {
	topology, exit := dec["nettopology"]
	if !exit {
		return true
	}
	netMap := topology.(map[string]interface{})
	netTop := netMap["NetTopologyData"]
	if netTop == nil {
		return true
	}
	if len(netTop.([]interface{})) != len(genesis.NetTopology.NetTopologyData) {
		t.Error("NetTopologyData Length is error")
		return false
	}
	allMap := netTop.([]interface{})
	for i, item := range genesis.NetTopology.NetTopologyData {
		itemMap := allMap[i].(map[string]interface{})
		account := itemMap["Account"].(string)
		key := base58.Base58EncodeToString("MAN", item.Account)
		if key != account {
			t.Error("NetTopologyData Account Error : ", key)
		}
		value := itemMap["Position"]
		_val := reflect.ValueOf(value)
		if _val.Kind() == reflect.Float32 || _val.Kind() == reflect.Float64 {
			if _val.Float()-float64(item.Position) > 0.9 {
				t.Error("NetTopologyData Account Pos Error : ", key)
			}
		} else if _val.Kind() >= reflect.Int && _val.Kind() <= reflect.Int64 {
			if _val.Int() != int64(item.Position) {
				t.Error("NetTopologyData Account Pos Error : ", key)
			}
		}
	}
	return true
}
func checkConfig(dec map[string]interface{}, genesis *core.Genesis, t *testing.T) bool {
	config, exist := dec["config"]
	if !exist {
		return true
	}
	allMap := config.(map[string]interface{})
	for key, item := range allMap {
		switch key {
		case "chainID":
			if !checkBigInt(genesis.Config.ChainId, item) {
				t.Error("ChainID error")
			}
		case "byzantiumBlock":
			if !checkBigInt(genesis.Config.ByzantiumBlock, item) {
				t.Error("ByzantiumBlock error")
			}
		case "homesteadBlock":
			if !checkBigInt(genesis.Config.HomesteadBlock, item) {
				t.Error("homesteadBlock error")
			}
		case "eip155Block":
			if !checkBigInt(genesis.Config.EIP155Block, item) {
				t.Error("eip155Block error")
			}
		case "eip158Block":
			if !checkBigInt(genesis.Config.EIP158Block, item) {
				t.Error("eip158Block error")
			}

		}
	}
	if genesis.Config.ChainId == nil {
		t.Error("ChainID Nil")
	}
	if genesis.Config.ByzantiumBlock == nil {
		t.Error("ByzantiumBlock Nil")
	}
	if genesis.Config.HomesteadBlock == nil {
		t.Error("HomesteadBlock Nil")
	}
	if genesis.Config.EIP155Block == nil {
		t.Error("EIP155Block Nil")
	}
	if genesis.Config.EIP158Block == nil {
		t.Error("EIP158Block Nil")
	}
	return true
}
func checkBigInt(src *big.Int, dest interface{}) bool {
	bal := new(big.Int)
	_val := reflect.ValueOf(dest)
	switch {
	case _val.Kind() == reflect.String:

		json.Unmarshal([]byte(dest.(string)), bal)
	case _val.Kind() <= reflect.Int && _val.Kind() >= reflect.Int64:
		bal.SetInt64(_val.Int())
	case _val.Kind() <= reflect.Uint && _val.Kind() >= reflect.Uint64:
		bal.SetUint64(_val.Uint())
	case _val.Kind() == reflect.Float32 || _val.Kind() == reflect.Float64:
		value := _val.Float()

		if float64(src.Int64())-value > 0.9 {
			return false
		} else {
			return true
		}
	default:
		log.Info("Not support big int type")
	}
	if bal.Cmp(src) != 0 {
		return false
	}
	return true
}
