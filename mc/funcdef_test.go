// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package mc

import (
	//"reflect"
	"testing"

	"fmt"
	"github.com/MatrixAINetwork/go-matrix/common"
	"math/big"
)

func getTestNodeList() (list []TopologyNodeInfo) {
	list = append(list, TopologyNodeInfo{
		Account:  common.HexToAddress("0x0001"),
		Position: 1,
		Type:     common.RoleNil,
		//Stock:      0,
		NodeNumber: 0,
	})

	list = append(list, TopologyNodeInfo{
		Account:  common.HexToAddress("0x0002"),
		Position: 2,
		Type:     common.RoleNil,
		//	Stock:      0,
		NodeNumber: 0,
	})

	list = append(list, TopologyNodeInfo{
		Account:  common.HexToAddress("0x0007"),
		Position: 7,
		Type:     common.RoleNil,
		//	Stock:      0,
		NodeNumber: 0,
	})

	list = append(list, TopologyNodeInfo{
		Account:  common.HexToAddress("0x0003"),
		Position: 3,
		Type:     common.RoleNil,
		//	Stock:      0,
		NodeNumber: 0,
	})

	return list
}

func TestTopologyGraph_sort(t *testing.T) {
	graph := &TopologyGraph{
		NodeList:      getTestNodeList(),
		CurNodeNumber: 0,
	}
	graph.sort()
}

func TestTopologyGraph_Transfer2NextGraph(t *testing.T) {
	type fields struct {
		Number        uint64
		NodeList      []TopologyNodeInfo
		ElectList     []TopologyNodeInfo
		CurNodeNumber uint8
	}
	type args struct {
		number        uint64
		blockTopology *common.NetTopology
		electList     []common.Elect
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *TopologyGraph
		wantErr bool
	}{
		{
			name: "test1",
			fields: fields{
				Number:        87,
				NodeList:      getTestNodeList(),
				ElectList:     getTestNodeList(),
				CurNodeNumber: 0,
			},
			args: args{
				number: 88,
				blockTopology: &common.NetTopology{
					Type: common.NetTopoTypeChange,
					NetTopologyData: []common.NetTopologyData{
						{
							Account:  common.Address{},
							Position: 2,
						},
					},
				},
				electList: make([]common.Elect, 0),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//
			//got, err := self.Transfer2NextGraph(tt.args.number, tt.args.blockTopology, tt.args.electList)
			//	if (err != nil) != tt.wantErr {
			//	t.Errorf("TopologyGraph.Transfer2NextGraph() error = %v, wantErr %v", err, tt.wantErr)
			//	return
			//}
			// !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("TopologyGraph.Transfer2NextGraph() = %v, want %v", got, tt.want)
			//}
		})
	}
}

type A struct {
	A *big.Int
	B *big.Int
}

func Test111(t *testing.T) {
	a := A{
		A: big.NewInt(int64(100)),
		B: big.NewInt(int64(200)),
	}
	fmt.Println("a", a)
	a.A = a.B
	a.B = big.NewInt(int64(300))
	fmt.Println("a", a)
}
