// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package blkverify

import (
	"container/list"
	"time"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/pkg/errors"

	"github.com/MatrixAINetwork/go-matrix/params/manparams"
)

type voteInfo struct {
	time     int64 // 时间戳，收到的时间
	sign     common.Signature
	signHash common.Hash
	from     common.Address
}

type unverifiedVotePool struct {
	voteMap               map[common.Address]map[common.Hash]*voteInfo // 投票缓存
	timeIndex             *list.List                                   // 按投票到来先后的索引，用于删除过期数据
	timeoutInterval       int64                                        // 超时时间
	AccountVoteCountLimit int                                          // 每个用户的投票数量限制
	logInfo               string
}

func newUnverifiedVotePool(logInfo string) *unverifiedVotePool {
	return &unverifiedVotePool{
		voteMap:               make(map[common.Address]map[common.Hash]*voteInfo),
		timeIndex:             list.New(),
		timeoutInterval:       manparams.VotePoolTimeout,
		AccountVoteCountLimit: manparams.VotePoolCountLimit,
		logInfo:               logInfo,
	}
}

func (vp *unverifiedVotePool) AddVote(signHash common.Hash, sign common.Signature, from common.Address) error {
	if (signHash == common.Hash{}) || (sign == common.Signature{}) || (from == common.Address{}) {
		return ErrParamIsNil
	}

	vote := &voteInfo{
		time:     time.Now().UnixNano() / 1000000,
		sign:     sign,
		signHash: signHash,
		from:     from,
	}

	if err := vp.addVoteToMap(vote); err != nil {
		return err
	}

	vp.fixPoolByTimeout(vote.time)
	vp.fixPoolByCountLimit(from, vote.time)

	return nil
}

func (vp *unverifiedVotePool) GetVotes(signHash common.Hash) (votes []*voteInfo) {
	for _, accountVoteMap := range vp.voteMap {
		for key, value := range accountVoteMap {
			if signHash.Equal(key) {
				votes = append(votes, value)
			}
		}
	}
	return
}

func (vp *unverifiedVotePool) DelVotes(signHash common.Hash) {
	if (signHash == common.Hash{}) {
		return
	}

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

func (vp *unverifiedVotePool) Clear() {
	vp.timeIndex.Init()
	vp.voteMap = make(map[common.Address]map[common.Hash]*voteInfo)
}

func (vp *unverifiedVotePool) addVoteToMap(vote *voteInfo) error {
	accountVoteMap, OK := vp.voteMap[vote.from]
	if !OK {
		accountVoteMap = make(map[common.Hash]*voteInfo)
		vp.voteMap[vote.from] = accountVoteMap
	}

	_, exist := accountVoteMap[vote.signHash]
	if exist {
		//log.ERROR(vp.logInfo, "添加票池失败,已存在票 hash", signHash.TerminalString(), "from", vote.sign.Account.Hex())
		return errors.Errorf("Vote is already exist")
	}

	accountVoteMap[vote.signHash] = vote
	vp.timeIndex.PushBack(vote)

	//log.INFO(vp.logInfo, "加入票池成功 from", vote.fromAccount.Hex(), "sighHash", vote.signHash, "from总票数", len(accountVoteMap))

	return nil
}

func (vp *unverifiedVotePool) fixPoolByTimeout(curTime int64) {
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

		accountVoteMap, OK := vp.voteMap[vote.from]
		if OK {
			beforeLen := len(accountVoteMap)
			delete(accountVoteMap, vote.signHash)
			afterLen := len(accountVoteMap)

			if beforeLen != afterLen {
				//log.INFO(vp.logInfo, "超时删除投票 hash", vote.signHash.TerminalString(),
				//"from", vote.sign.Account.Hex(), "times", (curTime-vote.time)/1000, "删前数量", beforeLen, "删后数量", afterLen)
				if afterLen == 0 {
					delete(vp.voteMap, vote.from)
				}
			}
		}
		vp.timeIndex.Remove(e)
	}
}

func (vp *unverifiedVotePool) fixPoolByCountLimit(fromAccount common.Address, curTime int64) {
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

		//log.INFO(vp.logInfo, "数量删除投票 hash", earliest.signHash.TerminalString(),
		//	"from", earliest.sign.Account.Hex(), "times", (curTime-earliest.time)/1000, "总数量", len(accountVoteMap))

		delete(accountVoteMap, earliest.signHash)
	}
}
