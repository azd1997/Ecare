package tx

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"github.com/azd1997/Ecare/ecoin/crypto"
	"time"

	"github.com/azd1997/Ecare/ecoin/account"
	"github.com/azd1997/Ecare/ecoin/common"
	"github.com/azd1997/Ecare/ecoin/storage"
	"github.com/azd1997/Ecare/ecoin/utils"
)


// TxP2D 病人向下班医生发起的心电诊断交易，阶段一		TODO: 暂时只支持找指定医生诊断；后边考虑广播交易等待医生解决
type TxP2D struct {

	Id             crypto.Hash        `json:"id"`
	Time           common.TimeStamp   `json:"time"`
	From           account.UserId     `json:"from"`
	To             account.UserId     `json:"to"`
	Amount         common.Coin        `json:"amount"`
	PurchaseTarget storage.TargetData `json:"purchaseTarget"`
	Description    string             `json:"description"`
	Sig            crypto.Signature   `json:"sig"`
}

// newTxP2D 新建P2D转账交易。
func newTxP2D(args *P2DArgs) (tx *TxP2D, err error) {

	// 构造tx
	tx = &TxP2D{
		Id:             nil,
		Time:           common.TimeStamp(time.Now().Unix()),
		From:           args.From,
		To:             args.To,
		Amount:         args.Amount,
		PurchaseTarget: args.PurchaseTarget,
		Description:    args.Description,
		Sig:            nil,
	}

	// 设置Id
	id, err := tx.Hash()
	if err != nil {
		return nil, utils.WrapError("newTxP2D", err)
	}
	tx.Id = id
	// 设置签名
	sig, err := args.FromAccount.Sign(id)
	if err != nil {
		return nil, utils.WrapError("newTxP2D", err)
	}
	tx.Sig = sig
	return tx, nil
}

/*******************************************************实现接口*********************************************************/


// TypeNo 获取交易类型编号
func (tx *TxP2D) TypeNo() uint {
	return TX_P2D
}

// Id 对于已生成的交易，获取其ID
func (tx *TxP2D) ID() crypto.Hash {
	return tx.Id
}

// Hash 计算交易哈希值，作为交易ID
func (tx *TxP2D) Hash() (hash crypto.Hash, err error) {
	txCopy := *tx
	txCopy.Id, txCopy.Sig = crypto.Hash{}, crypto.Signature{}
	var res []byte
	if res, err = txCopy.Serialize(); err != nil {
		return crypto.Hash{}, utils.WrapError("TxP2D_Hash", err)
	}
	hash1 := sha256.Sum256(res)
	return hash1[:], nil
}

// Serialize 交易序列化为字节切片
func (tx *TxP2D) Serialize() (result []byte, err error) {
	return utils.GobEncode(tx)
}

// String 转换为字符串，用于打印输出
func (tx *TxP2D) String() string {
	return utils.JsonMarshalIndentToString(tx)
}

// Deserialize 反序列化，必须提前 tx := &TxP2D{} 再调用
func (tx *TxP2D) Deserialize(p2dBytes []byte) (err error) {
	// 防止非空TxP2D调用该方法改变了自身内容

	// 反序列化
	var buf bytes.Buffer
	buf.Write(p2dBytes)
	err = gob.NewDecoder(&buf).Decode(tx)
	if err != nil {
		return utils.WrapError("TxP2D_Deserialize", err)
	}
	return nil
}

// IsValid 验证交易是否合乎规则
func (tx *TxP2D) IsValid() (err error) {

	// 检查交易时间有效性
	if tx.Time >= common.TimeStamp(time.Now().Unix()) {
		return utils.WrapError("TxP2D_IsValid", ErrWrongTime)
	}

	// 检查to id有效性和账号是否可用
	if err = tx.To.IsValid(account.Single, account.Doctor); err != nil {
		return utils.WrapError("TxP2D_IsValid", err)
	}

	// 检查fromID的有效性、可用性和from余额是否足够,from签名是否匹配
	if err = tx.From.IsValid(account.Single, account.Patient); err != nil {
		return utils.WrapError("TxP2D_IsValid", err)
	}

	// TODO： PurchaseTarget可用性检查。这部分交给交易双方自己做，除非达到仲裁条件，由验证节点进行仲裁才会再上层的handleTX方法中去处理

	// 验证交易ID是不是正确设置
	txHash, _ := tx.Hash()
	if string(txHash) != string(tx.Id) {
		return utils.WrapError("TxP2D_IsValid", ErrWrongTxId)
	}

	return nil
}

/*******************************************************实现接口*********************************************************/
