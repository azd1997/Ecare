package ecoin

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"time"
)

/*********************************************************************************************************************
                                                    TxP2R相关
*********************************************************************************************************************/

// TxP2RArgs 新建交易函数newTxP2R()的传参
type TxP2RArgs struct {
	//	BaseArgs
	From        *Account
	FromID      UserID
	R2PBytes    []byte
	Response    []byte
	Description string
}

// CheckArgsValue 检查参数值是否合规
func (args *TxP2RArgs) CheckArgsValue(gsm *GlobalStateMachine) (err error) {
	// 检查from? 不需要，因为就是往上给account调用的

	// 检查FromID
	fromID, err := args.From.UserID()
	if err != nil {
		return err
	}
	if args.FromID != fromID {
		return ErrWrongArguments
	}

	// 检查r2pBytes
	r2p := &TxR2P{}
	if err = r2p.Deserialize(args.R2PBytes); err != nil {
		return ErrNotTxBytes
	}
	// 检查r2p是否在未完成交易池中
	if _, ok := gsm.UCTXP.Map[string(r2p.ID)]; !ok {
		return ErrTXNotInUCTXP
	}
	// 检查r2p内to是否和此时的from对应，都是本机拥有的账户
	selfId, err := args.From.UserID()
	if err != nil {
		return err
	}
	if selfId != r2p.To {
		return ErrUnmatchedTxReceiver
	}

	// TODO: 检查 response 有效性

	// TODO: 检查 description 格式，以及代码注入？

	// 参数有效
	return nil
}

// TxP2R 第三方研究机构向病人发起的数据交易的阶段二交易
type TxP2R struct {
	ID          Hash          `json:"id"`
	Time        UnixTimeStamp `json:"time"`
	From        UserID        `json:"from"`
	R2PBytes    []byte        `json:"r2pBytes"`
	Response    []byte        `json:"response"` // 比如说请求数据的密码
	Description string        `json:"description"`
	Sig         Signature     `json:"sig"`
}

// newTxP2R 新建P2R转账交易(R2P交易二段)。
func newTxP2R(args *TxP2RArgs) (tx *TxP2R, err error) {
	//// 检验参数
	//if err = args.CheckArgsValue(); err != nil {
	//	return nil, utils.WrapError("newTxP2R", err)
	//}
	//
	//// 获取转账者UserID
	//fromID, err := args.From.UserID(args.Gsm.opts.ChecksumLength(), args.Gsm.opts.Version())
	//if err != nil {
	//	return nil, utils.WrapError("newTxP2R", err)
	//}

	// 构造tx
	tx = &TxP2R{
		ID:          Hash{},
		Time:        UnixTimeStamp(time.Now().Unix()),
		From:        args.FromID,
		R2PBytes:    args.R2PBytes,
		Response:    args.Response,
		Description: args.Description,
		Sig:         Signature{},
	}

	// 设置Id
	id, err := tx.Hash()
	if err != nil {
		return nil, WrapError("newTxP2R", err)
	}
	tx.ID = id
	// 设置签名
	sig, err := args.From.Sign(id)
	if err != nil {
		return nil, WrapError("newTxP2R", err)
	}
	tx.Sig = sig
	return tx, nil
}

// TypeNo 获取交易类型编号
func (tx *TxP2R) TypeNo() uint {
	return TX_P2R
}

// Id 对于已生成的交易，获取其ID
func (tx *TxP2R) Id() Hash {
	return tx.ID
}

// Hash 计算交易哈希值，作为交易ID
func (tx *TxP2R) Hash() (hash Hash, err error) {
	txCopy := *tx
	txCopy.ID, txCopy.Sig = Hash{}, Signature{}
	var res []byte
	if res, err = txCopy.Serialize(); err != nil {
		return Hash{}, WrapError("TxP2R_Hash", err)
	}
	hash1 := sha256.Sum256(res)
	return hash1[:], nil
}

// Serialize 交易序列化为字节切片
func (tx *TxP2R) Serialize() (result []byte, err error) {
	return GobEncode(tx)
}

// String 转换为字符串，用于打印输出
func (tx *TxP2R) String() string {

	type TxP2RForPrint struct {
		ID          []byte          `json:"id"`
		Time        string `json:"time"`
		From        UserID        `json:"from"`
		R2PBytes    []byte        `json:"r2pBytes"`
		Response    []byte        `json:"response"` // 比如说请求数据的密码
		Description string        `json:"description"`
		Sig         Signature     `json:"sig"`
	}
	txPrint := &TxP2RForPrint{
		ID:          tx.ID[:],
		Time:        time.Unix(int64(tx.Time), 0).Format("2006/01/02 15:04:05"),
		From:tx.From,
		R2PBytes:tx.R2PBytes,
		Response:tx.Response,
		Description: tx.Description,
		Sig:tx.Sig,
	}

	return JsonMarshalIndentToString(txPrint)
}

// Deserialize 反序列化，必须提前 tx := &TxR2P{} 再调用
func (tx *TxP2R) Deserialize(p2rBytes []byte) (err error) {
	// 防止非空TxR2P调用该方法改变了自身内容

	// 反序列化
	var buf bytes.Buffer
	buf.Write(p2rBytes)
	err = gob.NewDecoder(&buf).Decode(tx)
	if err != nil {
		return WrapError("TxP2R_Deserialize", err)
	}
	return nil
}

// IsValid 验证交易是否合乎规则
func (tx *TxP2R) IsValid(gsm *GlobalStateMachine) (err error) {

	/*	tx = &TxP2R{
		Id:          Hash{},
		Time:        UnixTimeStamp(time.Now().Unix()),
		From:fromID,
		R2PBytes:    r2pBytes,
		Response:    response,
		Description: description,
		Sig:         Signature{},
	}*/

	// 检查交易时间有效性
	if tx.Time >= UnixTimeStamp(time.Now().Unix()) {
		return WrapError("TxP2R_IsValid", ErrWrongTimeTX)
	}

	// 检查fromID的有效性、可用性和from签名是否匹配
	userIDValid, _ := tx.From.IsValid()
	if !userIDValid {
		return WrapError("TxP2R_IsValid", ErrInvalidUserID)
	}
	fromEcoinAccount, ok := gsm.Accounts.Map[tx.From.ID]
	if !ok {
		return WrapError("TxP2R_IsValid", ErrNonexistentUserID)
	}
	if !fromEcoinAccount.Available() {
		return WrapError("TxP2R_IsValid", ErrUnavailableUserID)
	}
	if !VerifySignature(tx.ID[:], tx.Sig, fromEcoinAccount.PubKey()) {
		return WrapError("TxP2R_IsValid", ErrInconsistentSignature)
	}

	// TODO： Response可用性检查。这部分交给交易双方自己做，除非达到仲裁条件，由验证节点进行仲裁才会再上层的handleTX方法中去处理

	// 检查前部交易是不是一个R2P交易，为空则正确；不为空必须是符合R2P交易体且交易ID在未完成交易池中，否则认为是不合法交易
	if tx.R2PBytes == nil || bytes.Compare(tx.R2PBytes, []byte{}) == 0 {
		return WrapError("TxP2R_IsValid", ErrEmptySoureTX)
	}
	if bytes.Compare(tx.R2PBytes, []byte{}) != 0 {
		prevTx := &TxR2P{}
		err := prevTx.Deserialize(tx.R2PBytes)
		if err != nil {
			return WrapError("TxP2R_IsValid", err)
		}
		if _, ok := gsm.UCTXP.Map[string(prevTx.ID)]; !ok {
			return WrapError("TxP2R_IsValid", ErrNotUncompletedTX)
		}
	}

	// 验证交易ID是不是正确设置
	txHash, _ := tx.Hash()
	if string(txHash) != string(tx.ID) {
		return WrapError("TxP2R_IsValid", ErrWrongTXID)
	}

	return nil
}
