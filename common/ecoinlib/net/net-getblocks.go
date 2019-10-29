package net

import (
	"bytes"
	"encoding/gob"
	"github.com/azd1997/Ecare/common/ecoinlib/log"
	"github.com/azd1997/Ecare/common/ecoinlib/types"
	"github.com/azd1997/Ecare/common/ecoinlib/utils"
)

// 向可用转发节点集合中并发发送请求区块的请求
func (e *ecoin) RequestBlocks() {
	for _, node := range e.addrList.L1 {
		go e.SendGetBlocks(node)	// 协程函数执行完毕后，go GC会自动清除该协程，自己手动清除有可能在cpu较忙时执行，造成卡顿
	}
}

type GetBlocks struct {
	AddrFrom types.Address
}

func (e *ecoin) SendGetBlocks(to types.Address) error {
	payload, err := utils.GobEncode(GetBlocks{e.Opts.NodeAddress()})
	if err != nil {
		return utils.WrapError("SendGetBlocks", err)
	}
	request := append(CmdToBytes("getblocks", e.Opts.CommandLength()), payload...)
	if err = e.SendData(to, request); err != nil {
		return utils.WrapError("SendGetBlocks", err)
	}
	return nil
}

func (e *ecoin) HandleGetBlocks(request []byte) {

	// Gob Decode
	var buf bytes.Buffer
	var payload GetBlocks
	buf.Write(request[e.Opts.CommandLength():])
	dec := gob.NewDecoder(&buf)
	err := dec.Decode(&payload)
	if err != nil {
		log.Error("HandleGetBlocks: %s", err)
	}

	// 发送自己拥有的区块存证，也就是所有区块的哈希集合(从后到前)
	blocks, err := e.ledger.GetBlockHashes()
	if err != nil {
		log.Error("HandleGetBlocks: %s", err)
	}
	err = e.SendInv(payload.AddrFrom, "block", blocks)
	if err != nil {
		log.Error("HandleGetBlocks: %s", err)
	}
}























