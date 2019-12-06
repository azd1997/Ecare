package eaccount

import (
	"github.com/azd1997/Ecare/ecoin/account"
	"github.com/azd1997/Ecare/ecoin/common"
	"github.com/azd1997/Ecare/ecoin/utils"
)


// EAccount 每个区块链账户的公开信息。
// EcoinAccount太长，缩写为EAccount
// account.go中Account指完全权限的账户信息，含有私钥信息，又因为全局状态机中会维护这些公私钥以外的信息，所以Account只有公私钥信息
type EAccount struct {
	UserId         account.UserId `json:"userId"`
	PubKeyField    account.PubKey `json:"pubKey"`
	BalanceCoin    common.Coin    `json:"balance"`
	RoleField      common.Role    `json:"role"` // 存储角色信息，和UserID中都存了一次RoleNo。必须保证相同
	AvailableField bool           `json:"available"`
	RegisterInfo   RegisterInfo   `json:"registerInfo"`
}

// String 打印方法
func (a *EAccount) String() string {
	return utils.JsonMarshalIndentToString(a)
}

// PubKey 获取账户公钥
func (a *EAccount) PubKey() []byte {
	return a.PubKeyField
}

// Balance 获取余额
func (a *EAccount) Balance() common.Coin {
	return a.BalanceCoin
}

// Role 获取账户的角色
func (a *EAccount) Role() *common.Role {
	return &a.RoleField
}

// Available 账户是否可用
func (a *EAccount) Available() bool {
	return a.AvailableField
}
