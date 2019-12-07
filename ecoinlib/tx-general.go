package ecoin

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"time"
)

/*********************************************************************************************************************
                                                    TxGeneral相关
*********************************************************************************************************************/

// TxGeneralArgs 新建交易函数newTxGeneral()的传参
type TxGeneralArgs struct {
	//	BaseArgs
	From        *Account
	FromID      UserID
	To          UserID
	Amount      Coin
	Description string
}

// CheckArgsValue 检查参数值是否合规
func (args *TxGeneralArgs) CheckArgsValue(gsm *GlobalStateMachine) (err error) {
	// 检查from? 不需要，因为就是往上给account调用的

	// 检查FromID
	fromID, err := args.From.UserID()
	if err != nil {
		return err
	}
	if args.FromID != fromID {
		return ErrWrongArguments
	}

	// 检查 to 的有效性
	if valid, _ := args.To.IsValid(); !valid {
		return ErrInvalidUserID
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

	return nil
}

// TxGeneral 通用交易， 一方转给另一方，无需确认
type TxGeneral struct {
	BaseTransaction `json:"baseTransaction"`
	From            UserID    `json:"from"`
	Sig             Signature `json:"sig"`
}

// newTxGeneral 新建普通转账交易。
func newTxGeneral(args *TxGeneralArgs) (tx *TxGeneral, err error) {
	//// 检验参数
	//if err = args.CheckArgsValue(); err != nil {
	//	return nil, utils.WrapError("newTxGeneral", err)
	//}

	//// 获取转账者UserID
	//fromID, err := args.From.UserID(args.Gsm.opts.ChecksumLength(), args.Gsm.opts.Version())
	//if err != nil {
	//	return nil, utils.WrapError("newTxGeneral", err)
	//}

	// 构造tx
	tx = &TxGeneral{
		BaseTransaction: BaseTransaction{
			ID:          Hash{},
			Time:        UnixTimeStamp(time.Now().Unix()),
			To:          args.To,
			Amount:      args.Amount,
			Description: args.Description,
		},
		From: args.FromID,
		Sig:  Signature{},
	}

	// 设置Id
	id, err := tx.Hash()
	if err != nil {
		return nil, WrapError("newTxGeneral", err)
	}
	tx.ID = id
	// 设置签名
	sig, err := args.From.Sign(id)
	if err != nil {
		return nil, WrapError("newTxGeneral", err)
	}
	tx.Sig = sig
	return tx, nil
}

// TypeNo 获取交易类型编号
func (tx *TxGeneral) TypeNo() uint {
	return TX_GENERAL
}

// Id 对于已生成的交易，获取其ID
func (tx *TxGeneral) Id() Hash {
	return tx.ID
}

// Hash 计算交易哈希值，作为交易ID
func (tx *TxGeneral) Hash() (hash Hash, err error) {
	txCopy := *tx
	txCopy.ID, txCopy.Sig = Hash{}, Signature{} // 置空值
	var res []byte
	if res, err = txCopy.Serialize(); err != nil {
		return Hash{}, WrapError("TxGeneral_Hash", err)
	}
	hash1 := sha256.Sum256(res)
	return hash1[:], nil
}

// Serialize 交易序列化为字节切片
func (tx *TxGeneral) Serialize() (result []byte, err error) {
	return GobEncode(tx)
}

// String 转换为字符串，用于打印输出
func (tx *TxGeneral) String() string {

	type TxGeneralForPrint struct {
		ID          []byte          `json:"id"`
		Time        string `json:"time"`
		From        UserID        `json:"from"`
		To UserID `json:"to"`
		Amount Coin `json:"amount"`
		Description string        `json:"description"`
		Sig         Signature     `json:"sig"`
	}
	txPrint := &TxGeneralForPrint{
		ID:          tx.ID[:],
		Time:        time.Unix(int64(tx.Time), 0).Format("2006/01/02 15:04:05"),
		From:tx.From,
		To:tx.To,
		Amount:tx.Amount,
		Description: tx.Description,
		Sig:tx.Sig,
	}

	return JsonMarshalIndentToString(txPrint)
}

// Deserialize 反序列化，必须提前 tx := &TxGeneral{} 再调用
func (tx *TxGeneral) Deserialize(data []byte) (err error) {
	// 防止非空TxR2P调用该方法改变了自身内容

	// 反序列化
	var buf bytes.Buffer
	buf.Write(data)
	err = gob.NewDecoder(&buf).Decode(tx)
	if err != nil {
		return WrapError("TxGeneral_Deserialize", err)
	}
	return nil
}

// IsValid 验证交易是否合乎规则
func (tx *TxGeneral) IsValid(gsm *GlobalStateMachine) (err error) {

	/*	tx = &TxGeneral{
			BaseTransaction: BaseTransaction{
				Id:          Hash{},
				Time:        UnixTimeStamp(time.Now().Unix()),
				To:          to,
				Amount:      amount,
				Description: description,
			},
			From: fromID,
			Sig:  Signature{},
		}
	*/

	// 检查交易时间有效性
	if tx.Time >= UnixTimeStamp(time.Now().Unix()) {
		return WrapError("TxGeneral_IsValid", ErrWrongTimeTX)
	}

	// 检查to id有效性和账号是否可用
	userIDValid, _ := tx.To.IsValid() // 另起一个变量userIDValid，避免阅读时被误导而已。
	if !userIDValid {
		return WrapError("TxGeneral_IsValid", ErrInvalidUserID)
	}
	toEcoinAccount, ok := gsm.Accounts.Map[tx.To.ID]
	if !ok {
		return WrapError("TxGeneral_IsValid", ErrNonexistentUserID)
	}
	if !toEcoinAccount.Available() {
		return WrapError("TxGeneral_IsValid", ErrUnavailableUserID)
	}

	// 检查fromID的有效性、可用性和from余额是否足够,from签名是否匹配
	userIDValid, _ = tx.From.IsValid()
	if !userIDValid {
		return WrapError("TxGeneral_IsValid", ErrInvalidUserID)
	}
	fromEcoinAccount, ok := gsm.Accounts.Map[tx.From.ID]
	if !ok {
		return WrapError("TxGeneral_IsValid", ErrNonexistentUserID)
	}
	if !fromEcoinAccount.Available() {
		return WrapError("TxGeneral_IsValid", ErrUnavailableUserID)
	}
	if tx.Amount > fromEcoinAccount.Balance() {
		return WrapError("TxGeneral_IsValid", ErrNotSufficientBalance)
	}
	if !VerifySignature(tx.ID[:], tx.Sig, fromEcoinAccount.PubKey()) {
		return WrapError("TxGeneral_IsValid", ErrInconsistentSignature)
	}

	// 验证交易ID是不是正确设置
	txHash, _ := tx.Hash()
	if string(txHash) != string(tx.ID) {
		return WrapError("TxGeneral_IsValid", ErrWrongTXID)
	}

	return nil
}
