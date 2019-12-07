package tx

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"github.com/azd1997/Ecare/ecoin/account"
	"github.com/azd1997/Ecare/ecoin/common"
	"github.com/azd1997/Ecare/ecoin/utils"
	"time"
)


// TxD2P 病人向下班医生发起的心电诊断交易，阶段二
type TxD2P struct {
	Id          common.Hash          `json:"id"`
	Time        common.TimeStamp `json:"time"`
	From        account.UserId        `json:"from"`
	P2D    *TxP2D        `json:"p2dBytes"`
	Response    []byte        `json:"response"` // 比如说请求数据的密码
	Description string        `json:"description"`
	Sig         common.Signature     `json:"sig"`
}

// newTxD2P 新建D2P转账交易(P2D交易二段)。
func newTxD2P(args *D2PArgs) (tx *TxD2P, err error) {
	//// 检验参数
	//if err = args.CheckArgsValue(); err != nil {
	//	return nil, utils.WrapError("newTxD2P", err)
	//}
	//
	//// 获取转账者UserID
	//fromID, err := args.From.UserID(args.Gsm.opts.ChecksumLength(), args.Gsm.opts.Version())
	//if err != nil {
	//	return nil, utils.WrapError("newTxD2P", err)
	//}

	// 构造tx
	tx = &TxD2P{
		Id:          nil,
		Time:        common.TimeStamp(time.Now().Unix()),
		From:        args.From,
		P2D:         args.P2D,
		Response:    args.Response,
		Description: args.Description,
		Sig:         nil,
	}

	// 设置Id
	id, err := tx.Hash()
	if err != nil {
		return nil, utils.WrapError("newTxD2P", err)
	}
	tx.Id = id
	// 设置签名
	sig, err := args.FromAccount.Sign(id)
	if err != nil {
		return nil, utils.WrapError("newTxD2P", err)
	}
	tx.Sig = sig
	return tx, nil
}

// TypeNo 获取交易类型编号
func (tx *TxD2P) TypeNo() uint {
	return TX_D2P
}

// Id 对于已生成的交易，获取其ID
func (tx *TxD2P) ID() common.Hash {
	return tx.Id
}

// Hash 计算交易哈希值，作为交易ID
func (tx *TxD2P) Hash() (hash common.Hash, err error) {
	txCopy := *tx
	txCopy.Id, txCopy.Sig = common.Hash{}, common.Signature{}
	var res []byte
	if res, err = txCopy.Serialize(); err != nil {
		return common.Hash{}, utils.WrapError("TxD2P_Hash", err)
	}
	hash1 := sha256.Sum256(res)
	return hash1[:], nil
}

// Serialize 交易序列化为字节切片
func (tx *TxD2P) Serialize() (result []byte, err error) {
	return utils.GobEncode(tx)
}

// String 转换为字符串，用于打印输出
func (tx *TxD2P) String() string {
	return utils.JsonMarshalIndentToString(tx)
}

// Deserialize 反序列化，必须提前 tx := &TxD2P{} 再调用
func (tx *TxD2P) Deserialize(d2pBytes []byte) (err error) {
	// 防止非空TxR2P调用该方法改变了自身内容

	// 反序列化
	var buf bytes.Buffer
	buf.Write(d2pBytes)
	err = gob.NewDecoder(&buf).Decode(tx)
	if err != nil {
		return utils.WrapError("TxD2P_Deserialize", err)
	}
	return nil
}

// IsValid 验证交易是否合乎规则
func (tx *TxD2P) IsValid() (err error) {

	/*	tx = &TxD2P{
		ID:          Hash{},
		Time:        UnixTimeStamp(time.Now().Unix()),
		From:fromID,
		P2DBytes:    p2dBytes,
		Response:    response,
		Description: description,
		Sig:         Signature{},
	}*/

	// 检查交易时间有效性
	if tx.Time >= common.TimeStamp(time.Now().Unix()) {
		return utils.WrapError("TxD2P_IsValid", ErrWrongTimeTX)
	}

	// 检查FromId有效性及是否与P2D的to匹配
	if err = tx.From.IsValid(account.Single, account.Doctor); err != nil {
		return utils.WrapError("TxD2P_IsValid", err)
	}
	if tx.From != tx.P2D.To {
		return utils.WrapError("TxD2P_IsValid", ErrUnmatchedTxReceiver)
	}

	// 检查fromID的有效性、可用性和from签名是否匹配

	// TODO： Response可用性检查。这部分交给交易双方自己做，除非达到仲裁条件，由验证节点进行仲裁才会再上层的handleTX方法中去处理

	// 检查前部交易是不是一个P2D交易，为空则错误；不为空必须是符合P2D交易体且交易ID在未完成交易池中，否则认为是不合法交易

	// 验证交易ID是不是正确设置
	txHash, _ := tx.Hash()
	if string(txHash) != string(tx.Id) {
		return utils.WrapError("TxD2P_IsValid", ErrWrongTXID)
	}

	return nil
}


