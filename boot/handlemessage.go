// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package boot

import (
	"encoding/json"
	"time"

	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/p2p"
	"github.com/matrix/go-matrix/p2p/discover"
	"github.com/matrix/go-matrix/params"
)

const (
	//Getheightreq gmaneightreq_code
	Getheightreq = 0x0001
	//Getheightrsp gmaneightrsp_code
	Getheightrsp = 0x0002
	//Getmainnodereq getmainnodersp_code
	Getmainnodereq = 0x0003
	//Getmainnodersp getmainnodersp_code
	Getmainnodersp = 0x0004
	// Getpingreq get ping req_code
	Getpingreq = 0x0005
	// Getpongrsp get ping rsq_code
	Getpongrsp = 0x0006
	//P2PBootCode p2pboot_code
	P2PBootCode = 1
)

// GetHeightReq get height req
type GetHeightReq struct {
}

//GetHeightRsp get height rsp
type GetHeightRsp struct {
	BlockHeight uint64
}

//GetMainNodeReq get mainnode req
type GetMainNodeReq struct {
}

//GetMainNodeRsp get main node rsp
type GetMainNodeRsp struct {
	//MainNode []election.NodeInfo
}

//GetPingReq get ping req
type GetPingReq struct {
}

//GetPingRsp get ping rsp
type GetPingRsp struct {
	Msg uint64
}

//LocalHeightInfo local height info
type LocalHeightInfo struct {
	IP  string
	Len uint64
}

//LocalMainNodeList local mainnode list
type LocalMainNodeList struct {
	IP string
	//MainNodeList []election.NodeInfo
}

//LocalPongInfo local pong info
type LocalPongInfo struct {
	IP   string
	flag bool
}

//DeleteIndex delete a index in []
func (TBoot *Boots) DeleteIndex(index uint64) int {
	for i := 0; i < len(TBoot.NeedAck); i++ {
		if TBoot.NeedAck[i] == index {
			TBoot.NeedAck = append(TBoot.NeedAck[:i], TBoot.NeedAck[i+1:]...)
			return i
		}
	}
	return -1
}

//SendData send data from p2p
func SendData(ReadySend []p2p.Custsend) {
	//	log.INFO(Module, "Boot SendData data", ReadySend)
	p2p.SendChan <- ReadySend
}

func GetIPFromID(ID string) string {
	for _, url := range params.MainnetBootnodes {
		if bootNode, err := discover.ParseNode(url); err == nil {
			if bootNode.ID.String() == ID {
				return bootNode.IP.String()
			}
		}
	}
	return ""
}

//GetPingPong get specified ID pingpong
func (TBoot *Boots) GetPingPong(ListIDString []string) []LocalPongInfo {
	ReadySend := make([]p2p.Custsend, 0)
	MeNeedAck := make([]uint64, 0)
	for i := 0; i < len(ListIDString); i++ {
		tempsend := TBoot.MakeSendMsg("", GetIPFromID(ListIDString[i]), true, P2PBootCode, Getpingreq)
		ReadySend = append(ReadySend, tempsend)
		TBoot.NeedAck = append(TBoot.NeedAck, tempsend.Data.Seq)
		MeNeedAck = append(MeNeedAck, tempsend.Data.Seq)
	}
	go SendData(ReadySend)

	var ShouldReceive []LocalPongInfo
	for i := 0; i < len(MeNeedAck); i++ {
		select {
		case res := <-TBoot.ChanPing:
			//	log.INFO(Module, "Receive Pong msg", res)
			ShouldReceive = append(ShouldReceive, res)
		case <-time.After(TimeOutLimit):
			log.WARN(Module, "BOOT get ping pong overtime ", " ")

		}
	}
	for i := 0; i < len(MeNeedAck); i++ {
		TBoot.DeleteIndex(MeNeedAck[i])
	}
	return ShouldReceive

}

//MakeSendMsg make send msg to p2p
func (TBoot *Boots) MakeSendMsg(fromip string, toip string, istcp bool, code uint64, Type uint64) p2p.Custsend {

	var datastruct []byte
	switch Type {
	case Getheightreq:
		datastruct, _ = json.MarshalIndent(GetHeightReq{}, "", "   ")
	case Getmainnodereq:
		datastruct, _ = json.MarshalIndent(GetMainNodeReq{}, "", "   ")
	case Getpingreq:
		datastruct, _ = json.MarshalIndent(GetPingReq{}, "", "   ")
	default:
		log.WARN(Module, "Make_Send_Msg err Type:", Type)
	}
	TBoot.PublicSeq = TBoot.PublicSeq + 1
	return p2p.Custsend{
		FromIp: TBoot.LocalID,
		ToIp:   toip,
		IsTcp:  istcp,
		Code:   code,
		Data: p2p.Data_Format{
			Type:        Type,
			Seq:         TBoot.PublicSeq,
			Data_struct: datastruct,
		},
	}

}

//HandleGetPingPongReq handle get pingpong req
func (TBoot *Boots) HandleGetPingPongReq(AnalyData p2p.Custsend) {

	data, err := json.MarshalIndent(GetPingRsp{Msg: 1}, "", "   ")
	if err != nil {
		log.WARN(Module, "Boot:Json 0x0005 marshaling failed err", err)
	}
	AnalyData.Data.Data_struct = data
	AnalyData.Data.Type = Getpongrsp
	AnalyData.FromIp, AnalyData.ToIp = AnalyData.ToIp, AnalyData.FromIp
	SendMainNode := []p2p.Custsend{AnalyData}
	go SendData(SendMainNode)
}

//HandleGetPingPongRsp handle get pingpong rsp
func (TBoot *Boots) HandleGetPingPongRsp(AnalyData p2p.Custsend) {

	if TBoot.DeleteIndex(AnalyData.Data.Seq) == -1 {
		log.DEBUG(Module, "0x0006 overtime pingpong timeout", "0x0006")
		return
	}
	var realdata GetPingRsp
	if err := json.Unmarshal(AnalyData.Data.Data_struct, &realdata); err != nil {
		log.ERROR(Module, "0x0006 json Unmarshal failed data", AnalyData.Data.Data_struct)
	}
	TBoot.ChanPing <- LocalPongInfo{IP: AnalyData.FromIp, flag: true}
}

//HandleP2PMessage :read mine recv chan
func (TBoot *Boots) HandleP2PMessage() {
	for {
		select {
		case AnalyData := <-TBoot.MyRecvChan:
			//log.INFO(Module, "receive data", AnalyData)
			TBoot.HandleMessage[int(AnalyData.Data.Type)](AnalyData)
			/*		default:
					continue*/
		}
	}

}

//ReadRecvChanfromP2P  read recv chan from p2p
func (TBoot *Boots) ReadRecvChanfromP2P() {
	for {
		select {
		case data := <-p2p.RecvChan:
			TBoot.MyRecvChan <- data
			/*default:
			continue*/

		}
	}
}
