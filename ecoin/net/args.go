package net

import (
	"github.com/azd1997/Ecare/ecoin/account"
	"github.com/azd1997/Ecare/ecoin/blockchain/singlechain"
	"github.com/azd1997/Ecare/ecoin/eaddr"
	eaccount "github.com/azd1997/Ecare/ecoin/ecoinaccount"
)

// Args 外部数据结构的参数，如区块链、数据存储...，注入到P2P节点中以方便调用
type Args struct {

	// Server参数
	Ip string
	Port int
	Name string


	Account account.Account

	Chain    singlechain.Chain
	EAccouts eaccount.IEcoinAccounts
	EAddrs   eaddr.EAddrs
}
