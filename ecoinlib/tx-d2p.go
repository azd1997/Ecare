package ecoin

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"time"
)

/*********************************************************************************************************************
                                                    TxD2P相关
*********************************************************************************************************************/

// TxD2PArgs 新建交易函数newTxD2P()的传参
type TxD2PArgs struct {
	//	BaseArgs
	From        *Account
	FromID      UserID
	P2DBytes    []byte
	Response    []byte
	Description string
}

// CheckArgsValue 检查参数值是否合规
func (args *TxD2PArgs) CheckArgsValue(gsm *GlobalStateMachine) (err error) {
	// 检查from? 不需要，因为就是往上给account调用的

	// 检查p2dBytes
	p2d := &TxP2D{}
	if err = p2d.Deserialize(args.P2DBytes); err != nil {
		return ErrNotTxBytes
	}
	// 检查p2h是否在未完成交易池中
	if _, ok := gsm.UCTXP.Map[string(p2d.ID)]; !ok {
		return ErrTXNotInUCTXP
	}
	// 检查p2h内to是否和此时的from对应，都是本机拥有的账户
	selfId, err := args.From.UserID()
	if err != nil {
		return err
	}
	if selfId != p2d.To {
		return ErrUnmatchedTxReceiver
	}

	// TODO: 检查 response 有效性

	// TODO: 检查 description 格式，以及代码注入？

	// 参数有效
	return nil
}

// TxP2D 病人向下班医生发起的心电诊断交易，阶段一
type TxD2P struct {
	ID          Hash          `json:"id"`
	Time        UnixTimeStamp `json:"time"`
	From        UserID        `json:"from"`
	P2DBytes    []byte        `json:"p2dBytes"`
	Response    []byte        `json:"response"` // 比如说请求数据的密码
	Description string        `json:"description"`
	Sig         Signature     `json:"sig"`
}

// newTxD2P 新建D2P转账交易(P2D交易二段)。
func newTxD2P(args *TxD2PArgs) (tx *TxD2P, err error) {
	//// 检验参数
	//if err = args.CheckArgsValue(); err != nil {
	//	return nil, utils.WrapError("newTxD2P", err)
	//}
	//
	//// 获取转账者UserID
	//fromID, err := args.From.UserID(args.Gsm.opts.ChecksumLength(), args.Gsm.opts.Version())
	//if err != nil {
	//	return nil, utils.WrapError("newTxD2P", err)
	//}

	// 构造tx
	tx = &TxD2P{
		ID:          Hash{},
		Time:        UnixTimeStamp(time.Now().Unix()),
		From:        args.FromID,
		P2DBytes:    args.P2DBytes,
		Response:    args.Response,
		Description: args.Description,
		Sig:         Signature{},
	}

	// 设置Id
	id, err := tx.Hash()
	if err != nil {
		return nil, WrapError("newTxD2P", err)
	}
	tx.ID = id
	// 设置签名
	sig, err := args.From.Sign(id)
	if err != nil {
		return nil, WrapError("newTxD2P", err)
	}
	tx.Sig = sig
	return tx, nil
}

// TypeNo 获取交易类型编号
func (tx *TxD2P) TypeNo() uint {
	return TX_D2P
}

// Id 对于已生成的交易，获取其ID
func (tx *TxD2P) Id() Hash {
	return tx.ID
}

// Hash 计算交易哈希值，作为交易ID
func (tx *TxD2P) Hash() (hash Hash, err error) {
	txCopy := *tx
	txCopy.ID, txCopy.Sig = Hash{}, Signature{}
	var res []byte
	if res, err = txCopy.Serialize(); err != nil {
		return Hash{}, WrapError("TxD2P_Hash", err)
	}
	hash1 := sha256.Sum256(res)
	return hash1[:], nil
}

// Serialize 交易序列化为字节切片
func (tx *TxD2P) Serialize() (result []byte, err error) {
	return GobEncode(tx)
}

// String 转换为字符串，用于打印输出
func (tx *TxD2P) String() string {

	type TxD2PForPrint struct {
		ID          []byte          `json:"id"`
		Time        string `json:"time"`
		From        UserID        `json:"from"`
		P2DBytes    []byte        `json:"p2dBytes"`
		Response    []byte        `json:"response"` // 比如说请求数据的密码
		Description string        `json:"description"`
		Sig         Signature     `json:"sig"`
	}
	txPrint := &TxD2PForPrint{
		ID:          tx.ID[:],
		Time:        time.Unix(int64(tx.Time), 0).Format("2006/01/02 15:04:05"),
		From:tx.From,
		P2DBytes:tx.P2DBytes,
		Response:tx.Response,
		Description: tx.Description,
		Sig:tx.Sig,
	}

	return JsonMarshalIndentToString(txPrint)
}

// Deserialize 反序列化，必须提前 tx := &TxD2P{} 再调用
func (tx *TxD2P) Deserialize(d2pBytes []byte) (err error) {
	// 防止非空TxR2P调用该方法改变了自身内容

	// 反序列化
	var buf bytes.Buffer
	buf.Write(d2pBytes)
	err = gob.NewDecoder(&buf).Decode(tx)
	if err != nil {
		return WrapError("TxD2P_Deserialize", err)
	}
	return nil
}

// IsValid 验证交易是否合乎规则
func (tx *TxD2P) IsValid(gsm *GlobalStateMachine) (err error) {

	/*	tx = &TxD2P{
		Id:          Hash{},
		Time:        UnixTimeStamp(time.Now().Unix()),
		From:fromID,
		P2DBytes:    p2dBytes,
		Response:    response,
		Description: description,
		Sig:         Signature{},
	}*/

	// 检查交易时间有效性
	if tx.Time >= UnixTimeStamp(time.Now().Unix()) {
		return WrapError("TxD2P_IsValid", ErrWrongTimeTX)
	}

	// 检查fromID的有效性、可用性和from签名是否匹配
	userIDValid, _ := tx.From.IsValid()
	if !userIDValid {
		return WrapError("TxD2P_IsValid", ErrInvalidUserID)
	}
	fromEcoinAccount, ok := gsm.Accounts.Map[tx.From.ID]
	if !ok {
		return WrapError("TxD2P_IsValid", ErrNonexistentUserID)
	}
	if !fromEcoinAccount.Available() {
		return WrapError("TxD2P_IsValid", ErrUnavailableUserID)
	}
	if !VerifySignature(tx.ID[:], tx.Sig, fromEcoinAccount.PubKey()) {
		return WrapError("TxD2P_IsValid", ErrInconsistentSignature)
	}

	// TODO： Response可用性检查。这部分交给交易双方自己做，除非达到仲裁条件，由验证节点进行仲裁才会再上层的handleTX方法中去处理

	// 检查前部交易是不是一个P2D交易，为空则错误；不为空必须是符合P2D交易体且交易ID在未完成交易池中，否则认为是不合法交易
	if tx.P2DBytes == nil || bytes.Compare(tx.P2DBytes, []byte{}) == 0 {
		return WrapError("TxD2P_IsValid", ErrEmptySoureTX)
	}
	if bytes.Compare(tx.P2DBytes, []byte{}) != 0 {
		prevTx := &TxP2D{}
		err := prevTx.Deserialize(tx.P2DBytes)
		if err != nil {
			return WrapError("TxD2P_IsValid", err)
		}
		if _, ok := gsm.UCTXP.Map[string(prevTx.ID)]; !ok {
			return WrapError("TxD2P_IsValid", ErrNotUncompletedTX)
		}
	}

	// 验证交易ID是不是正确设置
	txHash, _ := tx.Hash()
	if string(txHash) != string(tx.ID) {
		return WrapError("TxD2P_IsValid", ErrWrongTXID)
	}

	return nil
}

