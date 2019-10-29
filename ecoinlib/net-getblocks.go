package ecoin

import (
	"bytes"
	"encoding/gob"
	"github.com/azd1997/Ecare/ecoinlib/log"
)

// RequestBlocks 向可用转发节点集合中并发发送请求区块的请求
func (gsm *GlobalStateMachine) RequestBlocks() {
	for _, node := range gsm.Addrs.L1 {
		go gsm.SendGetBlocks(node.String())	// 协程函数执行完毕后，go GC会自动清除该协程，自己手动清除有可能在cpu较忙时执行，造成卡顿
	}
}

// GetBlocks 获取区块的消息体
type GetBlocksMsg struct {
	AddrFrom string
}

// SendGetBlocks 发送获取区块请求
func (gsm *GlobalStateMachine) SendGetBlocks(to string) (err error) {
	payload, err := GobEncode(GetBlocksMsg{gsm.Opts().NodeAddress().String()})
	if err != nil {
		return WrapError("SendGetBlocks", err)
	}
	request := append(CmdToBytes("getblocks"), payload...)
	if err = gsm.SendData(to, request); err != nil {
		return WrapError("SendGetBlocks", err)
	}
	return nil
}

// HandleGetBlocks 处理获取区块请求
func (gsm *GlobalStateMachine) HandleGetBlocks(request []byte) {

	// Gob Decode
	var buf bytes.Buffer
	var payload GetBlocksMsg
	buf.Write(request[COMMAD_LENGTH:])
	dec := gob.NewDecoder(&buf)
	err := dec.Decode(&payload)
	if err != nil {
		log.Error("HandleGetBlocks: %s", err)
	}

	// 发送自己拥有的区块存证，也就是所有区块的哈希集合(从后到前)
	blocks, err := gsm.Ledger.GetBlockHashes()
	if err != nil {
		log.Error("HandleGetBlocks: %s", err)
	}
	err = gsm.SendInv(payload.AddrFrom, "block", blocks)
	if err != nil {
		log.Error("HandleGetBlocks: %s", err)
	}
}























