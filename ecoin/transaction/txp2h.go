package tx

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"github.com/azd1997/Ecare/ecoin/utils"
	"time"

	"github.com/azd1997/Ecare/ecoin/account"
	"github.com/azd1997/Ecare/ecoin/common"
	"github.com/azd1997/Ecare/ecoin/storage"
)



// TxP2H 病人向医院发起的心电数据诊断，分人工和机器自动分析两种。阶段一
type TxP2H struct {
	Id             common.Hash        `json:"id"`
	Time           common.TimeStamp   `json:"time"`
	From           account.UserId     `json:"from"`
	To             account.UserId     `json:"to"`
	Amount         common.Coin        `json:"amount"`
	PurchaseTarget storage.TargetData `json:"purchaseTarget"`
	PurchaseType    uint8              `json:"purchaseType"` // Auto/Doctor 0/1
	Description    string             `json:"description"`
	Sig            common.Signature   `json:"sig"`
}

// newTxP2H 新建P2H转账交易。
func newTxP2H(args *P2HArgs) (tx *TxP2H, err error) {

	//// 检验参数
	//if err = args.CheckArgsValue(); err != nil {
	//	return nil, utils.WrapError("newTxP2H", err)
	//}
	//
	//// 获取转账者UserID
	//fromID, err := args.From.UserID(args.Gsm.opts.ChecksumLength(), args.Gsm.opts.Version())
	//if err != nil {
	//	return nil, utils.WrapError("newTxP2H", err)
	//}

	// 构造tx
	tx = &TxP2H{
		Id:             nil,
		Time:           common.TimeStamp(time.Now().Unix()),
		From:           args.From,
		To:             args.To,
		Amount:         args.Amount,
		PurchaseTarget: args.PurchaseTarget,
		PurchaseType:   args.PurchaseType,
		Description:    args.Description,
		Sig:            nil,
	}

	// 设置Id
	id, err := tx.Hash()
	if err != nil {
		return nil, utils.WrapError("newTxP2H", err)
	}
	tx.Id = id
	// 设置签名
	sig, err := args.FromAccount.Sign(id)
	if err != nil {
		return nil, utils.WrapError("newTxP2H", err)
	}
	tx.Sig = sig
	return tx, nil
}

// TypeNo 获取交易类型编号
func (tx *TxP2H) TypeNo() uint {
	return TX_P2H
}

// Id 对于已生成的交易，获取其ID
func (tx *TxP2H) ID() common.Hash {
	return tx.Id
}

// Hash 计算交易哈希值，作为交易ID
func (tx *TxP2H) Hash() (hash common.Hash, err error) {
	txCopy := *tx
	txCopy.Id, txCopy.Sig = common.Hash{}, common.Signature{}
	var res []byte
	if res, err = txCopy.Serialize(); err != nil {
		return common.Hash{}, utils.WrapError("TxP2H_Hash", err)
	}
	hash1 := sha256.Sum256(res)
	return hash1[:], nil
}

// Serialize 交易序列化为字节切片
func (tx *TxP2H) Serialize() (result []byte, err error) {
	return utils.GobEncode(tx)
}

// String 转换为字符串，用于打印输出
func (tx *TxP2H) String() string {
	return utils.JsonMarshalIndentToString(tx)
}

// Deserialize 反序列化，必须提前 tx := &TxP2H{} 再调用
func (tx *TxP2H) Deserialize(p2hBytes []byte) (err error) {
	// 防止非空TxR2P调用该方法改变了自身内容

	// 反序列化
	var buf bytes.Buffer
	buf.Write(p2hBytes)
	err = gob.NewDecoder(&buf).Decode(tx)
	if err != nil {
		return utils.WrapError("TxP2H_Deserialize", err)
	}
	return nil
}

// IsValid 验证交易是否合乎规则
func (tx *TxP2H) IsValid() (err error) {

	// 检查交易时间有效性
	if tx.Time >= common.TimeStamp(time.Now().Unix()) {
		return utils.WrapError("TxP2H_IsValid", ErrWrongTimeTX)
	}

	// 检查to id有效性和账号是否可用
	if err = tx.To.IsValid(account.Single, account.Hospital); err != nil {
		return utils.WrapError("Args_Check", err)
	}

	// 检查fromID的有效性、可用性和from余额是否足够,from签名是否匹配
	if err = tx.From.IsValid(account.Single, account.Patient); err != nil {
		return utils.WrapError("Args_Check", err)
	}

	// TODO： PurchaseTarget可用性检查。这部分交给交易双方自己做，除非达到仲裁条件，由验证节点进行仲裁才会再上层的handleTX方法中去处理

	// 检查purchaseType
	if tx.PurchaseType != ECG_DIAG_AUTO && tx.PurchaseType != ECG_DIAG_DOCTOR {
		return utils.WrapError("TxP2H_IsValid", ErrUnknownPurchaseType)
	}

	// 验证交易ID是不是正确设置
	txHash, _ := tx.Hash()
	if string(txHash) != string(tx.Id) {
		return utils.WrapError("TxP2H_IsValid", ErrWrongTXID)
	}

	return nil
}


