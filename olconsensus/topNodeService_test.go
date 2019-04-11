// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package olconsensus

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/MatrixAINetwork/go-matrix/accounts/keystore"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/consensus/mtxdpos"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/crypto"
	"github.com/MatrixAINetwork/go-matrix/event"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	//"github.com/pborman/uuid"
	"math/rand"

	"github.com/MatrixAINetwork/go-matrix/consensus"
	//"unicode"
	"github.com/MatrixAINetwork/go-matrix/messageState"
)

var (
	testServs  []testNodeService
	fullstate  = []uint8{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	offState   = []uint8{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0}
	dposStocks = make(map[common.Address]uint16)
	nodeInfo   = make([]NodeOnLineInfo, 11)
)

type testDPOSEngine struct {
	dops   *mtxdpos.MtxDPOS
	reader consensus.StateReader
}

func (tsdpos *testDPOSEngine) VerifyBlock(reader consensus.StateReader, header *types.Header) error {
	return tsdpos.dops.VerifyBlock(reader, header)
}

func (tsdpos *testDPOSEngine) VerifyBlocks(reader consensus.StateReader, headers []*types.Header) error {
	return nil
}

//verify hash in current block
func (tsdpos *testDPOSEngine) VerifyHash(reader consensus.StateReader, signHash common.Hash, signs []common.Signature) ([]common.Signature, error) {
	return tsdpos.dops.VerifyHash(reader, signHash, signs)
}

//verify hash in given number block
func (tsdpos *testDPOSEngine) VerifyHashWithBlock(reader consensus.StateReader, signHash common.Hash, signs []common.Signature, blockHash common.Hash) ([]common.Signature, error) {
	return tsdpos.dops.VerifyHashWithBlock(reader, signHash, signs, blockHash)
}

//VerifyHashWithStocks(signHash common.Hash, signs []common.Signature, stocks map[common.Address]uint16) ([]common.Signature, error)

func (tsdpos *testDPOSEngine) VerifyHashWithVerifiedSigns(reader consensus.StateReader, signs []*common.VerifiedSign) ([]common.Signature, error) {
	return tsdpos.dops.VerifyHashWithVerifiedSigns(reader, signs)
}

func (tsdpos *testDPOSEngine) VerifyHashWithVerifiedSignsAndBlock(reader consensus.StateReader, signs []*common.VerifiedSign, blockHash common.Hash) ([]common.Signature, error) {
	return tsdpos.dops.VerifyHashWithVerifiedSignsAndBlock(reader, signs, blockHash)
}

func (tsdpos *testDPOSEngine) VerifyStocksWithBlock(reader consensus.StateReader, validators []common.Address, blockHash common.Hash) bool {
	return true
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
		//randNumber := rand.Intn(1000)
		randNumber := 0
		select {
		case data := <-serv.msgChan:
			//fmt.Printf("Sleep %d Millisecond\n", randNumber)
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
	testDpos := testDPOSEngine{dops: mtxdpos.NewMtxDPOS()}
	testServ := NewTopNodeService(&testDpos)
	testServ.topNodeState = testInfo
	testServ.validatorSign = testInfo
	testServ.msgSender = testInfo
	testServ.msgCenter = newCenter()
	testServ.Start()
	//todo: add fake validatorReader

	return testServ

}
func newTestServer() []testNodeService {

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
	return testServs
}
func setLeader(index int, number uint64, turn uint8) {
	serv := testServs[index]
	leader := mc.LeaderChangeNotify{
		ConsensusState: true,
		Leader:         serv.testInfo.self.Address,
		Number:         number,
		ReelectTurn:    uint32(turn),
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

func TestLeaderChangeNotifyHandler(t *testing.T) {
	log.InitLog(3)

	t.Run("leaderIsNil", testLeaderIsNil)
	t.Run("testLeaderIsNotSelf", testLeaderIsNotSelf)
	t.Run("testLeaderIsSelf", testLeaderIsSelf)
}

func testLeaderIsNil(t *testing.T) {
	testService := testNodeService{
		msgChan: make(chan interface{}, 10),
	}
	testService.testInfo = newTestNodeState(1)
	testService.TN = newTestNodeService(testService.testInfo, 1)
	leader := common.BytesToAddress([]byte(""))
	testService.TN.LeaderChangeNotifyHandler(leader)

}

func testLeaderIsNotSelf(t *testing.T) {
	testService := testNodeService{
		msgChan: make(chan interface{}, 10),
	}
	testService.testInfo = newTestNodeState(1)
	testService.TN = newTestNodeService(testService.testInfo, 1)
	leader := common.BytesToAddress([]byte("leader"))
	testService.TN.LeaderChangeNotifyHandler(leader)

}

func testLeaderIsSelf(t *testing.T) {
	testService := newTestServer()
	testService[0].testInfo = newTestNodeState(0)
	testService[0].TN = newTestNodeService(testService[0].testInfo, 0)
	setLeader(0, 1, 0)
	leader := testService[0].testInfo.self.Address
	testService[0].TN.LeaderChangeNotifyHandler(leader)

}

func TestConsensusReqMsgHandler(t *testing.T) {
	log.InitLog(3)

	t.Run("pointerOfRequestIsNil", TestPointerOfRequestIsNil)
	t.Run("RequestIsNil", TestRequestIsNil)
	t.Run("RequestIsNotNil", TestRequestIsNotNil)
	t.Run("ReqIsNil", TestReqIsNil)
}

func TestPointerOfRequestIsNil(t *testing.T) {
	log.InitLog(3)

	testService := newTestServer()
	testService[0].testInfo = newTestNodeState(0)
	testService[0].TN = newTestNodeService(testService[0].testInfo, 0)
	setLeader(0, 1, 0)

	testService[0].TN.consensusReqMsgHandler(nil)

}

func TestRequestIsNil(t *testing.T) {
	log.InitLog(3)

	testService := newTestServer()
	testService[0].testInfo = newTestNodeState(0)
	testService[0].TN = newTestNodeService(testService[0].testInfo, 0)
	setLeader(0, 1, 0)

	requests := make([]*mc.OnlineConsensusReq, 0)
	testService[0].TN.consensusReqMsgHandler(requests)

}

func TestRequestIsNotNil(t *testing.T) {
	log.InitLog(3)

	testService := newTestServer()
	testService[0].testInfo = newTestNodeState(0)
	testService[0].TN = newTestNodeService(testService[0].testInfo, 0)
	setLeader(0, 1, 0)
	req := &mc.OnlineConsensusReq{
		Leader:      testService[0].testInfo.self.Address,
		Seq:         1,
		Node:        common.Address{},
		OnlineState: onLine,
	}
	requests := make([]*mc.OnlineConsensusReq, 0)
	requests = append(requests, req)
	testService[0].TN.consensusReqMsgHandler(requests)

}

func TestReqIsNil(t *testing.T) {
	log.InitLog(3)

	testService := newTestServer()
	testService[0].testInfo = newTestNodeState(0)
	testService[0].TN = newTestNodeService(testService[0].testInfo, 0)
	setLeader(0, 1, 0)
	req := &mc.OnlineConsensusReq{}
	requests := make([]*mc.OnlineConsensusReq, 0)
	requests = append(requests, req)
	testService[0].TN.consensusReqMsgHandler(requests)

}

func TestConsensusVoteMsgHandler(t *testing.T) {
	log.InitLog(3)

	t.Run("pointerOfMsgtIsNil", TestPointerOfMsgIsNil)
	t.Run("msgIsNil", TestMsgIsNil)
	t.Run("msgIsNotNil", TestMsgIsNotNil)
	t.Run("voteIsNil", TestVoteIsNil)
}

func TestPointerOfMsgIsNil(t *testing.T) {
	log.InitLog(3)

	testService := newTestServer()
	testService[0].testInfo = newTestNodeState(0)
	testService[0].TN = newTestNodeService(testService[0].testInfo, 0)
	setLeader(0, 1, 0)

	testService[0].TN.consensusVoteMsgHandler(nil)
}

func TestMsgIsNil(t *testing.T) {
	log.InitLog(3)

	testService := newTestServer()
	testService[0].testInfo = newTestNodeState(0)
	testService[0].TN = newTestNodeService(testService[0].testInfo, 0)
	setLeader(0, 1, 0)
	msg := make([]mc.HD_ConsensusVote, 0)
	testService[0].TN.consensusVoteMsgHandler(msg)

}

func TestMsgIsNotNil(t *testing.T) {
	log.InitLog(3)

	testService := newTestServer()
	testService[0].testInfo = newTestNodeState(0)
	testService[0].TN = newTestNodeService(testService[0].testInfo, 0)
	setLeader(0, 1, 0)
	consensusVote := mc.HD_ConsensusVote{
		SignHash: common.Hash{},
		Round:    1,
		Sign:     common.Signature{},
		From:     testService[0].testInfo.self.Address,
	}
	msg := make([]mc.HD_ConsensusVote, 0)
	msg = append(msg, consensusVote)
	testService[0].TN.consensusVoteMsgHandler(msg)

}

func TestVoteIsNil(t *testing.T) {
	log.InitLog(3)

	testService := newTestServer()
	testService[0].testInfo = newTestNodeState(0)
	testService[0].TN = newTestNodeService(testService[0].testInfo, 0)
	setLeader(0, 1, 0)
	consensusVote := mc.HD_ConsensusVote{}
	msg := make([]mc.HD_ConsensusVote, 0)
	msg = append(msg, consensusVote)
	testService[0].TN.consensusVoteMsgHandler(msg)

}

func TestOnlineConsensusVoteResultMsgHandler(t *testing.T) {
	log.InitLog(3)
	t.Run("TestPointerOfVoteResultMsgIsNil", TestPointerOfVoteResultMsgIsNil)
	t.Run("VoteResultmsgIsNil", TestVoteResultMsgIsNil)
	t.Run("TestSignListOfVoteResultIsBlank", TestSignListOfVoteResultIsBlank)
	t.Run("VoteResultmsgIsNotNil", TestVoteResultMsgIsNotNil)
	t.Run("voteIsNil", TestVoteIsNil)

}

func TestPointerOfVoteResultMsgIsNil(t *testing.T) {
	log.InitLog(3)

	testService := newTestServer()
	testService[0].testInfo = newTestNodeState(0)
	testService[0].TN = newTestNodeService(testService[0].testInfo, 0)
	setLeader(0, 1, 0)

	testService[0].TN.OnlineConsensusVoteResultMsgHandler(nil)
}

func TestVoteResultMsgIsNil(t *testing.T) {
	log.InitLog(3)

	testService := newTestServer()
	testService[0].testInfo = newTestNodeState(0)
	testService[0].TN = newTestNodeService(testService[0].testInfo, 0)
	setLeader(0, 1, 0)
	msg := new(mc.HD_OnlineConsensusVoteResultMsg)
	fmt.Println("msg", msg)
	testService[0].TN.OnlineConsensusVoteResultMsgHandler(msg)

}

func TestSignListOfVoteResultIsBlank(t *testing.T) {
	log.InitLog(3)

	testService := newTestServer()
	testService[0].testInfo = newTestNodeState(0)
	testService[0].TN = newTestNodeService(testService[0].testInfo, 0)
	setLeader(0, 1, 0)
	voteResultMsg := mc.HD_OnlineConsensusVoteResultMsg{
		Req: &mc.OnlineConsensusReq{
			Leader:      testService[0].testInfo.self.Address,
			Seq:         1,
			Node:        testService[1].testInfo.self.Address,
			OnlineState: 1,
		},
		SignList: make([]common.Signature, 1),
		From:     testService[0].testInfo.self.Address,
	}

	testService[0].TN.OnlineConsensusVoteResultMsgHandler(&voteResultMsg)

}

func TestVoteResultMsgIsNotNil(t *testing.T) {
	log.InitLog(3)

	testService := newTestServer()
	testService[0].testInfo = newTestNodeState(0)
	testService[0].TN = newTestNodeService(testService[0].testInfo, 0)
	setLeader(0, 1, 0)
	voteResultMsg := mc.HD_OnlineConsensusVoteResultMsg{
		Req: &mc.OnlineConsensusReq{
			Leader:      testService[0].testInfo.self.Address,
			Seq:         1,
			Node:        testService[1].testInfo.self.Address,
			OnlineState: 1,
		},
		SignList: make([]common.Signature, 1),
		From:     testService[0].testInfo.self.Address,
	}
	sign := common.BytesToSignature([]byte("signlist"))
	voteResultMsg.SignList = append(voteResultMsg.SignList, sign)

	testService[0].TN.OnlineConsensusVoteResultMsgHandler(&voteResultMsg)

}

func TestCheckPosVoteResults(t *testing.T) {
	t.Run("TestProposeIsNil", TestProposeIsNil)
	t.Run("TestVotesIsNil", TestVotesIsNil)
	t.Run("TestSighHashOfVotesIsNil", TestSighHashOfVotesIsNil)
	t.Run("TestRountOfVotesIsZero", TestRountOfVotesIsZero)
	t.Run("TestSighOfVotesIsBlank", TestSighOfVotesIsBlank)

}

func TestProposeIsNil(t *testing.T) {
	log.InitLog(3)

	testService := newTestServer()
	testService[0].testInfo = newTestNodeState(0)
	testService[0].TN = newTestNodeService(testService[0].testInfo, 0)
	setLeader(0, 1, 0)
	vote := &mc.HD_ConsensusVote{
		SignHash: common.BytesToHash([]byte("hash")),
		Round:    1,
		Sign:     common.BytesToSignature([]byte("sigh")),
		From:     testService[0].testInfo.self.Address,
	}
	insVote := voteInfo{messageState.RlpFnvHash(vote), vote}
	votes := make([]voteInfo, 0)
	votes = append(votes, insVote)
	testService[0].TN.checkPosVoteResults(nil, votes)
}

func TestVotesIsNil(t *testing.T) {
	log.InitLog(3)

	testService := newTestServer()
	testService[0].testInfo = newTestNodeState(0)
	testService[0].TN = newTestNodeService(testService[0].testInfo, 0)
	setLeader(0, 1, 0)
	testService[0].TN.checkPosVoteResults(nil, nil)
}

func TestSighHashOfVotesIsNil(t *testing.T) {
	log.InitLog(3)

	testService := newTestServer()
	testService[0].testInfo = newTestNodeState(0)
	testService[0].TN = newTestNodeService(testService[0].testInfo, 0)
	setLeader(0, 1, 0)
	vote := &mc.HD_ConsensusVote{
		SignHash: common.Hash{},
		Round:    1,
		Sign:     common.BytesToSignature([]byte("sigh")),
		From:     testService[0].testInfo.self.Address,
	}
	insVote := voteInfo{messageState.RlpFnvHash(vote), vote}
	votes := make([]voteInfo, 0)
	votes = append(votes, insVote)
	testService[0].TN.checkPosVoteResults(nil, votes)
}
func TestRountOfVotesIsZero(t *testing.T) {
	log.InitLog(3)

	testService := newTestServer()
	testService[0].testInfo = newTestNodeState(0)
	testService[0].TN = newTestNodeService(testService[0].testInfo, 0)
	setLeader(0, 1, 0)
	vote := &mc.HD_ConsensusVote{
		SignHash: common.BytesToHash([]byte("hash")),
		Round:    0,
		Sign:     common.BytesToSignature([]byte("sigh")),
		From:     testService[0].testInfo.self.Address,
	}
	insVote := voteInfo{messageState.RlpFnvHash(vote), vote}
	votes := make([]voteInfo, 0)
	votes = append(votes, insVote)
	testService[0].TN.checkPosVoteResults(nil, votes)
}

func TestSighOfVotesIsBlank(t *testing.T) {
	log.InitLog(3)

	testService := newTestServer()
	testService[0].testInfo = newTestNodeState(0)
	testService[0].TN = newTestNodeService(testService[0].testInfo, 0)
	setLeader(0, 1, 0)
	vote := &mc.HD_ConsensusVote{
		SignHash: common.BytesToHash([]byte("hash")),
		Round:    1,
		Sign:     common.Signature{},
		From:     testService[0].testInfo.self.Address,
	}
	insVote := voteInfo{messageState.RlpFnvHash(vote), vote}
	votes := make([]voteInfo, 0)
	votes = append(votes, insVote)
	testService[0].TN.checkPosVoteResults(nil, votes)
}
