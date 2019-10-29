package net

import (
	"bytes"
	"encoding/gob"
	"github.com/azd1997/Ecare/common/ecoinlib/log"
	"github.com/azd1997/Ecare/common/ecoinlib/types"
	"github.com/azd1997/Ecare/common/ecoinlib/utils"
)

func (e *ecoin) SendBlock(to types.Address, block *types.Block) error {
	payload, err := block.Serialize()
	if err != nil {
		return utils.WrapError("SendBlock", err)
	}
	request := append(CmdToBytes("block", e.Opts.CommandLength()), payload...)
	if err = e.SendData(to, request); err != nil {
		return utils.WrapError("SendBlock", err)
	}
	return nil
}

func (e *ecoin) HandleBlock(request []byte) {
	// 解析request
	var buf bytes.Buffer
	var b types.Block
	buf.Write(request[e.Opts.CommandLength():])
	err := gob.NewDecoder(&buf).Decode(&b)
	utils.LogErr("HandleBlock", err)

	// 检查block是否已存在
	exists, err := e.ledger.HasBlock(&b)
	utils.LogErr("HandleBlock", err)
	if exists && err == nil {
		log.Warn("HandleBlock: %s", ErrBlockAlreadyExists)
		return
	}

	// 本地不存在则将block进行检查(要求必须是追加在最后区块的后边)
	valid, err := e.ledger.VerifyBlock(&b)
	utils.LogErr("HandleBlock", err)
	if !valid {
		log.Error("HandleBlock: %s", ErrInvalidBlock)
	}

	// 检查没有问题后添加到本地区块链。
	err = e.ledger.AddBlock(&b)
	utils.LogErr("HandleBlock", err)

	//
}


