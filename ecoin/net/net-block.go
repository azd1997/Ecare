package net

import (
	"bytes"
	"encoding/gob"
	"github.com/azd1997/Ecare/ecoinlib/log"
)

// SendBlock 发送区块
func (gsm *GlobalStateMachine) SendBlock(to string, block *Block) (err error) {
	payload, err := block.Serialize()
	if err != nil {
		return WrapError("SendBlock", err)
	}
	request := append(CmdToBytes("block"), payload...)
	if err = gsm.SendMsg(to, request); err != nil {
		return WrapError("SendBlock", err)
	}
	return nil
}

// HandleBlock 处理接收到的区块
func (gsm *GlobalStateMachine) HandleBlock(request []byte) {
	// 解析request
	var buf bytes.Buffer
	var b Block
	buf.Write(request[COMMAD_LENGTH:])
	err := gob.NewDecoder(&buf).Decode(&b)
	LogErr("HandleBlock", err)

	// 检查block是否已存在
	exists, err := gsm.Ledger.HasBlock(&b)
	LogErr("HandleBlock", err)
	if exists && err == nil {
		log.Warn("HandleBlock: %s", ErrBlockAlreadyExists)
		return
	}

	// 本地不存在则将block进行检查(要求必须是追加在最后区块的后边)
	err = b.IsValid(gsm)
	if err != nil {
		log.Error("HandleBlock: %s", ErrInvalidBlock)
	}

	// 检查没有问题后添加到本地区块链。
	err = gsm.Ledger.AddBlock(&b)
	LogErr("HandleBlock", err)

	// todo
}


