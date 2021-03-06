package p2p

import (
	"github.com/azd1997/Ecare/ecoin/account"
	"github.com/azd1997/Ecare/ecoin/blockchain/singlechain"
	eaccount "github.com/azd1997/Ecare/ecoin/ecoinaccount"
	"github.com/azd1997/ego/enet/etcp"
)

// P2P 点对点通信接口。这里实现基于TCP和HTTP rest两种协议的P2P通信
type Node interface {

}

// Args 外部数据结构的参数，如区块链、数据存储...，注入到P2P节点中以方便调用
type Args struct {

	// Server参数
	Ip string
	Port int
	Name string
	OnConnStartHook func(conn etcp.IConnection)
	OnConnStopHook func(conn etcp.IConnection)
	Routers map[uint32]etcp.IRouter

	Account account.Account

	Chain singlechain.Chain
	EAccouts eaccount.IEcoinAccounts
}
