package ecoin

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"time"
)

/*********************************************************************************************************************
                                                    TxP2D相关
*********************************************************************************************************************/

// TxP2DArgs 新建交易函数newTxP2D()的传参
type TxP2DArgs struct {
	//	BaseArgs
	From           *Account
	FromID         UserID
	To             UserID
	Amount         Coin
	Description    string
	PurchaseTarget TargetData
	Storage        DataStorage
}

// CheckArgsValue 检查参数值是否合规
func (args *TxP2DArgs) CheckArgsValue(gsm *GlobalStateMachine) (err error) {
	// 检查from? 不需要，因为就是往上给account调用的

	// 检查 to 的有效性
	if valid, _ := args.To.IsValid(); !valid {
		return ErrInvalidUserID
	}
	// 检查to账号是不是医生（11）
	if args.To.RoleNo != 11 {
		return ErrWrongRoleUserID
	}

	// 检查 amount 有效性(余额是否足够)
	selfId, err := args.From.UserID()
	if err != nil {
		return err
	}
	if args.Amount > gsm.Accounts.Map[selfId.ID].Balance() {
		return ErrNotSufficientBalance
	}

	// TODO: 检查 description 格式，以及代码注入？

	// 检查storage是否有效
	if !args.Storage.IsOk() {
		return ErrNotOkStorage
	}

	// 检查 purchaseTarget是否存在？
	if ok, _ := args.PurchaseTarget.IsOk(args.Storage); !ok {
		return ErrNonexistentTargetData
	}

	// 参数有效
	return nil
}

// TxP2D 病人向下班医生发起的心电诊断交易，阶段一		TODO: 暂时只支持找指定医生诊断；后边考虑广播交易等待医生解决
type TxP2D struct {
	BaseTransaction `json:"baseTransaction"`
	From            UserID     `json:"from"`
	Sig             Signature  `json:"sig"`
	PurchaseTarget  TargetData `json:"purchaseTarget"`
}

// newTxP2D 新建P2D转账交易。
func newTxP2D(args *TxP2DArgs) (tx *TxP2D, err error) {

	//// 检验参数
	//if err = args.CheckArgsValue(); err != nil {
	//	return nil, utils.WrapError("newTxP2D", err)
	//}
	//
	//// 获取转账者UserID
	//fromID, err := args.From.UserID(args.Gsm.opts.ChecksumLength(), args.Gsm.opts.Version())
	//if err != nil {
	//	return nil, utils.WrapError("newTxP2D", err)
	//}

	// 构造tx
	tx = &TxP2D{
		BaseTransaction: BaseTransaction{
			ID:          Hash{},
			Time:        UnixTimeStamp(time.Now().Unix()),
			To:          args.To,
			Amount:      args.Amount,
			Description: args.Description,
		},
		From:           args.FromID,
		Sig:            Signature{},
		PurchaseTarget: args.PurchaseTarget,
	}

	// 设置Id
	id, err := tx.Hash()
	if err != nil {
		return nil, WrapError("newTxP2D", err)
	}
	tx.ID = id
	// 设置签名
	sig, err := args.From.Sign(id)
	if err != nil {
		return nil, WrapError("newTxP2D", err)
	}
	tx.Sig = sig
	return tx, nil
}

// TypeNo 获取交易类型编号
func (tx *TxP2D) TypeNo() uint {
	return TX_P2D
}

// Id 对于已生成的交易，获取其ID
func (tx *TxP2D) Id() Hash {
	return tx.ID
}

// Hash 计算交易哈希值，作为交易ID
func (tx *TxP2D) Hash() (hash Hash, err error) {
	txCopy := *tx
	txCopy.ID, txCopy.Sig = Hash{}, Signature{}
	var res []byte
	if res, err = txCopy.Serialize(); err != nil {
		return Hash{}, WrapError("TxP2D_Hash", err)
	}
	hash1 := sha256.Sum256(res)
	return hash1[:], nil
}

// Serialize 交易序列化为字节切片
func (tx *TxP2D) Serialize() (result []byte, err error) {
	return GobEncode(tx)
}

// String 转换为字符串，用于打印输出
func (tx *TxP2D) String() string {

	type TxP2DForPrint struct {
		ID          []byte          `json:"id"`
		Time        string `json:"time"`
		From        UserID        `json:"from"`
		To          UserID        `json:"to"`
		Description string        `json:"description"`
		Sig         Signature     `json:"sig"`
		PurchaseTarget TargetData `json:"purchaseTarget"`
	}
	txPrint := &TxP2DForPrint{
		ID:          tx.ID[:],
		Time:        time.Unix(int64(tx.Time), 0).Format("2006/01/02 15:04:05"),
		From:tx.From,
		To:tx.To,
		Description: tx.Description,
		Sig:tx.Sig,
		PurchaseTarget:tx.PurchaseTarget,
	}

	return JsonMarshalIndentToString(txPrint)
}

// Deserialize 反序列化，必须提前 tx := &TxP2D{} 再调用
func (tx *TxP2D) Deserialize(p2dBytes []byte) (err error) {
	// 防止非空TxP2D调用该方法改变了自身内容

	// 反序列化
	var buf bytes.Buffer
	buf.Write(p2dBytes)
	err = gob.NewDecoder(&buf).Decode(tx)
	if err != nil {
		return WrapError("TxP2D_Deserialize", err)
	}
	return nil
}

// IsValid 验证交易是否合乎规则
func (tx *TxP2D) IsValid(gsm *GlobalStateMachine) (err error) {

	/*	tx = &TxP2D{
		BaseTransaction: BaseTransaction{
			Id:          Hash{},
			Time:        UnixTimeStamp(time.Now().Unix()),
			To:          to,
			Amount:      amount,
			Description: description,
		},
		From:           fromID,
		Sig:            Signature{},
		PurchaseTarget: purchaseTarget,
	}*/

	// 检查交易时间有效性
	if tx.Time >= UnixTimeStamp(time.Now().Unix()) {
		return WrapError("TxP2D_IsValid", ErrWrongTimeTX)
	}

	// 检查to id有效性和账号是否可用
	userIDValid, _ := tx.To.IsValid() // 另起一个变量userIDValid，避免阅读时被误导而已。
	if !userIDValid {
		return WrapError("TxP2D_IsValid", ErrInvalidUserID)
	}
	toEcoinAccount, ok := gsm.Accounts.Map[tx.To.ID]
	if !ok {
		return WrapError("TxP2D_IsValid", ErrNonexistentUserID)
	}
	if !toEcoinAccount.Available() {
		return WrapError("TxP2D_IsValid", ErrUnavailableUserID)
	}

	// 检查fromID的有效性、可用性和from余额是否足够,from签名是否匹配
	userIDValid, _ = tx.From.IsValid()
	if !userIDValid {
		return WrapError("TxP2D_IsValid", ErrInvalidUserID)
	}
	fromEcoinAccount, ok := gsm.Accounts.Map[tx.From.ID]
	if !ok {
		return WrapError("TxP2D_IsValid", ErrNonexistentUserID)
	}
	if !fromEcoinAccount.Available() {
		return WrapError("TxP2D_IsValid", ErrUnavailableUserID)
	}
	if tx.Amount > fromEcoinAccount.Balance() {
		return WrapError("TxP2D_IsValid", ErrNotSufficientBalance)
	}
	if !VerifySignature(tx.ID[:], tx.Sig, fromEcoinAccount.PubKey()) {
		return WrapError("TxP2D_IsValid", ErrInconsistentSignature)
	}

	// TODO： PurchaseTarget可用性检查。这部分交给交易双方自己做，除非达到仲裁条件，由验证节点进行仲裁才会再上层的handleTX方法中去处理

	// 验证交易ID是不是正确设置
	txHash, _ := tx.Hash()
	if string(txHash) != string(tx.ID) {
		return WrapError("TxP2D_IsValid", ErrWrongTXID)
	}

	return nil
}
