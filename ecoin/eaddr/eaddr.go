package eaddr

import (
	"bytes"
	"encoding/gob"
	"github.com/azd1997/Ecare/ecoin/account"
	"github.com/azd1997/Ecare/ecoin/common"
	"github.com/azd1997/Ecare/ecoin/net"
	"github.com/azd1997/Ecare/ecoin/utils"
	"sync"
	"time"
)

// EAddr 节点地址，需要实现排序接口
// 每个EAddr都维护了一个自身的历史作恶记录表，这个表需要有时间顺序，而且一般也不太关心很久以前的记录，
// 但有的时候需要追溯倒数若干条记录， 用双链表比较合适。
// 由于可能在多goroutine中被修改，所以对它的操作还要加锁
type EAddr struct {
	Addr  Addr
	alias string

	userId account.UserId // 节点登录用的账户，需要在共识节点上线时向周遭发出，接收到的节点会检查其地址和账户状态，合格后向周遭扩散

	potMsg net.PotMsg // 当前出块周期该节点的PotMsg

	// Ping相关 每个周期都会定时触发Ping操作，继而更新Ping表
	pingDelay time.Duration  // 通信延迟。 0是不可达标志
	pingStart int64
	PingNum int
	PingLock *sync.RWMutex

	// 网络可达与否。 每次作为客户端 *主动发起连接* 时会检测连接失败与否
	// 一旦连接失败，视为不可达，除非对方主动发消息过来否则不会再向其发送数据
	// 这意味着，当对方节点主动发起连接时，需要先判断对方的可达状态并更新
	Reachable bool
	ReachableLock *sync.RWMutex

	// Honest诚实与否。 当信誉分降为0后，honest=false
	Honest    bool `json:"honest"` // 诚实与否的结果
	HonestLock *sync.RWMutex

	// 信誉分
	Credit           int // 信誉分. 假定信誉分初始为0， 每出一个区块加1，TODO
	CreditLock *sync.RWMutex

	// 作恶记录
	continuousBadNum int // 作恶记录链表的长度，节点作恶记录会持续记录，直至信誉分被扣光，节点被封禁。
	// 当节点发送某些特殊信息或者是赎金交易之后，恢复节点能力，此时，信誉分清零，重新开始记录作恶链表。但TotalBadNum会记录作恶总数
	totalBadNum int
	badRecords  *BadRecord // 这个链表一直从头部插入。（某种意义上是个栈）链头节点（也就是这个）就是最新的作恶记录
	BadRecordLock *sync.RWMutex

	sync.RWMutex
}


//==============================作恶记录===============================

func (eaddr *EAddr) TotalBadNum() int {
	return eaddr.totalBadNum
}

func (eaddr *EAddr) ContinuousBadNum() int {
	return eaddr.continuousBadNum
}

func (eaddr *EAddr) BadRecords() *BadRecord {
	return eaddr.badRecords
}

func (eaddr *EAddr) Record(behaviour uint8) {
	// 判断behaviour在不在creditPolicy表
	if v, ok := CreditPolicy[behaviour]; !ok { // v是对应的积分增量
		// do nothing
		return
	} else {

		// 合规记录
		if behaviour >= 100 {
			eaddr := eaddrs.EAddr(a)
			eaddr.Credit += v
			eaddrs.SetEAddr(&eaddr)
		} else {	// 作恶记录
			eaddr := eaddrs.EAddr(a)

			// 对副本进行操作
			eaddr.Credit += v
			// 检查Reachable
			if behaviour == BadConnFail {
				eaddr.Reachable = false
			}
			// 检查Honest
			if eaddr.Credit < 0 {
				eaddr.Honest = false
			}

			// 清除PotMsg和PingTime
			eaddr.PotMsg = net.PotMsg{}
			eaddr.PingTime = 0

			// 更新badnum
			badRecord := BadRecord{
				Time:    common.TimeStamp(time.Now().Unix()),
				BadType: behaviour,
				Punish:  v,
				Prev:    nil,
				Next:    eaddr.BadRecords,
			}
			eaddr.BadRecords.Prev = &badRecord
			eaddr.BadRecords = &badRecord
			eaddr.ContinuousBadNum++
			eaddr.TotalBadNum++


			eaddrs.SetEAddr(&eaddr)
		}
	}

}

func (eaddr *EAddr) PotMsg() net.PotMsg {
	return eaddr.potMsg
}

func (eaddr *EAddr) SetPotMsg(potMsg net.PotMsg) {
	eaddr.potMsg = potMsg
}

func (eaddr *EAddr) UserId() account.UserId {
	return eaddr.userId
}

func (eaddr *EAddr) SetUserId(userId account.UserId) {
	eaddr.userId = userId
}

func (eaddr *EAddr) Alias() string {
	return eaddr.alias
}

func (eaddr *EAddr) SetAlias(alias string) {
	eaddr.alias = alias
}

func (eaddr *EAddr) Serialize() ([]byte, error) {
	eaddr.RLock()
	defer eaddr.RUnlock()
	return utils.GobEncode(eaddr)
}

func (eaddr *EAddr) Deserialize(data []byte) error {
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(eaddr); err != nil {
		return err
	}
	return nil
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
		alias: alias,
	}
}

// BadType
const (
	BadUnknown  = iota
	BadConnFail // 注意连接失败一次扣固定分值。并视为节点掉线，不继续扣分，也不会再给它发消息

)

// Credit
var CreditPolicy = map[uint8]int{

	// 作恶情况
	BadUnknown:  -1,
	BadConnFail: -1,

	// 合规情况
	GoodUnknown: 1,
	GoodBlock:   1,
}

// GoodType
const (
	GoodUnknown = 100 + iota
	GoodBlock   // 出了一个好的区块
)
