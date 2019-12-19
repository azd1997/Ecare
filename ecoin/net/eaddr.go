package net

import (
	"github.com/azd1997/Ecare/ecoin/account"
	"github.com/azd1997/Ecare/ecoin/common"
	"sync"
	"time"
)

// EAddr 节点地址，需要实现排序接口
// 每个EAddr都维护了一个自身的历史作恶记录表，这个表需要有时间顺序，而且一般也不太关心很久以前的记录，
// 但有的时候需要追溯倒数若干条记录， 用双链表比较合适。
// 由于可能在多goroutine中被修改，所以对它的操作还要加锁
type EAddr struct {
	Addr Addr
	Alias            string        `json:"alias"`

	UserId account.UserId	// 节点登录用的账户，需要在共识节点上线时向周遭发出，接收到的节点会检查其地址和账户状态，合格后向周遭扩散

	PotMsg PotMsg	// 当前出块周期该节点的PotMsg

	PingTime         time.Duration `json:"pingTime"` // 通信延迟。 0是不可达标志

	Honest           bool          `json:"honest"`   // 诚实与否的结果
	Reachable 		bool		// 网络可达与否。一旦连接失败，视为不可达，除非对方主动发消息过来否则不会再向其发送数据
								// 这意味着，当对方节点主动发起连接时，需要先判断对方的可达状态并更新
	Credit           int           // 信誉分. 假定信誉分初始为0， 每出一个区块加1，TODO
	ContinuousBadNum int           // 作恶记录链表的长度，节点作恶记录会持续记录，直至信誉分被扣光，节点被封禁。
	// 当节点发送某些特殊信息或者是赎金交易之后，恢复节点能力，此时，信誉分清零，重新开始记录作恶链表。但TotalBadNum会记录作恶总数
	TotalBadNum int
	BadRecords  *BadRecord // 这个链表一直从头部插入。（某种意义上是个栈）链头节点（也就是这个）就是最新的作恶记录

	sync.RWMutex // 保护EAddr的修改
}

// 作恶记录，双链表节点
type BadRecord struct {
	Time       common.TimeStamp
	BadType    uint8 // 作恶类型
	Punish     int   // 惩罚，负值
	Prev, Next *BadRecord
}

// NewEAddr 新建一个Address。ping为0, 表示未知， honest为true
func NewAddr(addr Addr, alias string) *EAddr {
	return &EAddr{
		Addr:  addr,
		Alias: alias,
	}
}

// BadType
const (
	BadUnknown = iota
	BadConnFail				// 注意连接失败一次扣固定分值。并视为节点掉线，不继续扣分，也不会再给它发消息

)

// Credit
var CreditPolicy = map[uint8]int{

	// 作恶情况
	BadUnknown: -1,
	BadConnFail: -1,

	// 合规情况
	GoodUnknown: 1,
	GoodBlock: 1,
}

// GoodType
const (
	GoodUnknown = 100 + iota
	GoodBlock	// 出了一个好的区块
)