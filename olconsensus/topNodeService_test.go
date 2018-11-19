// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package olconsensus

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/matrix/go-matrix/accounts/keystore"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/consensus/mtxdpos"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/crypto"
	"github.com/matrix/go-matrix/event"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/pborman/uuid"
	"math/rand"
	"github.com/matrix/go-matrix/consensus"
)

var (
	testServs  []testNodeService
	fullstate  = []uint8{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	offState   = []uint8{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0}
	dposStocks = make(map[common.Address]uint16)
	nodeInfo   = make([]NodeOnLineInfo, 11)
)

type testDPOSEngine struct {
	dops *mtxdpos.MtxDPOS
	reader consensus.ValidatorReader
}

func (tsdpos *testDPOSEngine) VerifyBlock(header *types.Header) error {
	return tsdpos.dops.VerifyBlock(, header)
}

func (tsdpos *testDPOSEngine) VerifyBlocks(headers []*types.Header) error {
	return nil
}

//verify hash in current block
func (tsdpos *testDPOSEngine) VerifyHash(signHash common.Hash, signs []common.Signature) ([]common.Signature, error) {
	return tsdpos.dops.VerifyHash(signHash, signs)
}

//verify hash in given number block
func (tsdpos *testDPOSEngine) VerifyHashWithNumber(signHash common.Hash, signs []common.Signature, number uint64) ([]common.Signature, error) {
	return tsdpos.dops.VerifyHashWithStocks(signHash, signs, dposStocks)
}

//VerifyHashWithStocks(signHash common.Hash, signs []common.Signature, stocks map[common.Address]uint16) ([]common.Signature, error)

func (tsdpos *testDPOSEngine) VerifyHashWithVerifiedSigns(signs []*common.VerifiedSign) ([]common.Signature, error) {
	return tsdpos.dops.VerifyHashWithVerifiedSigns(signs)
}

func (tsdpos *testDPOSEngine) VerifyHashWithVerifiedSignsAndNumber(signs []*common.VerifiedSign, number uint64) ([]common.Signature, error) {
	return tsdpos.dops.VerifyHashWithVerifiedSignsAndNumber(signs, number)
}

type Center struct {
	FeedMap map[mc.EventCode]*event.Feed
}

func newCenter() *Center {
	msgCenter := &Center{FeedMap: make(map[mc.EventCode]*event.Feed)}
	for i := 0; i < int(mc.LastEventCode); i++ {
		msgCenter.FeedMap[mc.EventCode(i)] = new(event.Feed)
	}
	return msgCenter
}
func (cen *Center) SubscribeEvent(aim mc.EventCode, ch interface{}) (event.Subscription, error) {
	feed, ok := cen.FeedMap[aim]
	if !ok {
		return nil, mc.SubErrorNoThisEvent
	}
	return feed.Subscribe(ch), nil
}

func (cen *Center) PublishEvent(aim mc.EventCode, data interface{}) error {
	feed, ok := cen.FeedMap[aim]
	if !ok {
		return mc.PostErrorNoThisEvent
	}
	feed.Send(data)
	return nil
}

type testNodeState struct {
	self         keystore.Key
	electNode    map[common.Address]OnlineState //
	onlineNode   []common.Address               //
	offlineNode  []common.Address               //
	consensusOn  []common.Address
	consensusOff []common.Address
}

func newTestNodeState(id int) *testNodeState {
	key, _ := crypto.GenerateKey()
	//id := uuid.NewRandom()
	keystore := keystore.Key{
		//Id:         id,
		Address:    crypto.PubkeyToAddress(key.PublicKey),
		PrivateKey: key,
	}
	fmt.Println("id", id, "address", keystore.Address.String())
	electNode := make(map[common.Address]OnlineState, 0)
	online := make([]common.Address, 0)
	offline := make([]common.Address, 0)
	consensusOn := make([]common.Address, 0)
	consensusOff := make([]common.Address, 0)
	return &testNodeState{keystore, electNode, online, offline, consensusOn, consensusOff}
}
func (ts *testNodeState) GetTopNodeOnlineState() []NodeOnLineInfo {

	return nodeInfo
}
func (ts *testNodeState) SendNodeMsg(subCode mc.EventCode, msg interface{}, Roles common.RoleType, address []common.Address) {
	switch msg.(type) {
	case *mc.HD_OnlineConsensusReqs:
		data := msg.(*mc.HD_OnlineConsensusReqs)
		for i := 0; i < len(data.ReqList); i++ {
			data.ReqList[i].Leader = ts.self.Address
		}
		for _, serv := range testServs {
			serv.msgChan <- msg
		}

		//		serv.TN.msgCenter.PublishEvent(mc.HD_TopNodeConsensusReq,data.(*mc.OnlineConsensusReqs))
	case *mc.HD_OnlineConsensusVotes:
		data := msg.(*mc.HD_OnlineConsensusVotes)
		for i := 0; i < len(data.Votes); i++ {
			data.Votes[i].From = ts.self.Address
		}
		for _, serv := range testServs {
			serv.msgChan <- msg
		}

		//		testServs[1].msgChan <-msg
		//		serv.TN.msgCenter.PublishEvent(mc.HD_TopNodeConsensusVote,data.(*mc.HD_OnlineConsensusVotes))
	case *mc.HD_OnlineConsensusVoteResultMsg:
		data := msg.(*mc.HD_OnlineConsensusVoteResultMsg)
		for i := 0; i < len(data.SignList); i++ {
			data.From = ts.self.Address
		}
		for _, serv := range testServs {
			serv.msgChan <- msg
		}

	default:
		for _, serv := range testServs {
			serv.msgChan <- msg
		}
		//		log.Error("Type Error","type",reflect.TypeOf(data))
	}
	//	for _,serv := range testServs{
	//		serv.msgChan <-msg
	//	}
}

func (ts *testNodeState) SignWithValidate(hash []byte, validate bool) (common.Signature, error) {
	sigByte, err := crypto.SignWithValidate(hash, validate, ts.self.PrivateKey)
	if err != nil {
		return common.Signature{}, err
	}
	return common.BytesToSignature(sigByte), nil
}

func (ts *testNodeState) IsSelfAddress(addr common.Address) bool {
	return ts.self.Address == addr
}

type testNodeService struct {
	TN       *TopNodeService
	msgChan  chan interface{}
	testInfo *testNodeState
}

func (serv *testNodeService) getMessageLoop() {
	for {
		rand.Seed(time.Now().UnixNano())
		randNumber := rand.Intn(1000)
		select {
		case data := <-serv.msgChan:
			fmt.Printf("Sleep %d Millisecond\n", randNumber)
			time.Sleep(time.Duration(randNumber) * time.Millisecond)

			switch data.(type) {
			case *mc.RoleUpdatedMsg:
				serv.TN.msgCenter.PublishEvent(mc.CA_RoleUpdated, data.(*mc.RoleUpdatedMsg))
			case *mc.LeaderChangeNotify:
				serv.TN.msgCenter.PublishEvent(mc.Leader_LeaderChangeNotify, data.(*mc.LeaderChangeNotify))
			case *mc.HD_OnlineConsensusReqs:
				serv.TN.msgCenter.PublishEvent(mc.HD_TopNodeConsensusReq, data.(*mc.HD_OnlineConsensusReqs))
			case *mc.HD_OnlineConsensusVotes:
				serv.TN.msgCenter.PublishEvent(mc.HD_TopNodeConsensusVote, data.(*mc.HD_OnlineConsensusVotes))
			case *mc.HD_OnlineConsensusVoteResultMsg:
				serv.TN.msgCenter.PublishEvent(mc.HD_TopNodeConsensusVoteResult, data.(*mc.HD_OnlineConsensusVoteResultMsg))
			default:
				log.Error("Type Error", "type", reflect.TypeOf(data))
			}
		}
	}
}
func newTestNodeService(testInfo *testNodeState, id int) *TopNodeService {
	testDpos := testDPOSEngine{dops: mtxdpos.NewMtxDPOS(nil)}
	testServ := NewTopNodeService(&testDpos, id)
	testServ.topNodeState = testInfo
	testServ.validatorSign = testInfo
	testServ.msgSender = testInfo
	testServ.msgCenter = newCenter()

	testServ.Start()

	return testServ

}
func newTestServer() {

	testServs = make([]testNodeService, 11)
	nodes := make([]common.Address, 11)
	for i := 0; i < 11; i++ {
		testServs[i].msgChan = make(chan interface{}, 10)
		testServs[i].testInfo = newTestNodeState(i)
		testServs[i].TN = newTestNodeService(testServs[i].testInfo, i)
		nodes[i] = testServs[i].testInfo.self.Address
		dposStocks[nodes[i]] = 1
		go testServs[i].getMessageLoop()
	}
	for i := 0; i < 11; i++ {
		testServs[i].TN.stateMap.setElectNodes(nodes, 10)
	}
	for i := 0; i < 11; i++ {
		nodeInfo[i].Address = testServs[i].testInfo.self.Address
		if i == 9 {
			nodeInfo[i].OnlineState = offState
		} else {
			nodeInfo[i].OnlineState = fullstate
		}
	}

}
func setLeader(index int, number uint64, turn uint8) {
	serv := testServs[index]
	leader := mc.LeaderChangeNotify{
		ConsensusState: true,
		Leader:         serv.testInfo.self.Address,
		Number:         number,
		ReelectTurn:    turn,
	}
	roleChange := mc.RoleUpdatedMsg{
		Leader:   serv.testInfo.self.Address,
		BlockNum: number,
		Role:     common.RoleValidator,
	}
	serv.TN.msgSender.SendNodeMsg(mc.CA_RoleUpdated, &roleChange, common.RoleValidator, nil)

	serv.TN.msgSender.SendNodeMsg(mc.Leader_LeaderChangeNotify, &leader, common.RoleValidator, nil)
}
func TestNewTopNodeService(t *testing.T) {
	log.InitLog(1)
	newTestServer()
	go setLeader(0, 1, 0)
	time.Sleep(time.Second * 5)
	for i := 0; i < 11; i++ {
		t.Log(testServs[i].TN.stateMap.finishedProposal.DPosVoteS[0].Proposal)
	}

}

func setCurrentOnlineState(id int) {
	online := make([]common.Address, 0)
	elect := make([]common.Address, 0)
	for i := 0; i < 11; i++ {
		if i != id {
			online = append(online, nodeInfo[i].Address)
		}
		elect = append(elect, nodeInfo[i].Address)
		fmt.Println("len(online)", len(testServs[i].testInfo.onlineNode))
		testServs[i].TN.SetCurrentOnlineState(online, elect)
	}
}

func TestNewTopNodeServiceRound(t *testing.T) {
	log.InitLog(1)
	newTestServer()
	go func() {
		setLeader(0, 1, 0)
		time.Sleep(20 * time.Second)
		//nodeInfo[7].OnlineState = offState
		//nodeInfo[9].OnlineState = fullstate

		setLeader(1, 2, 0)
		//setCurrentOnlineState(9)
		//nodeInfo[9].OnlineState = fullstate
		time.Sleep(20 * time.Second)
		//setLeader(2, 3, 0)
		//time.Sleep(2 * time.Second)

	}()
	time.Sleep(time.Second * 40)
	//for i := 0; i < 11; i++ {
	//	t.Log(testServs[i].TN.stateMap.finishedProposal.DPosVoteS[0].Proposal)
	//	t.Log(testServs[i].TN.stateMap.finishedProposal.DPosVoteS[1].Proposal)
	//}

}
