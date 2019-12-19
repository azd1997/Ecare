package net

import (
	"bytes"
	"encoding/gob"

	"github.com/azd1997/Ecare/ecoinlib/log"
)

// VersionMsg 版本消息
type VersionMsg struct {
	AddrFrom string
	MaxID uint 	// 区块链最大区块ID
	Hash Hash // Hash = hash(prevHash + newBLock.Hash)
}

// SendVersion 发送版本消息。
// 会被其他调用，所以传入参数，而不是直接从ecoin获取
func (gsm *GlobalStateMachine) SendVersion(to string, newMaxID uint, newHash Hash) error {

	// 构造VersionMsg
	version := VersionMsg{
		AddrFrom: to,
		MaxID:    newMaxID,
		Hash:     newHash,
	}

	// payload
	payload, err := GobEncode(version)
	if err != nil {
		return err
	}

	// request
	request := append(CmdToBytes("version"), payload...)

	// 发送
	err = gsm.SendMsg(to, request)
	if err != nil {
		return err
	}

	return nil
}

// HandleVersion 处理version消息
func (gsm *GlobalStateMachine) HandleVersion(request []byte) {

	var payload []byte
	var versionMsg VersionMsg
	var err error
	var lastBlock *Block

	// 提取payload
	payload = request[COMMAD_LENGTH:]

	// 解码得到versionMsg
	if err = gob.NewDecoder(bytes.NewReader(payload)).Decode(versionMsg); err != nil {
		goto ERR
	}

	// 暂存对方newHash信息


	// 获取自身最新区块
	if lastBlock, err = gsm.Ledger.GetBlockByHash(gsm.Ledger.LastHash); err != nil {
		goto ERR
	}

	// 比较双方maxID
	if versionMsg.MaxID > lastBlock.Id {
		// 对方比自己长。
	}


	return
ERR:
	// TODO：错误处理
	log.Error("HandleVersion: %s", err)
	return
}