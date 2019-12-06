package tx



import (
"bytes"
"crypto/sha256"
"encoding/gob"
"time"
)

/*********************************************************************************************************************
                                                    TxCoinbase相关
*********************************************************************************************************************/

// TxCoinbaseArgs 新建交易函数newTxCoinbase()的传参
type TxCoinbaseArgs struct {
	//	BaseArgs
	To          UserID
	Amount      Coin
	Description string
}

// CheckArgsValue 检查参数值是否合规
func (args *TxCoinbaseArgs) CheckArgsValue(gsm *GlobalStateMachine) (err error) {
	// 检查 to 的有效性
	if valid, _ := args.To.IsValid(); !valid {
		return ErrInvalidUserID
	}
	// coinbase交易只允许出块节点构建，而出块节点的roleNo 0~9
	if args.To.RoleNo > 9 {
		return ErrInvalidUserID
	}

	// 检查 amount 有效性
	// TODO: 检查coinbase奖励是否合乎规则

	// TODO: 检查 description 格式，以及代码注入？

	return nil
}

// TxCoinbase 出块奖励交易，只允许A类账户接收，A类账户目前包括医院H和第三方研究机构R
// 由于coinbase交易没有转账者，且必须由出块者构建，所以不设置签名项划定归属。
type TxCoinbase struct {
	BaseTransaction `json:"baseTransaction"`
}

// newTxCoinbase 新建出块奖励交易。
func newTxCoinbase(args *TxCoinbaseArgs) (tx *TxCoinbase, err error) {
	// TODO： 注意： 参数的检查交给gsm去做了
	//// 检验参数
	//if err = args.CheckArgsValue(); err != nil {
	//	return nil, utils.WrapError("newTxCoinbase", err)
	//}

	// 构造tx
	tx = &TxCoinbase{
		BaseTransaction{
			ID:          Hash{},
			Time:        UnixTimeStamp(time.Now().Unix()),
			To:          args.To,
			Amount:      args.Amount,
			Description: args.Description,
		},
	}

	// 设置Id
	id, err := tx.Hash()
	if err != nil {
		return nil, WrapError("newTxCoinbase", err)
	}
	tx.ID = id
	return tx, nil
}

// TypeNo 获取交易类型编号
func (tx *TxCoinbase) TypeNo() uint {
	return TX_COINBASE
}

// Id 对于已生成的交易，获取其ID
func (tx *TxCoinbase) Id() Hash {
	return tx.ID
}

// Hash 计算交易哈希值，作为交易ID
func (tx *TxCoinbase) Hash() (hash Hash, err error) {
	txCopy := *tx
	txCopy.ID = Hash{}
	var res []byte
	if res, err = txCopy.Serialize(); err != nil {
		return Hash{}, WrapError("TxCoinbase_Hash", err)
	}
	hash1 := sha256.Sum256(res)
	return hash1[:], nil
}

// Serialize 交易序列化为字节切片
func (tx *TxCoinbase) Serialize() (result []byte, err error) {
	return GobEncode(tx)
}

// String 转换为字符串，用于打印输出
func (tx *TxCoinbase) String() string {
	type TxCoinbaseForPrint struct {
		ID          []byte          `json:"id"`
		Time        string `json:"time"`
		To          UserID        `json:"to"`
		Amount      Coin          `json:"amount"`
		Description string        `json:"description"`
	}
	txPrint := &TxCoinbaseForPrint{
		ID:          tx.ID[:],
		Time:        time.Unix(int64(tx.Time), 0).Format("2006/01/02 15:04:05"),
		To:          tx.To,
		Amount:      tx.Amount,
		Description: tx.Description,
	}
	return JsonMarshalIndentToString(txPrint)
}

// Deserialize 反序列化，必须提前 tx := &TxCoinbase{} 再调用
func (tx *TxCoinbase) Deserialize(data []byte) (err error) {
	// 防止非空TxR2P调用该方法改变了自身内容

	// 反序列化
	var buf bytes.Buffer
	buf.Write(data)
	err = gob.NewDecoder(&buf).Decode(tx)
	if err != nil {
		return WrapError("TxCoinbase_Deserialize", err)
	}
	return nil
}

// IsValid 验证交易是否合乎规则
func (tx *TxCoinbase) IsValid(gsm *GlobalStateMachine) (err error) {

	/*	tx = &TxCoinbase{
		BaseTransaction:BaseTransaction{
			ID:Hash{},
			Time:UnixTimeStamp(0),
			To:UserID{},
			Amount:Coin(1),
			Description:string(""),
		}}*/

	// 要记住检验交易有两种情况下被调用：一是加入未打包交易池之前要检查交易（情况A）；二是收到区块后要检查区块内交易（情况B）。

	// 检查时间戳是否比现在早（至于是不是早太多就不检查了，早太多的话余额那里是不会给过的）（情况A）； 时间戳是否比区块时间早（情况B）
	// 但是要注意情况A调用检查一定比情况B早，所以只要满足情况A就一定满足情况B (或者说，如果情况A不通过，也就不会进入到情况B检查)。所以，只检查情况A就好
	if tx.Time >= UnixTimeStamp(time.Now().Unix()) {
		return WrapError("TxCoinbase_IsValid", ErrWrongTimeTX)
	}

	// 检查coinbase接收者ID的有效性和角色的权限与可用性
	userIDValid, _ := tx.To.IsValid() // 另起一个变量userIDValid，避免阅读时被误导而已。
	if !userIDValid {
		return WrapError("TxCoinbase_IsValid", ErrInvalidUserID)
	}
	if tx.To.RoleNo >= 10 {
		return WrapError("TxCoinbase_IsValid", ErrNoCoinbasePermitRole)
	}
	toEcoinAccount, ok := gsm.Accounts.Map[tx.To.ID]
	if !ok {
		return WrapError("TxCoinbase_IsValid", ErrNonexistentUserID)
	}
	if !toEcoinAccount.Available() {
		return WrapError("TxCoinbase_IsValid", ErrUnavailableUserID)
	}

	// 检查coinbase金额
	if tx.Amount != toEcoinAccount.Role().CoinbaseReward() {
		return WrapError("TxCoinbase_IsValid", ErrWrongCoinbaseReward)
	}

	// 验证交易ID是不是正确设置
	txHash, _ := tx.Hash()
	if string(txHash) != string(tx.ID) {
		return WrapError("TxCoinbase_IsValid", ErrWrongTXID)
	}

	// TODO： Coinbase还有一个检查点：其由出块节点构造，但在验证过程中必须检查是不是填了出块节点账户。因此在出块节点检查区块时需要有一个区块的检查方法
	// 而这个方法检查所有交易有效性，并对coinbase（在打包交易时始终放在交易列表第一位）再增加这一个处理。
	// 如果要在这里做这个检查，就必须穿入一个*Block作参数。但是其他类型交易不需要这个参数，会破坏整体接口的实现。

	return nil
}

