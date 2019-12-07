package tx

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"github.com/azd1997/Ecare/ecoin/account"
	"github.com/azd1997/Ecare/ecoin/common"
	"github.com/azd1997/Ecare/ecoin/utils"
	"time"
)

// 仲裁交易，针对商业性质交易如TxR2P的“三次僵持”提出的交易体
type TxArbitrate struct {
	Id   common.Hash      `json:"id"`
	Time common.TimeStamp `json:"time"`
	// TargetTx 仲裁目标
	TargetTX CommercialTX `json:"targetTX"`

	// TargetTXErr 仲裁时发现的目标交易中存在的问题，只记录是双方某人作恶及作恶形式
	// 一旦出现三次僵持，必然有一方作恶，不管是有意还是无意。真正的仲裁函数需要根据不同情况处理，
	// 得到具体的错误信息，然后所有共识节点根据错误信息对交易双方作出惩处
	// 仲裁结果中这个错误是不会为空的。空值为仲裁异常；
	// 除非仲裁异常，否则目标交易被按照情况结束
	// 仲裁异常，则仲裁者受到惩处，目标交易在下一轮POT由新的出块者也就是仲裁者仲裁
	TargetTXErr error `json:"targetTXErr"`

	// TargetTXComplete 目标交易是否完成，true表示完成，转账生效，否则退回
	// 这个字段相当于仲裁者替目标交易作出的决定
	//TargetTXComplete bool `json:"targetTXComplete"`

	// Description 描述，可用来附加信息
	Description string `json:"description"`

	// Arbitrator 仲裁者
	Arbitrator account.UserId   `json:"arbitrator"`
	Sig        common.Signature `json:"sig"`
}

// newTxArbitrate 新建仲裁交易。
func newTxArbitrate(args *ArbitrateArgs) (tx *TxArbitrate, err error) {

	// 构造tx
	tx = &TxArbitrate{
		Id:          nil,
		Time:        common.TimeStamp(time.Now().Unix()),
		TargetTX:    args.TargetTX,
		TargetTXErr: args.TargetTXErr,
		Description: args.Description,
		Arbitrator:  args.Arbitrator,
		Sig:         nil,
	}

	// 设置Id
	id, err := tx.Hash()
	if err != nil {
		return nil, utils.WrapError("newTxArbitrate", err)
	}
	tx.Id = id
	// 设置签名
	sig, err := args.ArbitratorAccount.Sign(id)
	if err != nil {
		return nil, utils.WrapError("newTxArbitrate", err)
	}
	tx.Sig = sig
	return tx, nil
}

/*******************************************************实现接口*********************************************************/

// TypeNo 获取交易类型编号
func (tx *TxArbitrate) TypeNo() uint {
	return TX_ARBITRATE
}

// Id 对于已生成的交易，获取其ID
func (tx *TxArbitrate) ID() common.Hash {
	return tx.Id
}

// Hash 计算交易哈希值，作为交易ID
func (tx *TxArbitrate) Hash() (hash common.Hash, err error) {
	txCopy := *tx
	txCopy.Id, txCopy.Sig = common.Hash{}, common.Signature{}
	var res []byte
	if res, err = txCopy.Serialize(); err != nil {
		return common.Hash{}, utils.WrapError("TxArbitrate_Hash", err)
	}
	hash1 := sha256.Sum256(res)
	return hash1[:], nil
}

// Serialize 交易序列化为字节切片
func (tx *TxArbitrate) Serialize() (result []byte, err error) {
	return utils.GobEncode(tx)
}

// String 转换为字符串，用于打印输出
func (tx *TxArbitrate) String() string {
	return utils.JsonMarshalIndentToString(tx)
}

// Deserialize 反序列化，必须提前 tx := &TxArbitrate{} 再调用
func (tx *TxArbitrate) Deserialize(txAtbitrateBytes []byte) (err error) {
	// 防止非空TxArbitrate调用该方法改变了自身内容

	// 反序列化
	var buf bytes.Buffer
	buf.Write(txAtbitrateBytes)
	err = gob.NewDecoder(&buf).Decode(tx)
	if err != nil {
		return utils.WrapError("TxArbitrate_Deserialize", err)
	}
	return nil
}

// IsValid 验证交易是否合乎规则
func (tx *TxArbitrate) IsValid() (err error) {

	// 检查交易时间有效性
	if tx.Time >= common.TimeStamp(time.Now().Unix()) {
		return utils.WrapError("TxArbitrate_IsValid", ErrWrongTime)
	}

	// 检查arbitratorID的有效性、可用性、角色权限和from签名是否匹配
	if err = tx.Arbitrator.IsValid(account.A, 0); err != nil {
		return utils.WrapError("TxArbitrate_IsValid", err)
	}

	// 目标交易不能为空。 至于目标交易更多的验证不在这里做
	if tx.TargetTX == nil {
		return utils.WrapError("TxArbitrate_IsValid", ErrNilSourceTx)
	}

	// 验证交易ID是不是正确设置
	txHash, _ := tx.Hash()
	if string(txHash) != string(tx.Id) {
		return utils.WrapError("TxArbitrate_IsValid", ErrWrongTxId)
	}

	return nil
}

/*******************************************************实现接口*********************************************************/
