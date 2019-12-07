package tx

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"time"

	"github.com/azd1997/Ecare/ecoin/account"
	"github.com/azd1997/Ecare/ecoin/common"
	"github.com/azd1997/Ecare/ecoin/utils"
)

// TxP2R 第三方研究机构向病人发起的数据交易的阶段二交易
type TxP2R struct {
	Id          common.Hash      `json:"id"`
	Time        common.TimeStamp `json:"time"`
	From        account.UserId   `json:"from"`
	R2P         *TxR2P           `json:"r2p"`
	ResponseInfo    []byte           `json:"response"` // 比如说请求数据的密码
	Description string           `json:"description"`
	Sig         common.Signature `json:"sig"`
}

// newTxP2R 新建P2R转账交易(R2P交易二段)。
func newTxP2R(args *P2RArgs) (tx *TxP2R, err error) {

	// 构造tx
	tx = &TxP2R{
		Id:          nil,
		Time:        common.TimeStamp(time.Now().Unix()),
		From:        args.From,
		R2P:         args.R2P,
		ResponseInfo:    args.Response,
		Description: args.Description,
		Sig:         nil,
	}

	// 设置Id
	id, err := tx.Hash()
	if err != nil {
		return nil, utils.WrapError("newTxP2R", err)
	}
	tx.Id = id
	// 设置签名
	sig, err := args.FromAccount.Sign(id)
	if err != nil {
		return nil, utils.WrapError("newTxP2R", err)
	}
	tx.Sig = sig
	return tx, nil
}

/*******************************************************实现接口*********************************************************/

// TypeNo 获取交易类型编号
func (tx *TxP2R) TypeNo() uint {
	return TX_P2R
}

// Id 对于已生成的交易，获取其ID
func (tx *TxP2R) ID() common.Hash {
	return tx.Id
}

// Hash 计算交易哈希值，作为交易ID
func (tx *TxP2R) Hash() (hash common.Hash, err error) {
	txCopy := *tx
	txCopy.Id, txCopy.Sig = common.Hash{}, common.Signature{}
	var res []byte
	if res, err = txCopy.Serialize(); err != nil {
		return common.Hash{}, utils.WrapError("TxP2R_Hash", err)
	}
	hash1 := sha256.Sum256(res)
	return hash1[:], nil
}

// Serialize 交易序列化为字节切片
func (tx *TxP2R) Serialize() (result []byte, err error) {
	return utils.GobEncode(tx)
}

// String 转换为字符串，用于打印输出
func (tx *TxP2R) String() string {
	return utils.JsonMarshalIndentToString(tx)
}

// Deserialize 反序列化，必须提前 tx := &TxR2P{} 再调用
func (tx *TxP2R) Deserialize(p2rBytes []byte) (err error) {
	// 防止非空TxR2P调用该方法改变了自身内容

	// 反序列化
	var buf bytes.Buffer
	buf.Write(p2rBytes)
	err = gob.NewDecoder(&buf).Decode(tx)
	if err != nil {
		return utils.WrapError("TxP2R_Deserialize", err)
	}
	return nil
}

// IsValid 验证交易是否合乎规则
func (tx *TxP2R) IsValid() (err error) {

	// 检查交易时间有效性
	if tx.Time >= common.TimeStamp(time.Now().Unix()) {
		return utils.WrapError("TxP2R_IsValid", ErrWrongTimeTX)
	}

	// 检查From
	if err = tx.From.IsValid(account.Single, account.Patient); err != nil {
		return utils.WrapError("TxP2R_IsValid", err)
	}

	// 检查前部交易R2P不能为空。
	if tx.R2P == nil {
		return utils.WrapError("TxP2R_IsValid", ErrUnmatchedTxReceiver)
	}

	// 确保与来源交易接收者匹配
	if tx.From != tx.R2P.To {
		return utils.WrapError("TxP2R_IsValid", ErrUnmatchedTxReceiver)
	}

	// TODO： Response可用性检查。这部分交给交易双方自己做，除非达到仲裁条件，由验证节点进行仲裁才会再上层的handleTX方法中去处理


	// 验证交易ID是不是正确设置
	txHash, _ := tx.Hash()
	if string(txHash) != string(tx.Id) {
		return utils.WrapError("TxP2R_IsValid", ErrWrongTXID)
	}

	return nil
}

/*******************************************************实现接口*********************************************************/