package ecoinlib

import "errors"

// 每一个引用ecoinlib的地方只能使用这个变量，而不能新建
var EcoinWorld = newGlobalStateMachine()

type Balance uint

type ecoinAccount struct {
	userID    UserID
	pubKey    []byte
	balance   Balance
	role      Role
	available bool
}

// 全局状态机存储每个账户的可公开的信息：
// 包括余额、角色（角色中定义了其币相关的规则）、可用状态、交易历史（区块链）
type globalStateMachine struct {
	accounts map[UserID]*ecoinAccount
	ledger   Chain
	logger   Logger
	opts     Option
}

func newGlobalStateMachine() *globalStateMachine {
	return &globalStateMachine{
		accounts: map[UserID]*ecoinAccount{},
		ledger:   Chain{},
	}
}

/*balance相关*/
// 获取账户地址的余额，不作账户地址存在与否的检查，如不存在，返回默认值0，存在也有可能返回0
func (gsm *globalStateMachine) GetBalanceOfUserID(id UserID) Balance {
	return gsm.accounts[id].balance
}

// 更新余额表，单次单个地址
func (gsm *globalStateMachine) UpdateBalanceOfUserID(id UserID, newBalance Balance) {
	gsm.accounts[id].balance = newBalance
}

// 批量更新余额表
func (gsm *globalStateMachine) UpdateBalanceOfUserIDs(ids []UserID, newBalances []Balance) error {
	if len(ids) != len(newBalances) {
		return errors.New("UpdateBalanceOfAddresses: 不等长度的addr切片和balance切片")
	}
	for i, id := range ids {
		gsm.accounts[id].balance = newBalances[i]
	}
	return nil
}

/*检查账户是否存在*/
// 某个账户是否存在
func (gsm *globalStateMachine) HasUserID(id UserID) bool {
	_, ok := gsm.accounts[id]
	return ok
}

/*role相关*/
// 返回role信息(不能返回指针，因为不允许更改)
func (gsm *globalStateMachine) GetRoleOfUserID(id UserID) Role {
	return gsm.accounts[id].role
}

/*available相关*/
func (gsm *globalStateMachine) IsUserIDAvailable(id UserID) bool {
	return gsm.accounts[id].available
}

func (gsm *globalStateMachine) GetPubKeyOfUserID(id UserID) []byte {
	return gsm.accounts[id].pubKey
}
