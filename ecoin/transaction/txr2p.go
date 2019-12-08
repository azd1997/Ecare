package tx

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"github.com/azd1997/Ecare/ecoin/crypto"
	"time"

	"github.com/azd1997/Ecare/ecoin/account"
	"github.com/azd1997/Ecare/ecoin/common"
	"github.com/azd1997/Ecare/ecoin/storage"
	"github.com/azd1997/Ecare/ecoin/utils"
)

// TxR2P 第三方研究机构向病人发起的数据交易的阶段一交易
type TxR2P struct {
	Id   crypto.Hash      `json:"id"`
	Time common.TimeStamp `json:"time"`
	From account.UserId   `json:"from"`
	To   account.UserId   `json:"to"` // 为了避免复杂，后续只能是原来的To否则无效
	Sig  crypto.Signature `json:"sig"`

	PurchaseTarget storage.TargetData `json:"purchaseTarget"` // 非第一次发时一般置空（表示目标不变），否则也可以进行更新
	P2R            *TxP2R             `json:"p2r"`
	ResponseInfo   []byte             `json:"response"` // 比如说请求数据的密码
	Description    string             `json:"description"`
	TxComplete     bool               `json:"txComplete"` // 注意：在上层调用也就是block类中验证交易时，需要检查txComplete来进行“三次僵持“策略的实现
}

// newTxR2P 新建R2P转账交易。
func newTxR2P(args *R2PArgs) (tx *TxR2P, err error) {

	// 构造tx
	tx = &TxR2P{
		Id:             nil,
		Time:           common.TimeStamp(time.Now().Unix()),
		From:           args.From,
		To:             args.To,
		P2R:            args.P2R,
		ResponseInfo:   args.Response,
		Description:    args.Description,
		Sig:            nil,
		PurchaseTarget: args.PurchaseTarget,
		TxComplete:     args.TxComplete,
	}

	// 设置Id
	id, err := tx.Hash()
	if err != nil {
		return nil, utils.WrapError("newTxR2P", err)
	}
	tx.Id = id
	// 设置签名
	sig, err := args.FromAccount.Sign(id)
	if err != nil {
		return nil, utils.WrapError("newTxR2P", err)
	}
	tx.Sig = sig
	return tx, nil
}

/***********************************************实现CommercialTX接口***************************************************/

// commercial 商业性质
func (tx *TxR2P) commercial() {
	// 啥事也不干
}

/*****************************************************实现TX接口*********************************************************/

// TypeNo 获取交易类型编号
func (tx *TxR2P) TypeNo() uint {
	return TX_R2P
}

// Id 对于已生成的交易，获取其ID
func (tx *TxR2P) ID() crypto.Hash {
	return tx.Id
}

// Hash 计算交易哈希值，作为交易ID
func (tx *TxR2P) Hash() (hash crypto.Hash, err error) {
	txCopy := *tx
	txCopy.Id, txCopy.Sig = crypto.Hash{}, crypto.Signature{}
	var res []byte
	if res, err = txCopy.Serialize(); err != nil {
		return crypto.Hash{}, utils.WrapError("TxGeneral_Hash", err)
	}
	hash1 := sha256.Sum256(res)
	return hash1[:], nil
}

// Serialize 交易序列化为字节切片
func (tx *TxR2P) Serialize() (result []byte, err error) {
	return utils.GobEncode(tx)
}

// String 转换为字符串，用于打印输出
func (tx *TxR2P) String() string {
	return utils.JsonMarshalIndentToString(tx)
}

// Deserialize 反序列化，必须提前 tx := &TxR2P{} 再调用
func (tx *TxR2P) Deserialize(r2pBytes []byte) (err error) {
	// 防止非空TxR2P调用该方法改变了自身内容

	// 反序列化
	var buf bytes.Buffer
	buf.Write(r2pBytes)
	err = gob.NewDecoder(&buf).Decode(tx)
	if err != nil {
		return utils.WrapError("TxR2P_Deserialize", err)
	}
	return nil
}

//
//func (tx *TxR2P) Response() []byte {
//	return tx.ResponseInfo
//}

// IsValid 验证交易是否合乎规则
func (tx *TxR2P) IsValid() (err error) {

	// 检查交易时间有效性
	if tx.Time >= common.TimeStamp(time.Now().Unix()) {
		return utils.WrapError("TxR2P_IsValid", ErrWrongTime)
	}

	// 检查 From
	if err = tx.From.IsValid(account.Single, account.ResearchInstitution); err != nil {
		return utils.WrapError("TxR2P_IsValid", err)
	}

	// 检查 To
	if err = tx.To.IsValid(account.Single, account.Patient); err != nil {
		return utils.WrapError("TxR2P_IsValid", err)
	}

	// 检查to与P2R的from是否匹配
	if tx.P2R != nil {
		if tx.To != tx.P2R.From {
			return utils.WrapError("TxR2P_IsValid", ErrUnmatchedReceiver)
		}
		if tx.From != tx.P2R.R2P.From { // 只要tx.P2R非空，那么一定满足该条件，不会panic。
			return utils.WrapError("TxR2P_IsValid", ErrUnmatchedSender)
		}
	}

	// 检查 purchaseTarget是否有效？
	if err = tx.PurchaseTarget.IsOk(); err != nil {
		return utils.WrapError("TxR2P_IsValid", err)
	}

	// 验证交易ID是不是正确设置
	txHash, _ := tx.Hash()
	if string(txHash) != string(tx.Id) {
		return utils.WrapError("TxR2P_IsValid", ErrWrongTxId)
	}

	return nil
}

/*******************************************************实现接口*********************************************************/
