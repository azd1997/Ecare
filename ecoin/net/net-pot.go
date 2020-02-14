package net

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
)


// SendPot 发送Pot
func (gsm *GlobalStateMachine) SendPot() error {
	var err error
	var txs []TX
	var pot PotMsg
	var txsBytes []byte
	var payload []byte
	var request []byte
	var addrs []string
	var hash [32]byte

	// 如果本地不是A类节点，没有竞选权限
	if gsm.Opts().Account().RoleNo > 9 {
		return ErrInvalidUserID
	}

	// 检验一次UBTXP，统计有效的交易数量，移至TBTXP
	gsm.CheckUBTXPAndUpdateTBTXP()
	// 现在UBTXP空了，TBTXP存了一批交易，而且这批交易不可以动，直到POT竞选结束出块而后清空。至于接收区块，则只是查找有无已使用的交易然后剔除

	txs = gsm.TBTXP.All()

	// 构造Pot
	if txsBytes, err = GobEncode(txs); err != nil {
		return err
	}
	hash = sha256.Sum256(txsBytes)
	pot = PotMsg{
		AddrFrom:  gsm.Opts().NodeAddress().String(),
		Pot:Pot{
			Num:  uint(len(txs)),
			Hash: hash[:],
		},
	}

	// payload
	if payload, err = GobEncode(pot); err != nil {
		return err
	}

	// request
	request = append(CmdToBytes("pot"), payload...)

	// 广播pot请求
	addrs = gsm.Addrs.L1Ipv4Honest()
	gsm.Broadcast(addrs, request)

	return nil
}

// HandlePot 接收到Pot后处理
func (gsm *GlobalStateMachine) HandlePot(request []byte) {

	var potMsg PotMsg
	var err error

	// 解析request，得到PotMsg
	if err = gob.NewDecoder(bytes.NewReader(request[COMMAD_LENGTH:])).Decode(potMsg); err != nil {
		goto ERR
	}

	// TODO: 思考Handle时是否需要将对方添加入本地节点集合； 以及节点集合的更新方式：
	// TODO: 是否要检查来源地址？要的，看源地址是否在本地POTMap中，不在的话，
	//  说明要么其是不诚实被封杀了，要么是还不知道的节点（第一次出现的）。
	//  此外，第一次出现的节点首先是会去找一个指定的中间节点来使其他人认识他。然后这个中间人把它介绍给其他人。



	// 检查无误后，更新POTMap
	gsm.Workers.SetPot(potMsg.AddrFrom, potMsg.Pot)

	// TODO: 需要一个定时协程。出块或接收到新区块时通知其重置定时器。


	return
ERR:
	// TODO: 错误处理
	LogErr("HandlePot", err)
}




