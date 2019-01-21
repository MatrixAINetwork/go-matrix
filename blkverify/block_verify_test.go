// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package blkverify

/*func TestRoleUpdatedMsg(t *testing.T) {
	blkVerify, err := NewBlockVerify(nil)
	if err != nil {
		t.Fatalf("StartServer failed: %v\n", err)
	}

	roles := make([]common.RoleType, 0)
	roles = append(roles, common.RoleValidator)
	roles = append(roles, common.RoleBroadcast)
	roles = append(roles, common.RoleValidator)
	roles = append(roles, common.RoleValidator)
	roles = append(roles, common.RoleMiner)
	roles = append(roles, common.RoleValidator)
	roles = append(roles, common.RoleMiner)
	number := uint64(20)
	for _, role := range roles {
		number++
		msg := mc.RoleUpdatedMsg{Role: role, BlockNum: number, Leader: common.Address{}}
		blkVerify.handleRoleUpdatedMsg(&msg)
		t.Logf("msg:role(%d) number(%d), after handle msg server role(%d) number(%d)", role, number, blkVerify.curRole, blkVerify.curNumber)
		p := blkVerify.processManage.GetCurrentProcess()
		if role != common.RoleValidator && p != nil {
			t.Fatalf("role is not validator, but current process is not nil!")
		}

		if role == common.RoleValidator && p == nil {
			t.Fatalf("role is validator, but current process is nil!")
		}
	}
}

func TestLeaderChangeMsg_01(t *testing.T) {
	blkVerify, err := NewBlockVerify(nil)
	if err != nil {
		t.Fatalf("StartServer failed: %v\n", err)
	}

	number := uint64(29)
	blkVerify.handleRoleUpdatedMsg(&mc.RoleUpdatedMsg{Role: common.RoleBroadcast, BlockNum: number, Leader: common.Address{}})
	t.Logf("server curRole(%d) curNumber(%d)", blkVerify.curRole, blkVerify.curNumber)

	leaderMsg := &mc.LeaderChangeNotify{
		ConsensusState: true,
		Leader:         common.HexToAddress("ABCDEF"),
		Number:         number + 1,
		ReelectTurn:    0,
	}

	blkVerify.handleLeaderChangeNotify(leaderMsg)
	cp := blkVerify.processManage.GetCurrentProcess()
	if cp != nil {
		t.Fatalf("role is not validator, but after handle leader change, there is running process")
	}

	p, OK := blkVerify.processManage.processMap[number]
	if OK == false {
		t.Fatalf("handle msg err, there is no process create!")
	}

	t.Logf("process leader = %v", p.leader.String())
	if p.leader != leaderMsg.Leader {
		t.Fatalf("process leader err! %v != %v", p.leader, leaderMsg.Leader)
	}

	if p.state != StateIdle {
		t.Fatalf("process state err, state = %d", p.state)
	}
}

func TestLeaderChangeMsg_02(t *testing.T) {
	blkVerify, err := NewBlockVerify(nil)
	if err != nil {
		t.Fatalf("StartServer failed: %v\n", err)
	}

	number := uint64(40)
	blkVerify.handleRoleUpdatedMsg(&mc.RoleUpdatedMsg{Role: common.RoleValidator, BlockNum: number, Leader: common.Address{}})
	t.Logf("server curRole(%d) curNumber(%d)", blkVerify.curRole, blkVerify.curNumber)

	leaderMsg := &mc.LeaderChangeNotify{
		ConsensusState: true,
		Leader:         common.HexToAddress("FDABEFF1234"),
		Number:         number + 1,
		ReelectTurn:    0,
	}

	blkVerify.handleLeaderChangeNotify(leaderMsg)
	cp := blkVerify.processManage.GetCurrentProcess()
	if cp == nil {
		t.Fatalf("role is validator, but after handle leader change, there is no running process")
	}

	t.Logf("curprocess leader = %v", cp.leader.String())
	if cp.leader != leaderMsg.Leader {
		t.Fatalf("curprocess leader err! %v != %v", cp.leader, leaderMsg.Leader)
	}

	if cp.state != StateStart {
		t.Fatalf("curprocess state err, state = %d", cp.state)
	}
}

func TestRightRequestMsg_01(t *testing.T) {
	man := newMan(t, nil, false)

	reElection, err := reelection.New(man.BlockChain(), nil)
	if err != nil {
		t.Fatalf("create reelection server err!")
	}

	blkVerify, err := NewBlockVerify(man)
	if err != nil {
		t.Fatalf("StartServer failed: %v\n", err)
	}

	number := man.BlockChain().CurrentHeader().Number.Uint64()
	blkVerify.handleRoleUpdatedMsg(&mc.RoleUpdatedMsg{Role: common.RoleValidator, BlockNum: number, Leader: common.Address{}})
	leaderMsg := &mc.LeaderChangeNotify{
		ConsensusState: true,
		Leader:         common.HexToAddress("FDABEFF1234"),
		Number:         number + 1,
		ReelectTurn:    0,
	}
	blkVerify.handleLeaderChangeNotify(leaderMsg)
	t.Logf("server curRole(%d) curNumber(%d) leader(%s)", blkVerify.curRole, blkVerify.curNumber, leaderMsg.Leader.String())

	requstMsg := &mc.BlockVerifyConsensusReq{
		Header:  createHeader(t, man.BlockChain(), leaderMsg.Leader),
		TxsCode: nil,
	}

	blkVerify.handleRequestMsg(requstMsg)

	cp := blkVerify.processManage.GetCurrentProcess()
	if cp == nil {
		t.Fatalf("role is validator, but after handle leader change, there is no running process")
	}

	time.Sleep(time.Second * 20)

	if cp.state != StateDPOSVerify {
		t.Fatalf("process request err!")
	}
}

func TestVoteMsg_01(t *testing.T) {
	msgCenter := mc.NewCenter()
	verifyResultMsgCh := make(chan *mc.BlockVerifyConsensusOK, 1)
	sub, err := msgCenter.SubscribeEvent(mc.BlkVerify_VerifyConsensusOK, verifyResultMsgCh)
	if err != nil {
		t.Fatalf("sub reuslt msg err!")
	}
	defer sub.Unsubscribe()

	man := newMan(t, nil, false)

	blkVerify, err := NewBlockVerify(man)
	if err != nil {
		t.Fatalf("StartServer failed: %v\n", err)
	}

	number := man.BlockChain().CurrentHeader().Number.Uint64()
	blkVerify.handleRoleUpdatedMsg(&mc.RoleUpdatedMsg{Role: common.RoleValidator, BlockNum: number, Leader: common.Address{}})
	leaderMsg := &mc.LeaderChangeNotify{
		ConsensusState: true,
		Leader:         common.HexToAddress("FDABEFF1234"),
		Number:         number + 1,
		ReelectTurn:    0,
	}
	blkVerify.handleLeaderChangeNotify(leaderMsg)
	t.Logf("server curRole(%d) curNumber(%d) leader(%s)", blkVerify.curRole, blkVerify.curNumber, leaderMsg.Leader.String())

	requestMsg := &mc.BlockVerifyConsensusReq{
		Header:  createHeader(t, man.BlockChain(), leaderMsg.Leader),
		TxsCode: nil,
	}

	blkVerify.handleRequestMsg(requestMsg)

	//设置测试验证者
	validators, keys := generateTestValidators(2)
	self := mc.TopologyNodeInfo{
		Account:  common.HexToAddress(params.SignAccount),
		Position: 0,
		Type:     common.RoleValidator,
		Stock:    3,
	}
	validators = append(validators, self)

	//todo
	//ca.SetTestValidatorStocks(validators)

	//模拟投票
	hash := types.RlpHash(requestMsg)
	for addr, key := range keys {
		sign, err := crypto.SignWithValidate(hash.Bytes(), true, key)
		if err != nil {
			t.Fatalf("sign err(%s)", err)
		}
		voteMsg := &mc.ConsensusVote{
			SignHash:    hash,
			Sign:        common.BytesToSignature(sign),
			FromAccount: addr,
		}

		blkVerify.handleVoteMsg(voteMsg)
	}
	cp := blkVerify.processManage.GetCurrentProcess()
	if cp == nil {
		t.Fatalf("role is validator, but after handle leader change, there is no running process")
	}

	select {
	case result := <-verifyResultMsgCh:
		t.Logf("result leader(%v) number(%d)", result.Header.Leader.String(), result.Header.Number)
	}

	if cp.state != StateEnd {
		t.Fatalf("process vote err!")
	}
}

func newMan(t *testing.T, confOverride func(*man.Config), isBroadcastNode bool) *man.Matrix {
	// Create a temporary storage for the node keys and initialize it
	workspace, err := ioutil.TempDir("", "console-tester-")
	if err != nil {
		t.Fatalf("failed to create temporary keystore: %v", err)
	}

	// Create a networkless protocol stack and start an Matrix service within
	stack, err := node.New(&node.Config{DataDir: workspace, UseLightweightKDF: true, Name: "block_verify"})
	if err != nil {
		t.Fatalf("failed to create node: %v", err)
	}

	manConf := &man.Config{
		Genesis:   core.DeveloperGenesisBlock(15, common.Address{}),
		Manerbase: common.HexToAddress(testAddress),
		Manash: manash.Config{
			PowMode: manash.ModeTest,
		},
	}
	if confOverride != nil {
		confOverride(manConf)
	}
	if err = stack.Register(func(ctx *node.ServiceContext) (node.Service, error) { return man.New(ctx, manConf) }); err != nil {
		t.Fatalf("failed to register Matrix protocol: %v", err)
	}
	// Start the node and assemble the JavaScript console around it
	if err = stack.Start(); err != nil {
		t.Fatalf("failed to start test stack: %v", err)
	}
	stack.Attach()

	var matrix *man.Matrix
	stack.Service(&matrix)

	//创建账户
	createSignAccount(t, filepath.Join(stack.DataDir(), "keystore"))
	time.Sleep(time.Second * 8)

	return matrix
}

func createSignAccount(t *testing.T, dir string) *accounts.Account {
	// Create an encrypted keystore with standard crypto parameters
	ks := keystore.NewKeyStore(dir, keystore.StandardScryptN, keystore.StandardScryptP)

	// Create a new account with the specified encryption passphrase
	newAccount, err := ks.NewAccount("12345")
	if err != nil {
		t.Fatalf("Failed to create new account: %v", err)
	}

	params.SignAccount = newAccount.Address.String()
	params.SignAccountPassword = "12345"

	return &newAccount
}

func createHeader(t *testing.T, bc *core.BlockChain, leader common.Address) *types.Header {

	tstart := time.Now()
	parent := bc.CurrentBlock()

	tstamp := tstart.Unix()
	if parent.Time().Cmp(new(big.Int).SetInt64(tstamp)) >= 0 {
		tstamp = parent.Time().Int64() + 1
	}

	num := parent.Number()
	header := &types.Header{
		ParentHash: parent.Hash(),
		Number:     num.Add(num, common.Big1),
		GasLimit:   core.CalcGasLimit(parent),
		Extra:      nil,
		Time:       big.NewInt(tstamp),
		Leader:     leader,
	}

	if err := bc.Engine().Prepare(bc, header); err != nil {
		t.Fatalf("Failed to prepare header err = %s", err)
		return nil
	}

	return header
}

func generateTestValidators(count int) ([]mc.TopologyNodeInfo, map[common.Address]*ecdsa.PrivateKey) {
	validators := make([]mc.TopologyNodeInfo, 0)
	keys := make(map[common.Address]*ecdsa.PrivateKey)

	for len(validators) < count {
		key, err := crypto.GenerateKey()
		if err != nil {
			continue
		}

		info := mc.TopologyNodeInfo{
			Account:  crypto.PubkeyToAddress(key.PublicKey),
			Position: 0,
			Type:     common.RoleValidator,
			Stock:    3,
		}
		keys[info.Account] = key
		validators = append(validators, info)
	}

	return validators, keys
}
*/
