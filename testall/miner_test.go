// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package testall

import (
	"fmt"
	"testing"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/p2p/discover"
)

func TestAsd(t *testing.T) {
	var v_s_account = []string{"0x1a8557a5830113ad675a9cb6f2d8a46d471edb8e", "0x20f4c8656cbac7de0e56e3e39d63e872393f089d"}
	var v_account = []common.Address{}
	var v_s_id = []string{"18b37bc680e739836fe0d8cca7c03a08a1162ff30fbb5049151d7bda951ec6e053916cb5df78f9b87e28a657230cc17d96cbde5fddc2f4571e5b606ec2a3a7a8", "4b2f638f46c7ae5b1564ca7015d716621848a0d9be66f1d1e91d566d2a70eedc2f11e92b743acb8d97dec3fb412c1b2f66afd7fbb9399d4fb2423619eaa514c7"}
	var v_id = []discover.NodeID{}
	/*
		var m = []string{}

		var account2node = map[string]string{
			"0x1a8557a5830113ad675a9cb6f2d8a46d471edb8e": "18b37bc680e739836fe0d8cca7c03a08a1162ff30fbb5049151d7bda951ec6e053916cb5df78f9b87e28a657230cc17d96cbde5fddc2f4571e5b606ec2a3a7a8",
			"0x20f4c8656cbac7de0e56e3e39d63e872393f089d": "4b2f638f46c7ae5b1564ca7015d716621848a0d9be66f1d1e91d566d2a70eedc2f11e92b743acb8d97dec3fb412c1b2f66afd7fbb9399d4fb2423619eaa514c7",
		}
	*/
	for _, v := range v_s_account {
		v_account = append(v_account, common.HexToAddress(v))
	}
	fmt.Println(v_account)

	for _, v := range v_s_id {
		ans, err := discover.HexID(v)
		fmt.Println(ans, err)
		v_id = append(v_id, ans)
	}
	fmt.Println(v_id)
}
