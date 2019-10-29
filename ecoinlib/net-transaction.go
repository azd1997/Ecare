package ecoin

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"time"

	"github.com/azd1997/Ecare/ecoinlib/log"

)

// Tx 通信传输的Tx结构体
type Tx struct {
	AddrFrom string
	Transaction []byte // 编码后的交易
	TypeNo uint //交易类型编号
}

// SendTx 发送交易消息体
func (gsm *GlobalStateMachine) SendTx(to string, tx TX) error {
	txBytes, err := tx.Serialize()
	if err != nil {
		return WrapError("SendTx", err)
	}
	payload, err := GobEncode(Tx{gsm.Opts().NodeAddress().String(), txBytes, tx.TypeNo()})
	if err != nil {
		return WrapError("SendTx", err)
	}
	request := append(CmdToBytes("tx"), payload...)
	err = gsm.SendMsg(to, request)
	if err != nil {
		return WrapError("SendTx", err)
	}
	return nil
}

// BroadcastTx 广播交易消息体
func (gsm *GlobalStateMachine) BroadcastTx(addrs []string, tx TX) error {
	txBytes, err := tx.Serialize()
	if err != nil {
		return WrapError("SendTx", err)
	}
	payload, err := GobEncode(Tx{gsm.Opts().NodeAddress().String(), txBytes, tx.TypeNo()})
	if err != nil {
		return WrapError("SendTx", err)
	}
	request := append(CmdToBytes("tx"), payload...)
	// 广播
	gsm.Broadcast(addrs, request)
	return nil
}

// HandleTx 处理接收到的Tx消息体
func (gsm *GlobalStateMachine) HandleTx(request []byte) {

	var buf bytes.Buffer
	var payload Tx
	var err error
	var tx TX
	var addr string
	var txID Hash

	// 解析request，得到Tx消息体
	buf.Write(request[COMMAD_LENGTH:])
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
	if payload.TypeNo == TX_COINBASE || payload.TypeNo == TX_ARBITRATE {
		// TODO: LogErr
		//return
		goto ERR
	}
	// 解码出交易
	if tx, err = DeserializeTX(payload.TypeNo, payload.Transaction); err != nil {
		goto ERR
	}
	// 检查Tx消息体内交易TX有效性
	if err = tx.IsValid(gsm); err != nil {
		goto ERR
	}
	// 有效则存入UBTXP。 todo： 处理交易并将符合条件的放入UCTXP是在打包区块和验证区块这一环节进行
	gsm.UBTXP.Add(tx)
	// TODO: log
	log.Info("HandleTx")

	// 向除自己和对方以外的所有可用转发节点发送存证表示我收到了这个交易
	if txID, err = tx.Hash(); err != nil {
		goto ERR
	}
	for _, addr = range gsm.Addrs.L1Ipv4Honest() {
		if err = gsm.SendInv(addr,"tx", []Hash{txID}); err != nil {
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
func (gsm *GlobalStateMachine) MineTx() error {
	var txs []TX
	var tx TX
	var err error
	var txArgs TxCoinbaseArgs
	var coinbase TX
	var self *UserID
	var lastId int

	// 获取本机账户
	self = gsm.Opts().UserID()
	if self.RoleNo > 9 {
		return ErrInvalidUserID
	}

	// 首先构造coinbase交易，并加入txs
	if lastId, err = gsm.Ledger.GetMaxId(); err != nil {
		return WrapError("MineTx", err)
	}
	txArgs = TxCoinbaseArgs{
		To:          *self,
		Amount:      gsm.Accounts.Map[self.ID].Role().CoinbaseReward(),
		Description: fmt.Sprintf(COINBASE_MSG_TEMP, time.Now().Format("2006/01/02 15:04:05"), self.ID, lastId + 1),
	}
	if coinbase, err = gsm.NewTX(TX_COINBASE, &txArgs); err != nil {
		return err
	}
	txs = append(txs, coinbase)

	// 遍历交易池，将验证合格后的交易存入txs
	for _, tx = range gsm.UBTXP.All() {

		if err = tx.IsValid(gsm); err != nil {
			continue 	// 有问题的跳过
		}
		txs = append(txs, tx)
	}

	// 若交易池提取不到有效交易（没有或者无效）就退出
	if len(txs) < 2 {
		// log.Info("MineTx: %s", ErrNoValidTransaction)
		return WrapError("MineTx", ErrNoValidTransaction)
	}

	// 构建新区快
	newBlock, err := gsm.MineBlock(txs)
	if err != nil {
		return WrapError("MineTx", err)
	}
	log.Info("MineTx: %s", newBlock.Hash)

	// 向转发节点集合中其他节点发送出块存证
	for _, addr := range gsm.Addrs.L1Ipv4Honest() {
		if err = gsm.SendInv(addr, "block", []Hash{newBlock.Hash}); err != nil {
			continue
		}
	}

	// 清空交易池
	gsm.UBTXP.Clear()

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

