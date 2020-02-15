package net

import (
	"bytes"
	"encoding/gob"
	"github.com/azd1997/Ecare/ecoin/eaddr"
	"github.com/azd1997/Ecare/ecoin/utils"
	"github.com/azd1997/Ecare/ecoinlib/log"
	"github.com/azd1997/ego/enet/etcp"
)

// SendAddrs 向对方发送自己存有的L1集合。通常为A类节点调用
func (n *TCPNode) SendAddrs1(to eaddr.Addr) (err error) {
	// 得到所有可用的转发节点集合
	n.EAddrs.RLock()
	nodeList := n.EAddrs.ValidAddrs()
	n.EAddrs.RUnlock()

	payload, err := utils.GobEncode(nodeList)
	if err != nil {
		return utils.WrapError("SendAddr", err)
	}

	request, err := n.PackData(CmdSendAddrs, payload)
	if err != nil {
		return utils.WrapError("SendAddr", err)
	}

	if err = n.SendData(to, request); err != nil {
		return utils.WrapError("SendAddr", err)
	}
	return nil
}

// HandleAddr 处理接收到的AddrList
func (n *TCPNode) HandleAddr(request []byte) {
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

type AddrsRouter struct {
	etcp.BaseRouter
}

func (handler *AddrsRouter) Handle(r etcp.IRequest) {

}