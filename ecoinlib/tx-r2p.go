package ecoin

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"time"
)

/*********************************************************************************************************************
                                                    TxR2P相关
*********************************************************************************************************************/

// TxR2PArgs 新建交易函数newTxR2P()的传参
type TxR2PArgs struct {
	//	BaseArgs
	From           *Account
	FromID         UserID
	To             UserID
	Amount         Coin
	Description    string
	PurchaseTarget TargetData
	P2RBytes       []byte
	TxComplete     bool
	Storage        DataStorage
}

// CheckArgsValue 检查参数值是否合规
func (args *TxR2PArgs) CheckArgsValue(gsm *GlobalStateMachine) (err error) {
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

	// 检查storage是否有效
	if !args.Storage.IsOk() {
		return ErrNotOkStorage
	}

	// 检查 purchaseTarget是否存在？
	if ok, _ := args.PurchaseTarget.IsOk(args.Storage); !ok {
		return ErrNonexistentTargetData
	}

	// 检验p2rBytes。 要么是[]byte{}(表示是初始交易)，要么是可以反序列化为p2r交易
	if bytes.Compare(args.P2RBytes, []byte{}) != 0 {
		p2r := TxP2R{}
		if err = p2r.Deserialize(args.P2RBytes); err != nil {
			return ErrWrongSourceTX
		}
		// 检查p2r.ID是否在未完成交易池
		if _, ok := gsm.UCTXP.Map[string(p2r.ID)]; !ok {
			return ErrTXNotInUCTXP
		}
		// 检查p2r.From是否为args.To
		if p2r.From != args.To {
			return ErrUnmatchedTxReceiver
		}
	}

	// 参数有效
	return nil
}

// TxR2P 第三方研究机构向病人发起的数据交易的阶段一交易
type TxR2P struct {
	BaseTransaction `json:"baseTransaction"`
	From            UserID     `json:"from"`
	Sig             Signature  `json:"sig"`
	PurchaseTarget  TargetData `json:"purchaseTarget"`
	P2RBytes        []byte     `json:"p2rBytes, omitempty"`
	TxComplete      bool       `json:"txComplete"` // 注意：在上层调用也就是block类中验证交易时，需要检查txComplete来进行“三次僵持“策略的实现
}

// newTxR2P 新建R2P转账交易。
func newTxR2P(args *TxR2PArgs) (tx *TxR2P, err error) {

	//// 检验参数
	//if err = args.CheckArgsValue(); err != nil {
	//	return nil, utils.WrapError("newTxR2P", err)
	//}
	//
	//// 获取转账者UserID
	//fromID, err := args.From.UserID(args.Gsm.opts.ChecksumLength(), args.Gsm.opts.Version())
	//if err != nil {
	//	return nil, utils.WrapError("newTxR2P", err)
	//}

	// 构造tx
	tx = &TxR2P{
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
		P2RBytes:       args.P2RBytes,
		TxComplete:     args.TxComplete,
	}

	// 设置Id
	id, err := tx.Hash()
	if err != nil {
		return nil, WrapError("newTxR2P", err)
	}
	tx.ID = id
	// 设置签名
	sig, err := args.From.Sign(id)
	if err != nil {
		return nil, WrapError("newTxR2P", err)
	}
	tx.Sig = sig
	return tx, nil
}

// commercial 商业性质
func (tx *TxR2P) commercial() {
	// 啥事也不干
}

// TypeNo 获取交易类型编号
func (tx *TxR2P) TypeNo() uint {
	return TX_R2P
}

// Id 对于已生成的交易，获取其ID
func (tx *TxR2P) Id() Hash {
	return tx.ID
}

// Hash 计算交易哈希值，作为交易ID
func (tx *TxR2P) Hash() (hash Hash, err error) {
	txCopy := *tx
	txCopy.ID, txCopy.Sig = Hash{}, Signature{}
	var res []byte
	if res, err = txCopy.Serialize(); err != nil {
		return Hash{}, WrapError("TxGeneral_Hash", err)
	}
	hash1 := sha256.Sum256(res)
	return hash1[:], nil
}

// Serialize 交易序列化为字节切片
func (tx *TxR2P) Serialize() (result []byte, err error) {
	return GobEncode(tx)
}

// String 转换为字符串，用于打印输出
func (tx *TxR2P) String() string {

	type TxR2PForPrint struct {
		ID          []byte          `json:"id"`
		Time        string `json:"time"`
		From        UserID        `json:"from"`
		To          UserID        `json:"to"`
		Description string        `json:"description"`
		Sig         Signature     `json:"sig"`
		PurchaseTarget TargetData `json:"purchaseTarget"`
		P2RBytes        []byte     `json:"p2rBytes, omitempty"`
		TxComplete      bool       `json:"txComplete"`
	}
	txPrint := &TxR2PForPrint{
		ID:          tx.ID[:],
		Time:        time.Unix(int64(tx.Time), 0).Format("2006/01/02 15:04:05"),
		From:tx.From,
		To:tx.To,
		Description: tx.Description,
		Sig:tx.Sig,
		PurchaseTarget:tx.PurchaseTarget,
		P2RBytes:tx.P2RBytes,
		TxComplete:tx.TxComplete,
	}

	return JsonMarshalIndentToString(txPrint)
}

// Deserialize 反序列化，必须提前 tx := &TxR2P{} 再调用
func (tx *TxR2P) Deserialize(r2pBytes []byte) (err error) {
	// 防止非空TxR2P调用该方法改变了自身内容

	// 反序列化
	var buf bytes.Buffer
	buf.Write(r2pBytes)
	err = gob.NewDecoder(&buf).Decode(tx)
	if err != nil {
		return WrapError("TxR2P_Deserialize", err)
	}
	return nil
}

// IsValid 验证交易是否合乎规则
func (tx *TxR2P) IsValid(gsm *GlobalStateMachine) (err error) {

	/*	tx = &TxR2P{
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
		P2RBytes:       p2rBytes,
		TxComplete:     txComplete,
	}*/

	// 检查交易时间有效性
	if tx.Time >= UnixTimeStamp(time.Now().Unix()) {
		return WrapError("TxR2P_IsValid", ErrWrongTimeTX)
	}

	// 检查to id有效性和账号是否可用
	userIDValid, _ := tx.To.IsValid() // 另起一个变量userIDValid，避免阅读时被误导而已。
	if !userIDValid {
		return WrapError("TxR2P_IsValid", ErrInvalidUserID)
	}
	toEcoinAccount, ok := gsm.Accounts.Map[tx.To.ID]
	if !ok {
		return WrapError("TxR2P_IsValid", ErrNonexistentUserID)
	}
	if !toEcoinAccount.Available() {
		return WrapError("TxR2P_IsValid", ErrUnavailableUserID)
	}

	// 检查fromID的有效性、可用性和from余额是否足够,from签名是否匹配
	userIDValid, _ = tx.From.IsValid()
	if !userIDValid {
		return WrapError("TxR2P_IsValid", ErrInvalidUserID)
	}
	fromEcoinAccount, ok := gsm.Accounts.Map[tx.From.ID]
	if !ok {
		return WrapError("TxR2P_IsValid", ErrNonexistentUserID)
	}
	if !fromEcoinAccount.Available() {
		return WrapError("TxR2P_IsValid", ErrUnavailableUserID)
	}
	if tx.Amount > fromEcoinAccount.Balance() {
		return WrapError("TxR2P_IsValid", ErrNotSufficientBalance)
	}
	if VerifySignature(tx.ID[:], tx.Sig, fromEcoinAccount.PubKey()) {
		return WrapError("TxR2P_IsValid", ErrInconsistentSignature)
	}

	// TODO： PurchaseTarget可用性检查。这部分交给交易双方自己做，除非达到仲裁条件，由验证节点进行仲裁才会再上层的handleTX方法中去处理

	// 检查前部交易是不是一个P2R交易，为空则正确；不为空必须是符合P2R交易体且交易ID在未完成交易池中，否则认为是不合法交易
	if tx.P2RBytes == nil || bytes.Compare(tx.P2RBytes, []byte{}) == 0 {
		return WrapError("TxR2P_IsValid", ErrEmptySoureTX)
	}
	if bytes.Compare(tx.P2RBytes, []byte{}) != 0 {
		prevTx := &TxP2R{}
		err := prevTx.Deserialize(tx.P2RBytes)
		if err != nil {
			return WrapError("TxR2P_IsValid", err)
		}
		if _, ok := gsm.UCTXP.Map[string(prevTx.ID)]; !ok {
			return WrapError("TxR2P_IsValid", ErrNotUncompletedTX)
		}
	}

	// 验证交易ID是不是正确设置
	txHash, _ := tx.Hash()
	if string(txHash) != string(tx.ID) {
		return WrapError("TxR2P_IsValid", ErrWrongTXID)
	}

	return nil
}

