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

// TxGeneral 通用交易， 一方转给另一方，无需确认
type TxGeneral struct {
//	TxBase
	Id          common.Hash      `json:"id"`
	Time        common.TimeStamp `json:"time"`
	From        account.UserId   `json:"from"`
	To          account.UserId   `json:"to"`
	Amount      common.Coin      `json:"amount"`
	Sig         common.Signature `json:"sig"`
	Description string           `json:"description"`
}

// newTxGeneral 新建普通转账交易。
// 这里是内部函数，使用GeneralArgs传参；该函数的上一层进行包装时使用Args进行传参，而后使用断言进入本函数
func newTxGeneral(args *GeneralArgs) (tx *TxGeneral, err error) {

	// 构造tx
	tx = &TxGeneral{
		Id:          nil,
		Time:        common.TimeStamp(time.Now().Unix()),
		From:        args.From,
		To:          args.To,
		Amount:      args.Amount,
		Sig:         nil,
		Description: args.Description,
	}

	// 设置Id
	id, err := tx.Hash()
	if err != nil {
		return nil, utils.WrapError("newTxGeneral", err)
	}
	tx.Id = id
	// 设置签名
	sig, err := args.FromAccount.Sign(id)
	if err != nil {
		return nil, utils.WrapError("newTxGeneral", err)
	}
	tx.Sig = sig
	return tx, nil
}

/*******************************************************实现接口*********************************************************/

// TypeNo 获取交易类型编号
func (tx *TxGeneral) TypeNo() uint {
	return TX_GENERAL
}

// Id 对于已生成的交易，获取其ID
func (tx *TxGeneral) ID() common.Hash {
	return tx.Id
}

// Hash 计算交易哈希值，作为交易ID
func (tx *TxGeneral) Hash() (hash common.Hash, err error) {
	txCopy := *tx
	txCopy.Id, txCopy.Sig = common.Hash{}, common.Signature{} // 置空值
	var res []byte
	if res, err = txCopy.Serialize(); err != nil {
		return common.Hash{}, utils.WrapError("TxGeneral_Hash", err)
	}
	hash1 := sha256.Sum256(res)
	return hash1[:], nil
}

// Serialize 交易序列化为字节切片
func (tx *TxGeneral) Serialize() (result []byte, err error) {
	return utils.GobEncode(tx)
}

// String 转换为字符串，用于打印输出
func (tx *TxGeneral) String() string {
	return utils.JsonMarshalIndentToString(tx)
}

// Deserialize 反序列化，必须提前 tx := &TxGeneral{} 再调用
func (tx *TxGeneral) Deserialize(data []byte) (err error) {

	// 反序列化
	var buf bytes.Buffer
	buf.Write(data)
	err = gob.NewDecoder(&buf).Decode(tx)
	if err != nil {
		return utils.WrapError("TxGeneral_Deserialize", err)
	}
	return nil
}

// IsValid 验证交易是否合乎规则。注意这里和Args.Check做的事是一样的，只不过调用方不同
// 这个方法提供给检查交易方调用，而args.Check由制造交易者调用
func (tx *TxGeneral) IsValid() (err error) {

	// 检查交易时间有效性
	if tx.Time >= common.TimeStamp(time.Now().Unix()) {
		return utils.WrapError("TxGeneral_IsValid", ErrWrongTimeTX)
	}

	// 检查From
	err = tx.From.IsValid(account.All, 0)
	if err != nil {
		return utils.WrapError("TxGeneral_IsValid", err)
	}

	// 检查 To 的有效性
	if err = tx.To.IsValid(account.All, 0); err != nil {
		return utils.WrapError("TxGeneral_IsValid", err)
	}

	// 验证交易ID是不是正确设置
	txHash, _ := tx.Hash()
	if string(txHash) != string(tx.Id) {
		return utils.WrapError("TxGeneral_IsValid", ErrWrongTXID)
	}

	return nil
}

/*******************************************************实现接口*********************************************************/
