package net

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/azd1997/Ecare/common/ecoinlib/log"
	"github.com/azd1997/Ecare/common/ecoinlib/types"
	"github.com/azd1997/Ecare/common/ecoinlib/utils"
)

// Tx 通信传输的Tx结构体
type Tx struct {
	AddrFrom string
	Transaction []byte // 编码后的交易
	TypeNo uint //交易类型编号
}

// SendTx 发送交易消息体
func (e *ecoin) SendTx(to string, tx types.TX) error {
	txBytes, err := tx.Serialize()
	if err != nil {
		return utils.WrapError("SendTx", err)
	}
	payload, err := utils.GobEncode(Tx{e.Opts().NodeAddress().NodeAddr(), txBytes, tx.TypeNo()})
	if err != nil {
		return utils.WrapError("SendTx", err)
	}
	request := append(CmdToBytes("tx", e.Opts().CommandLength()), payload...)
	err = e.SendData(to, request)
	if err != nil {
		return utils.WrapError("SendTx", err)
	}
	return nil
}

// HandleTx 处理接收到的Tx消息体
func (e *ecoin) HandleTx(request []byte) {

	var buf bytes.Buffer
	var payload Tx
	var err error
	var tx types.TX
	var addr string
	var txID types.Hash

	// 解析request，得到Tx消息体
	buf.Write(request[e.Opts().CommandLength():])
	if err = gob.NewDecoder(&buf).Decode(&payload); err != nil {
		goto ERR
	}
	//utils.LogErr("HandleTx", err)

	// 检查Tx消息体内发送者的节点地址信息
	//if !payload.AddrFrom.Honest {
	//	log.Warn("HandleTx: %s", "received tx from an dishonest node ")
	//	return
	//}
	// TODO: HandleTx其实是在交易体验证中进行，验证账号是否可用，没必要验证节点是否诚实，因为很多交易制造者是B类用户，这类用户数量多，没必要再去做这方面验证

	// 接收到的交易不能是coinbase和仲裁交易。
	if payload.TypeNo == types.TX_COINBASE || payload.TypeNo == types.TX_ARBITRATE {
		// TODO: LogErr
		//return
		goto ERR
	}
	// 解码出交易
	if tx, err = types.DeserializeTX(payload.TypeNo, payload.Transaction); err != nil {
		goto ERR
	}
	// 检查Tx消息体内交易TX有效性
	if err = tx.IsValid(&e.GlobalStateMachine); err != nil {
		goto ERR
	}
	// 有效则存入UBTXP。 todo： 处理交易并将符合条件的放入UCTXP是在打包区块和验证区块这一环节进行
	e.UBTXP.Add(tx)
	// TODO: log
	log.Info("HandleTx")

	// 向除自己和对方以外的所有可用转发节点发送存证表示我收到了这个交易
	if txID, err = tx.Hash(); err != nil {
		goto ERR
	}
	for _, addr = range e.Addrs.L1 {
		if err = e.SendInv(addr,"tx", []types.Hash{txID}); err != nil {
			continue	// 跳过
		}
	}
	return

	// 错误处理
ERR:
	log.Error("HandleTx: %s", err)
	return
}

// MineTx 限定转发节点进行挖矿
func (e *ecoin) MineTx() error {
	var txs []types.TX
	var tx types.TX
	var err error
	var txArgs types.TxCoinbaseArgs
	var coinbase types.TX
	var self types.UserID

	// 获取本机账户
	self = e.Opts().UserID()
	if self.RoleNo > 9 {
		return types.ErrInvalidUserID
	}

	// 首先构造coinbase交易，并加入txs
	txArgs = types.TxCoinbaseArgs{
		To:          self,
		Amount:      e.Accounts.Map[self.ID].Role().CoinbaseReward(),
		Description: fmt.Sprintf("%s create a new block", self.ID),
	}
	if coinbase, err = e.NewTX(types.TX_COINBASE, &txArgs); err != nil {
		return err
	}
	txs = append(txs, coinbase)

	// 遍历交易池，将验证合格后的交易存入txs
	for _, tx = range e.UBTXP.All() {

		if err = tx.IsValid(&e.GlobalStateMachine); err != nil {
			continue 	// 有问题的跳过
		}
		txs = append(txs, tx)
	}


	// 若交易池提取不到有效交易（没有或者无效）就退出
	if len(txs) < 2 {
		// log.Info("MineTx: %s", ErrNoValidTransaction)
		return utils.WrapError("MineTx", ErrNoValidTransaction)
	}

	// 构建新区快
	newBlock, err := e.MineBlock(txs)
	if err != nil {
		return utils.WrapError("MineTx", err)
	}
	log.Info("MineTx: %s", newBlock.Hash)

	// 向转发节点集合中其他节点发送出块存证
	for _, addr := range e.Addrs.L1 {
		if err = e.SendInv(addr.NodeAddr(), "block", []types.Hash{newBlock.Hash}); err != nil {
			continue
		}
	}

	// 清空交易池
	e.UBTXP.Clear()

	// TODO： 刷新period定时

	//// 递归调用挖矿
	//if len(e.txPool) > 0 {
	//	err = e.MineTx()
	//	if err != nil {
	//		return utils.WrapError("MineTx", err)
	//	}
	//}


	return nil
}

