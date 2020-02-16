package common

import "github.com/azd1997/ego/ecrypto"

const NODE_VERSION = 1

type PingMsg struct {
	AddrFrom string
	Pong bool
}

type GetAddrsMsg struct {
	AddrFrom string
}

type AddrsMsg struct {
	AddrFrom string
	LocalAddrs []string
}

type VersionMsg struct {
	AddrFrom string
	NodeVersion uint8	// 节点版本
	MaxBlockID int		// 最大区块ID
	LatestBlockHash ecrypto.Hash	// “我的”最新区块的哈希
	SecondLatestBlockHash ecrypto.Hash	// 倒数第二个区块的哈希
	// 这个的意义在于校验latest区块的正确性
	// VersionMsg不仅用于自身链短想要同步，
	// 还用于自身链长同步给其他人
	// 正常运行中只考虑各个节点长度相差1，所以加上了这个SecondLatest...
}

const (
	InvBlock uint8 = iota
	InvTx
)

type GetInventoryMsg struct {
	InvType uint8	// 存证类型
	AddrFrom string
}

type InventoryMsg struct {
	InvType uint8	// 存证类型
	AddrFrom string
	Invs []ecrypto.Hash
}

