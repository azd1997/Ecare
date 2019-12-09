package transaction

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"github.com/azd1997/Ecare/ecoin/crypto"
	"time"

	"github.com/azd1997/Ecare/ecoin/account"
	"github.com/azd1997/Ecare/ecoin/common"
	"github.com/azd1997/Ecare/ecoin/utils"
)

// TxD2P 病人向下班医生发起的心电诊断交易，阶段二
type TxD2P struct {
	Id          crypto.Hash      `json:"id"`
	Time        common.TimeStamp `json:"time"`
	From        account.UserId   `json:"from"`
	P2D         *TxP2D           `json:"p2dBytes"`
	Response    []byte           `json:"response"` // 比如说请求数据的密码
	Description string           `json:"description"`
	Sig         crypto.Signature `json:"sig"`
}

// newTxD2P 新建D2P转账交易(P2D交易二段)。
func newTxD2P(args *D2PArgs) (tx *TxD2P, err error) {

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

/*******************************************************实现接口*********************************************************/

// TypeNo 获取交易类型编号
func (tx *TxD2P) TypeNo() uint {
	return TX_D2P
}

// Id 对于已生成的交易，获取其ID
func (tx *TxD2P) ID() crypto.Hash {
	return tx.Id
}

// Hash 计算交易哈希值，作为交易ID
func (tx *TxD2P) Hash() (hash crypto.Hash, err error) {
	txCopy := *tx
	txCopy.Id, txCopy.Sig = crypto.Hash{}, crypto.Signature{}
	var res []byte
	if res, err = txCopy.Serialize(); err != nil {
		return crypto.Hash{}, utils.WrapError("TxD2P_Hash", err)
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
func (tx *TxD2P) IsValid(txFunc ValidateTxFunc) (err error) {

	// 检查交易时间有效性
	if tx.Time >= common.TimeStamp(time.Now().Unix()) {
		return utils.WrapError("TxD2P_IsValid", ErrWrongTime)
	}

	// 检查From有效性
	if err = tx.From.IsValid(account.Single, account.Doctor); err != nil {
		return utils.WrapError("TxD2P_IsValid", err)
	}

	// P2D不能为空
	if tx.P2D == nil {
		return utils.WrapError("TxD2P_IsValid", ErrNilSourceTx)
	}

	// 是否与P2D的to匹配
	if tx.From != tx.P2D.To {
		return utils.WrapError("TxD2P_IsValid", ErrUnmatchedSender)
	}

	// 验证交易ID是不是正确设置
	txHash, _ := tx.Hash()
	if string(txHash) != string(tx.Id) {
		return utils.WrapError("TxD2P_IsValid", ErrWrongTxId)
	}


	// 其他的检查交给传入的检查方法去做
	if err = txFunc(tx); err != nil {
		return utils.WrapError("TxD2P_IsValid", err)
	}

	return nil
}

/*******************************************************实现接口*********************************************************/
