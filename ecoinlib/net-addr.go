package ecoin

import (
	"bytes"
	"encoding/gob"
	"github.com/azd1997/Ecare/ecoinlib/log"
)

// SendAddr 向对方发送自己存有的L1集合。通常为A类节点调用
func (gsm *GlobalStateMachine) SendAddr(to string) (err error) {
	// 得到所有可用的转发节点集合
	var nodeList []*Address
	var self *Address
	self = gsm.Opts().NodeAddress()
	nodeList = append([]*Address{self}, gsm.Addrs.L1...)
	payload, err := GobEncode(nodeList)
	if err != nil {
		return WrapError("SendAddr", err)
	}
	cmdBytes := CmdToBytes("addr")
	request := append(cmdBytes, payload...)
	if err = gsm.SendData(to, request); err != nil {
		return WrapError("SendAddr", err)
	}
	return nil
}

// HandleAddr 处理接收到的AddrList
func (gsm *GlobalStateMachine) HandleAddr(request []byte) {
	//获取request内容
	var buff bytes.Buffer
	var payload []*Address

	buff.Write(request[COMMAD_LENGTH:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Error("HandleAddr: %s", err)
	}

	//更新已知节点集和，并向已知节点集合的节点请求区块信息
	gsm.Addrs.L1 = MergeTwoNodeList(gsm.Addrs.L1, payload)
	log.Info("there are %d known forwarding nodes now", len(gsm.Addrs.L1))
	gsm.RequestBlocks()
}