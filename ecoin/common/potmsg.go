package common

import "github.com/azd1997/ego/ecrypto"

// Pot 交易量证明，所有节点在接收到最新区块后刷新本地period计时，period/2计时一到广播本机Pot，所有人得到一份Pot集合后决出POT胜者.
// 再period/2后
type PotMsg struct {
	AddrFrom string
	Pot *Pot	// 用来检查区块内所含交易是否与竞争POT时一致，不含coinbase和仲裁交易
}

func (msg *PotMsg) LargerThan(another *PotMsg) bool {
	return msg.Pot.LargerThan(another.Pot, msg.AddrFrom, another.AddrFrom)
}

// Pot 记录
type Pot struct {
	num int
	hashsum ecrypto.Hash
}

func (pot *Pot) LargerThan(another *Pot, selfAddr, anotherAddr string) bool {
	if pot.num > another.num {
		return true
	}
	if pot.hashsum.LargerThan(another.hashsum, selfAddr, anotherAddr) {
		return true
	}
	return false
}