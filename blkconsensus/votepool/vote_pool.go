//1542510702.5612884
//1542510682.1460223
//1542510630.8518646
//1542510600.8093905
//1542510521.805308
//1542510495.1412349
//1542510414.469612
//1542509541.30057
//1542509515.7504735
//1542509494.6583369
//1542509464.1425536
//1542508625.8268425
//1542507817.6679683
//1542507180.9524682
//1542506373.815198
//1542505707.6406555
//1542504841.544914
//1542503938.462168
// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package votepool

import (
	"container/list"
	"github.com/matrix/go-matrix/params/man"
	"time"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/crypto"
	"github.com/matrix/go-matrix/log"
	"github.com/pkg/errors"

	"sync"
)

type voteInfo struct {
	time        int64               // timestamp
	sign        common.VerifiedSign // signature
	fromAccount common.Address      // source of votes
	signHash    common.Hash         // hash of signature related message
}

// safe vote pool
type VotePool struct {
	// cache struct: map <from, map <msgHash, *data> >
	voteMap               map[common.Address]map[common.Hash]*voteInfo // vote cache
	timeIndex             *list.List                                   // time index used to delete obsolete data
	timeoutInterval       int64                                        // timeout
	AccountVoteCountLimit int                                          // vote limit per account
	legalRole             common.RoleType                              // roles that are legal
	logInfo               string
	mu                    sync.RWMutex
}

func NewVotePool(legalRole common.RoleType, logInfo string) *VotePool {
	return &VotePool{
		voteMap:               make(map[common.Address]map[common.Hash]*voteInfo),
		timeIndex:             list.New(),
		timeoutInterval:       int64(man.VotePoolTimeout),
		AccountVoteCountLimit: man.VotePoolCountLimit,
		legalRole:             legalRole,
		logInfo:               logInfo,
	}
}

func (vp *VotePool) AddVote(signHash common.Hash, sign common.Signature, fromAccount common.Address, height uint64, verifyFrom bool) error {
	signAccount, validate, err := crypto.VerifySignWithValidate(signHash.Bytes(), sign.Bytes())
	if err != nil {
		return err
	}

	if verifyFrom && signAccount.Equal(fromAccount) == false {
		return errors.Errorf("vote sign account[%s] != from account[%s]", signAccount.Hex(), fromAccount.Hex())
	}

	//todo 暂时关闭，经常因为高度获取不到 CA 导致丢票
	/*fromInfo, err := ca.GetAccountTopologyInfo(fromAccount, height-1)
	if err != nil {
		return fmt.Errorf("vote from node(%s) get role err(%s)", fromAccount.Hex(), err)
	}

	if fromInfo.Type != vp.legalRole {
		return fmt.Errorf("vote from node  role (%s) illegal! Legal role is (%s)", fromInfo.Type.String(), vp.legalRole.String())
	}*/

	vp.mu.Lock()
	defer vp.mu.Unlock()

	vote := &voteInfo{
		time:        time.Now().UnixNano() / 1000000,
		sign:        common.VerifiedSign{Sign: sign, Account: signAccount, Validate: validate, Stock: 0},
		fromAccount: signAccount,
		signHash:    signHash,
	}

	if err := vp.addVoteToMap(vote); err != nil {
		return err
	}

	vp.fixPoolByTimeout(vote.time)
	vp.fixPoolByCountLimit(vote.fromAccount, vote.time)

	return nil
}

func (vp *VotePool) GetVotes(signHash common.Hash) (signs []*common.VerifiedSign) {
	vp.mu.RLock()
	defer vp.mu.RUnlock()

	for _, accountVoteMap := range vp.voteMap {
		for key, value := range accountVoteMap {
			if signHash.Equal(key) {
				signs = append(signs, &value.sign)
			}
		}
	}
	return
}

func (vp *VotePool) DelVotes(signHash common.Hash) {
	if (signHash == common.Hash{}) {
		return
	}

	vp.mu.Lock()
	defer vp.mu.Unlock()

	for fromAccount, accountVoteMap := range vp.voteMap {
		for key := range accountVoteMap {
			if signHash.Equal(key) {
				if len(accountVoteMap) == 1 {
					delete(vp.voteMap, fromAccount)
				}
				delete(accountVoteMap, key)
			}
		}
	}
}

func (vp *VotePool) Clear() {
	vp.mu.Lock()
	defer vp.mu.Unlock()

	vp.timeIndex.Init()
	vp.voteMap = make(map[common.Address]map[common.Hash]*voteInfo)
}

func (vp *VotePool) addVoteToMap(vote *voteInfo) error {
	accountVoteMap, OK := vp.voteMap[vote.fromAccount]
	if !OK {
		accountVoteMap = make(map[common.Hash]*voteInfo)
		vp.voteMap[vote.fromAccount] = accountVoteMap
	}

	_, exist := accountVoteMap[vote.signHash]
	if exist {
		log.ERROR(vp.logInfo, "添加票池失败,已存在票 hash", vote.signHash.TerminalString(), "from", vote.fromAccount.Hex())
		return errors.Errorf("Vote is already exist")
	}

	accountVoteMap[vote.signHash] = vote
	vp.timeIndex.PushBack(vote)

	log.INFO(vp.logInfo, "加入票池成功 from", vote.fromAccount.Hex(), "sighHash", vote.signHash, "from总票数", len(accountVoteMap))

	return nil
}

func (vp *VotePool) fixPoolByTimeout(curTime int64) {
	deadLine := curTime - vp.timeoutInterval
	for {
		e := vp.timeIndex.Front()
		if nil == e {
			return
		}

		vote, OK := e.Value.(*voteInfo)
		if !OK {
			vp.timeIndex.Remove(e)
			log.WARN(vp.logInfo, "VotePool Data conversion error!", e.Value)
			continue
		}

		// whether there is no timeout vote now
		if vote.time >= deadLine {
			return
		}

		accountVoteMap, OK := vp.voteMap[vote.fromAccount]
		if OK {
			beforeLen := len(accountVoteMap)
			delete(accountVoteMap, vote.signHash)
			afterLen := len(accountVoteMap)

			if beforeLen != afterLen {
				log.INFO(vp.logInfo, "超时删除投票 hash", vote.signHash.TerminalString(),
					"from", vote.fromAccount.Hex(), "times", (curTime-vote.time)/1000, "删前数量", beforeLen, "删后数量", afterLen)
				if afterLen == 0 {
					delete(vp.voteMap, vote.fromAccount)
				}
			}
		}
		vp.timeIndex.Remove(e)
	}
}

func (vp *VotePool) fixPoolByCountLimit(fromAccount common.Address, curTime int64) {
	accountVoteMap, OK := vp.voteMap[fromAccount]
	if !OK {
		return
	}

	for {
		if len(accountVoteMap) <= vp.AccountVoteCountLimit {
			break
		}

		var earliest *voteInfo = nil
		for _, value := range accountVoteMap {
			if earliest == nil {
				earliest = value
			} else {
				if earliest.time > value.time {
					earliest = value
				}
			}
		}

		log.INFO(vp.logInfo, "数量删除投票 hash", earliest.signHash.TerminalString(),
			"from", earliest.fromAccount.Hex(), "times", (curTime-earliest.time)/1000, "总数量", len(accountVoteMap))

		delete(accountVoteMap, earliest.signHash)
	}
}
