package tx

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"github.com/azd1997/Ecare/ecoin/crypto"
	"github.com/azd1997/Ecare/ecoin/utils"
	"time"

	"github.com/azd1997/Ecare/ecoin/account"
	"github.com/azd1997/Ecare/ecoin/common"
)

// TxCoinbase 出块奖励交易，只允许A类账户接收，A类账户目前包括医院H和第三方研究机构R
// 由于coinbase交易没有转账者，且必须由出块者构建，所以不设置签名项划定归属。
type TxCoinbase struct {
//	TxBase
	Id          crypto.Hash      `json:"id"`
	Time        common.TimeStamp `json:"time"`
	To          account.UserId   `json:"to"`
	Amount      common.Coin      `json:"amount"`
	Description string           `json:"description"`
}

// newTxCoinbase 新建出块奖励交易。
func newTxCoinbase(args *CoinbaseArgs) (tx *TxCoinbase, err error) {

	// 构造tx
	tx = &TxCoinbase{}

	// 设置Id
	id, err := tx.Hash()
	if err != nil {
		return nil, utils.WrapError("newTxCoinbase", err)
	}
	tx.Id = id
	return tx, nil
}

/*******************************************************实现接口*********************************************************/

// TypeNo 获取交易类型编号
func (tx *TxCoinbase) TypeNo() uint {
	return TX_COINBASE
}

// Id 对于已生成的交易，获取其ID
func (tx *TxCoinbase) ID() crypto.Hash {
	return tx.Id
}

// Hash 计算交易哈希值，作为交易ID
func (tx *TxCoinbase) Hash() (hash crypto.Hash, err error) {
	txCopy := *tx
	txCopy.Id = crypto.Hash{}
	var res []byte
	if res, err = txCopy.Serialize(); err != nil {
		return crypto.Hash{}, utils.WrapError("TxCoinbase_Hash", err)
	}
	hash1 := sha256.Sum256(res)
	return hash1[:], nil
}

// Serialize 交易序列化为字节切片
func (tx *TxCoinbase) Serialize() (result []byte, err error) {
	return utils.GobEncode(tx)
}

// String 转换为字符串，用于打印输出
func (tx *TxCoinbase) String() string {
	return utils.JsonMarshalIndentToString(tx)
}

// Deserialize 反序列化，必须提前 tx := &TxCoinbase{} 再调用
func (tx *TxCoinbase) Deserialize(data []byte) (err error) {
	// 防止非空TxR2P调用该方法改变了自身内容

	// 反序列化
	var buf bytes.Buffer
	buf.Write(data)
	err = gob.NewDecoder(&buf).Decode(tx)
	if err != nil {
		return utils.WrapError("TxCoinbase_Deserialize", err)
	}
	return nil
}

// IsValid 验证交易是否合乎规则
func (tx *TxCoinbase) IsValid() (err error) {

	// 要记住检验交易有两种情况下被调用：一是加入未打包交易池之前要检查交易（情况A）；二是收到区块后要检查区块内交易（情况B）。

	// 检查时间戳是否比现在早（至于是不是早太多就不检查了，早太多的话余额那里是不会给过的）（情况A）； 时间戳是否比区块时间早（情况B）
	// 但是要注意情况A调用检查一定比情况B早，所以只要满足情况A就一定满足情况B (或者说，如果情况A不通过，也就不会进入到情况B检查)。所以，只检查情况A就好
	if tx.Time >= common.TimeStamp(time.Now().Unix()) {
		return utils.WrapError("TxCoinbase_IsValid", ErrWrongTime)
	}

	// 检查 To
	if err = tx.To.IsValid(account.A, 0); err != nil {
		return utils.WrapError("TxCoinbase_IsValid", err)
	}

	// 验证交易ID是不是正确设置
	txHash, _ := tx.Hash()
	if string(txHash) != string(tx.Id) {
		return utils.WrapError("TxCoinbase_IsValid", ErrWrongTxId)
	}

	// TODO： Coinbase还有一个检查点：其由出块节点构造，但在验证过程中必须检查是不是填了出块节点账户。因此在出块节点检查区块时需要有一个区块的检查方法
	// 而这个方法检查所有交易有效性，并对coinbase（在打包交易时始终放在交易列表第一位）再增加这一个处理。
	// 如果要在这里做这个检查，就必须穿入一个*Block作参数。但是其他类型交易不需要这个参数，会破坏整体接口的实现。

	return nil
}

/*******************************************************实现接口*********************************************************/
