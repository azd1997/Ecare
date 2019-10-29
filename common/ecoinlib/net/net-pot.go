package net

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"github.com/azd1997/Ecare/common/ecoinlib/types"
	"github.com/azd1997/Ecare/common/ecoinlib/utils"
)

// Pot 交易量证明，所有节点在接收到最新区块后刷新本地period计时，period/2计时一到广播本机Pot，所有人得到一份Pot集合后决出POT胜者.
// 再period/2后
type PotMsg struct {
	AddrFrom string
	NumOfTX uint
	WholeHash types.Hash	// 用来检查区块内所含交易是否与竞争POT时一致，不含coinbase和仲裁交易
}

// SendPot 发送Pot
func (e *ecoin) SendPot() error {
	var err error
	var pot PotMsg
	var txsBytes []byte
	var payload []byte
	var request []byte
	var addrs []string

	// 如果本地不是A类节点，没有竞选权限
	if e.Opts().UserID().RoleNo > 9 {
		return types.ErrInvalidUserID
	}

	// 检验一次UBTXP，统计有效的交易数量，移至TBTXP
	e.CheckUBTXPAndUpdateTBTXP()
	// 现在UBTXP空了

	// 构造Pot
	if txsBytes, err = utils.GobEncode(e.TBTXP.All()); err != nil {
		return err
	}
	pot = PotMsg{
		AddrFrom:  e.Opts().NodeAddress().NodeAddr(),
		NumOfTX:   uint(len(e.TBTXP.All())),
		WholeHash: sha256.Sum256(txsBytes),
	}

	// payload
	if payload, err = utils.GobEncode(pot); err != nil {
		return err
	}

	// request
	request = append(CmdToBytes("pot", e.Opts().CommandLength()), payload...)

	// 广播pot请求
	addrs = e.Addrs.L1Ipv4Honest()
	e.Broadcast(addrs, request)

	return nil
}

// HandlePot 接收到Pot后处理
func (e *ecoin) HandlePot(request []byte) {

	var potMsg PotMsg
	var err error

	// 解析request，得到PotMsg
	if err = gob.NewDecoder(bytes.NewReader(request[e.Opts().CommandLength():])).Decode(potMsg); err != nil {
		goto ERR
	}

	// TODO: 思考Handle时是否需要将对方添加入本地节点集合； 以及节点集合的更新方式：
	// TODO: 是否要检查来源地址？要的，看源地址是否在本地POTMap中，不在的话，
	//  说明要么其是不诚实被封杀了，要么是还不知道的节点（第一次出现的）。
	//  此外，第一次出现的节点首先是会去找一个指定的中间节点来使其他人认识他。然后这个中间人把它介绍给其他人。



	// 检查无误后，更新POTMap
	e.POT.Set(potMsg.AddrFrom, types.Pot{Num: potMsg.NumOfTX, Hash: potMsg.WholeHash})

	// TODO: 需要一个定时协程。出块或接收到新区块时通知其重置定时器。


	return
ERR:
	// TODO: 错误处理
	utils.LogErr("HandlePot", err)
}




