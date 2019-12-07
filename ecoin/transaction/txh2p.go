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



// TxH2P 病人向医院发起的心电数据诊断，分人工和机器自动分析两种。阶段二
type TxH2P struct {
	Id          common.Hash          `json:"id"`
	Time        common.TimeStamp `json:"time"`
	From        account.UserId        `json:"from"`
	P2H    *TxP2H        `json:"p2h"`
	Response    []byte        `json:"response"` // 比如说请求数据的密码
	Description string        `json:"description"`
	Sig         common.Signature     `json:"sig"`
}

// newTxH2P 新建H2P转账交易(P2H交易二段)。
func newTxH2P(args *H2PArgs) (tx *TxH2P, err error) {

	// 构造tx
	tx = &TxH2P{
		Id:          nil,
		Time:        common.TimeStamp(time.Now().Unix()),
		From:        args.From,
		P2H:         args.P2H,
		Response:    args.Response,
		Description: args.Description,
		Sig:         nil,
	}

	// 设置Id
	id, err := tx.Hash()
	if err != nil {
		return nil, utils.WrapError("newTxH2P", err)
	}
	tx.Id = id
	// 设置签名
	sig, err := args.FromAccount.Sign(id)
	if err != nil {
		return nil, utils.WrapError("newTxH2P", err)
	}
	tx.Sig = sig
	return tx, nil
}

// TypeNo 获取交易类型编号
func (tx *TxH2P) TypeNo() uint {
	return TX_H2P
}

// Id 对于已生成的交易，获取其ID
func (tx *TxH2P) ID() common.Hash {
	return tx.Id
}

// Hash 计算交易哈希值，作为交易ID
func (tx *TxH2P) Hash() (hash common.Hash, err error) {
	txCopy := *tx
	txCopy.Id, txCopy.Sig = common.Hash{}, common.Signature{}
	var res []byte
	if res, err = txCopy.Serialize(); err != nil {
		return common.Hash{}, utils.WrapError("TxH2P_Hash", err)
	}
	hash1 := sha256.Sum256(res)
	return hash1[:], nil
}

// Serialize 交易序列化为字节切片
func (tx *TxH2P) Serialize() (result []byte, err error) {
	return utils.GobEncode(tx)
}

// String 转换为字符串，用于打印输出
func (tx *TxH2P) String() string {
	return utils.JsonMarshalIndentToString(tx)
}

// Deserialize 反序列化，必须提前 tx := &TxH2P{} 再调用
func (tx *TxH2P) Deserialize(h2pBytes []byte) (err error) {
	// 防止非空TxR2P调用该方法改变了自身内容

	// 反序列化
	var buf bytes.Buffer
	buf.Write(h2pBytes)
	err = gob.NewDecoder(&buf).Decode(tx)
	if err != nil {
		return utils.WrapError("TxH2P_Deserialize", err)
	}
	return nil
}

// IsValid 验证交易是否合乎规则
func (tx *TxH2P) IsValid() (err error) {

	/*	tx = &TxH2P{
		Id:          Hash{},
		Time:        UnixTimeStamp(time.Now().Unix()),
		From:fromID,
		P2HBytes:    p2hBytes,
		Response:    response,
		Description: description,
		Sig:         Signature{},
	}*/

	// 检查交易时间有效性
	if tx.Time >= common.TimeStamp(time.Now().Unix()) {
		return utils.WrapError("TxH2P_IsValid", ErrWrongTimeTX)
	}

	// 检查fromID的有效性、可用性和from签名是否匹配
	if err = tx.From.IsValid(account.Single, account.Hospital); err != nil {
		return utils.WrapError("TxH2P_IsValid", err)
	}
	if tx.From != tx.P2H.To {
		return utils.WrapError("TxH2P_IsValid", ErrUnmatchedTxReceiver)
	}

	// 验证交易ID是不是正确设置
	txHash, _ := tx.Hash()
	if string(txHash) != string(tx.Id) {
		return utils.WrapError("TxH2P_IsValid", ErrWrongTXID)
	}

	return nil
}

