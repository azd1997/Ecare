package net

import (
	"bytes"
	"encoding/gob"
	"github.com/azd1997/Ecare/common/ecoinlib/log"
	"github.com/azd1997/Ecare/common/ecoinlib/types"
	"github.com/azd1997/Ecare/common/ecoinlib/utils"
)

// SendAddr 向对方发送自己存有的L1集合。通常为A类节点调用
func (e *ecoin) SendAddr(to types.Address) {
	nodeList := AddrList{e.addrList.L1, nil, nil}
	payload, err := utils.GobEncode(nodeList)
	if err != nil {
		log.Error("SendAddr: %s", err)
	}
	cmdBytes := CmdToBytes("addr", e.Opts.CommandLength())
	request := append(cmdBytes, payload...)
	if err = e.SendData(to, request); err != nil {
		log.Error("SendAddr: %s", err)
	}
}

// HandleAddr 处理接收到的AddrList
func (e *ecoin) HandleAddr(request []byte) {
	//获取request内容
	var buff bytes.Buffer
	var payload AddrList

	buff.Write(request[e.Opts.CommandLength():])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Error("HandleAddr: gob_Decode: %s", err)
	}

	//更新已知节点集和，并向已知节点集合的节点请求区块信息
	e.addrList.L1 = MergeTwoNodeList(e.addrList.L1, payload.L1)
	log.Info("there are %d known forwarding nodes now", len(e.addrList.L1))
	RequestBlocks(e.addrList)
}