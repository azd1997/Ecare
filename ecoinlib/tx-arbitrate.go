package ecoin

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"time"
)

/*********************************************************************************************************************
                                                    TxArbitrate相关
*********************************************************************************************************************/

// TxArbitrateArgs 新建交易函数newTxArbitrate()的传参
type TxArbitrateArgs struct {
	//	BaseArgs
	Arbitrator       *Account
	ArbitratorID     UserID
	TargetTXBytes    []byte
	TargetTXComplete bool
	Description      string
}

// CheckArgsValue 检查参数值是否合规
func (args *TxArbitrateArgs) CheckArgsValue(gsm *GlobalStateMachine) (err error) {
	// 检查from? 不需要，因为就是往上给account调用的

	// 检查targetTXBytes
	var targetTX CommercialTX // 这里不需要额外检查targetTXBytes空还是不空，反正会报错
	//if err = targetTX.Deserialize(args.TargetTXBytes); err != nil { // TODO: 直接通过未赋值的接口调用方法会怎样？假如有多个结构体实现了该接口。答案：会发生空指针调用，panic
	//	return ErrNotTxBytes
	//}
	if targetTX, err = DeserializeCommercialTX(args.TargetTXBytes); err != nil {
		return ErrNotTxBytes
	}
	// 检查targetTX是否在未完成交易池中
	targetTXID, err := targetTX.Hash()
	if err != nil {
		return err
	}
	if _, ok := gsm.UCTXP.Map[string(targetTXID)]; !ok {
		return ErrTXNotInUCTXP
	}

	// TODO: 检查 targetTXComplete 有效性。由上层去做

	// TODO: 检查 description 格式，以及代码注入？

	// 参数有效
	return nil
}

// 仲裁交易，针对商业性质交易如TxR2P的“三次僵持”提出的交易体
type TxArbitrate struct {
	ID   Hash          `json:"id"`
	Time UnixTimeStamp `json:"time"`
	// TargetTx 仲裁目标
	TargetTXBytes []byte `json:"targetTXBytes"`

	// ArbitrateResult    []byte        `json:"arbitrateResult"`

	// TargetTXComplete 目标交易是否完成，true表示完成，转账生效，否则退回
	TargetTXComplete bool `json:"targetTXComplete"`
	// Description 描述，可用来附加信息
	Description string `json:"description"`
	// Arbitrator 仲裁者
	Arbitrator UserID    `json:"arbitrator"`
	Sig        Signature `json:"sig"`
}

// newTxD2P 新建D2P转账交易(P2D交易二段)。
func newTxArbitrate(args *TxArbitrateArgs) (tx *TxArbitrate, err error) {
	//// 检验参数
	//if err = args.CheckArgsValue(); err != nil {
	//	return nil, utils.WrapError("newTxArbitrate", err)
	//}
	//
	//// 获取仲裁者UserID
	//arbitratorID, err := args.Arbitrator.UserID(args.Gsm.opts.ChecksumLength(), args.Gsm.opts.Version())
	//if err != nil {
	//	return nil, utils.WrapError("newTxArbitrate", err)
	//}

	// 构造tx
	tx = &TxArbitrate{
		ID:               Hash{},
		Time:             UnixTimeStamp(time.Now().Unix()),
		TargetTXBytes:    args.TargetTXBytes,
		TargetTXComplete: args.TargetTXComplete,
		Description:      args.Description,
		Arbitrator:       args.ArbitratorID,
		Sig:              Signature{},
	}

	// 设置Id
	id, err := tx.Hash()
	if err != nil {
		return nil, WrapError("newTxArbitrate", err)
	}
	tx.ID = id
	// 设置签名
	sig, err := args.Arbitrator.Sign(id)
	if err != nil {
		return nil, WrapError("newTxArbitrate", err)
	}
	tx.Sig = sig
	return tx, nil
}

// TypeNo 获取交易类型编号
func (tx *TxArbitrate) TypeNo() uint {
	return TX_ARBITRATE
}

// Id 对于已生成的交易，获取其ID
func (tx *TxArbitrate) Id() Hash {
	return tx.ID
}

// Hash 计算交易哈希值，作为交易ID
func (tx *TxArbitrate) Hash() (hash Hash, err error) {
	txCopy := *tx
	txCopy.ID, txCopy.Sig = Hash{}, Signature{}
	var res []byte
	if res, err = txCopy.Serialize(); err != nil {
		return Hash{}, WrapError("TxArbitrate_Hash", err)
	}
	hash1 := sha256.Sum256(res)
	return hash1[:], nil
}

// Serialize 交易序列化为字节切片
func (tx *TxArbitrate) Serialize() (result []byte, err error) {
	return GobEncode(tx)
}

// String 转换为字符串，用于打印输出
func (tx *TxArbitrate) String() string {
	type TxArbitrateForPrint struct {
		ID          []byte          `json:"id"`
		Time        string `json:"time"`
		TargetTXBytes []byte `json:"targetTXBytes"`
		TargetTXComplete bool `json:"targetTXComplete"`
		// Description 描述，可用来附加信息
		Description string `json:"description"`
		// Arbitrator 仲裁者
		Arbitrator UserID    `json:"arbitrator"`
		Sig        Signature `json:"sig"`
	}
	txPrint := &TxArbitrateForPrint{
		ID:          tx.ID,
		Time:        time.Unix(int64(tx.Time), 0).Format("2006/01/02 15:04:05"),
		TargetTXBytes:tx.TargetTXBytes,
		TargetTXComplete:tx.TargetTXComplete,
		Description: tx.Description,
		Arbitrator:tx.Arbitrator,
		Sig:tx.Sig,
	}
 	return JsonMarshalIndentToString(txPrint)
}

// Deserialize 反序列化，必须提前 tx := &TxArbitrate{} 再调用
func (tx *TxArbitrate) Deserialize(txAtbitrateBytes []byte) (err error) {
	// 防止非空TxArbitrate调用该方法改变了自身内容

	// 反序列化
	var buf bytes.Buffer
	buf.Write(txAtbitrateBytes)
	err = gob.NewDecoder(&buf).Decode(tx)
	if err != nil {
		return WrapError("TxArbitrate_Deserialize", err)
	}
	return nil
}

// IsValid 验证交易是否合乎规则
func (tx *TxArbitrate) IsValid(gsm *GlobalStateMachine) (err error) {

	/*	tx = &TxArbitrate{
		ID:          Hash{},
		Time:        UnixTimeStamp(time.Now().Unix()),
		TargetTXBytes:    targetTXBytes,
		TargetTXComplete:    targetTXComplete,
		Description: description,
		Arbitrator:arbitratorID,
		Sig:         Signature{},
	}*/

	// 检查交易时间有效性
	if tx.Time >= UnixTimeStamp(time.Now().Unix()) {
		return WrapError("TxArbitrate_IsValid", ErrWrongTimeTX)
	}

	// 检查arbitratorID的有效性、可用性、角色权限和from签名是否匹配
	userIDValid, _ := tx.Arbitrator.IsValid()
	if !userIDValid {
		return WrapError("TxArbitrate_IsValid", ErrInvalidUserID)
	}
	arbitratorEcoinAccount, ok := gsm.Accounts.Map[tx.Arbitrator.ID]
	if !ok {
		return WrapError("TxArbitrate_IsValid", ErrNonexistentUserID)
	}
	if !arbitratorEcoinAccount.Available() {
		return WrapError("TxArbitrate_IsValid", ErrUnavailableUserID)
	}
	if arbitratorEcoinAccount.Role().No() >= 10 {
		return WrapError("TxArbitrate_IsValid", ErrNoCoinbasePermitRole)
	}
	if !VerifySignature(tx.ID[:], tx.Sig, arbitratorEcoinAccount.PubKey()) {
		return WrapError("TxArbitrate_IsValid", ErrInconsistentSignature)
	}

	// TODO： 仲裁结果验证，这里不进行，丢给上层调用函数HandleTX去做。

	// 检查前部交易是不是一个未完成的商业性质交易，为空则错误；不为空必须是符合商业性质交易体且交易ID在未完成交易池中，否则认为是不合法交易
	if tx.TargetTXBytes == nil || bytes.Compare(tx.TargetTXBytes, []byte{}) == 0 {
		return WrapError("TxArbitrate_IsValid", ErrEmptySoureTX)
	}
	// 其实可以把这个判断条件去掉，但是算了
	if bytes.Compare(tx.TargetTXBytes, []byte{}) != 0 {
		// 反序列化出商业交易
		var prevTx CommercialTX
		prevTx, err = DeserializeCommercialTX(tx.TargetTXBytes)
		if err != nil {
			return WrapError("TxArbitrate_IsValid", err)
		}
		// 获取商业交易ID
		txId, err := prevTx.Hash()
		if err != nil {
			return WrapError("TxArbitrate_IsValid", err)
		}

		if _, ok := gsm.UCTXP.Map[string(txId)]; !ok {
			return WrapError("TxArbitrate_IsValid", ErrNotUncompletedTX)
		}
	}

	// 验证交易ID是不是正确设置
	txHash, _ := tx.Hash()
	if string(txHash) != string(tx.ID) {
		return WrapError("TxArbitrate_IsValid", ErrWrongTXID)
	}

	return nil
}