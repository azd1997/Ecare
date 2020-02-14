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

	// 当前出块周期该节点的PotMsg
	// 其读取和更新一定是不同时的，所以不会产生竞态
	potMsg net.PotMsg

	// Ping相关 每个周期都会定时触发Ping操作，继而更新Ping表
	// Ping的更新是固定时间间隔的，但是查取是用在EAddrs取排序节点时使用，所以读写时间不确定顺序，需要加个锁（不加影响也不大，其实这里所有成员都不加影响也不大）
	pingDelay time.Duration  // 通信延迟。 0是不可达标志
	pingStartTime int64
	pingNum int
	pingLock sync.RWMutex

	// 网络可达与否。 每次作为客户端 *主动发起连接* 时会检测连接失败与否
	// 一旦连接失败，视为不可达，除非对方主动发消息过来否则不会再向其发送数据
	// 这意味着，当对方节点主动发起连接时，需要先判断对方的可达状态并更新
	reachable bool
	reachableLock sync.RWMutex

	// Honest诚实与否。 当信誉分降为0后，honest=false
	honest    bool  // 诚实与否的结果
	honestLock sync.RWMutex

	// 信誉分
	credit           int // 信誉分. 假定信誉分初始为0， 每出一个区块加1，TODO
	creditLock sync.RWMutex

	// 作恶记录
	continuousBadNum int // 作恶记录链表的长度，节点作恶记录会持续记录，直至信誉分被扣光，节点被封禁。
	// 当节点发送某些特殊信息或者是赎金交易之后，恢复节点能力，此时，信誉分清零，重新开始记录作恶链表。但TotalBadNum会记录作恶总数
	totalBadNum int
	badRecords  *BadRecord // 这个链表一直从头部插入。（某种意义上是个栈）链头节点（也就是这个）就是最新的作恶记录
	badRecordLock sync.RWMutex

}


// 在EAddr中主要记录了两大类信息：网络通信延迟、节点诚信相关。
// 这两大信息部分需要自身持久化并在重上线后加载
// 而节点诚信信息需要全网保持同步。需要在节点上线时从其他节点同步
// 和EAccounts一样，有关诚信等相关的信息需要请求同步


//=============================节点别名================================

func (eaddr *EAddr) Alias() string {
	return eaddr.alias
}

func (eaddr *EAddr) SetAlias(alias string) {
	eaddr.alias = alias
}

//=============================地址绑定的账户标识================================

func (eaddr *EAddr) UserId() account.UserId {
	return eaddr.userId
}

func (eaddr *EAddr) SetUserId(userId account.UserId) {
	eaddr.userId = userId
}

//=============================PoT Msg================================

func (eaddr *EAddr) PotMsg() net.PotMsg {
	return eaddr.potMsg
}

func (eaddr *EAddr) SetPotMsg(potMsg net.PotMsg) {
	eaddr.potMsg = potMsg
}

//=============================Ping相关================================

func (eaddr *EAddr) PingNum() int {
	eaddr.pingLock.RLock()
	defer eaddr.pingLock.RUnlock()
	return eaddr.pingNum
}

func (eaddr *EAddr) PingStartTime() int64 {
	eaddr.pingLock.RLock()
	defer eaddr.pingLock.RUnlock()
	return eaddr.pingStartTime
}

func (eaddr *EAddr) PingDelay() time.Duration {
	eaddr.pingLock.RLock()
	defer eaddr.pingLock.RUnlock()
	return eaddr.pingDelay
}

func (eaddr *EAddr) clearPingDelay() {
	eaddr.pingLock.Lock()
	defer eaddr.pingLock.Unlock()
	eaddr.pingDelay = 0
}

func (eaddr *EAddr) PingStart() {
	eaddr.pingLock.Lock()
	defer eaddr.pingLock.Unlock()

	eaddr.pingNum++
	eaddr.pingStartTime = time.Now().UnixNano()
}

func (eaddr *EAddr) PingStop() {
	eaddr.pingLock.Lock()
	defer eaddr.pingLock.Unlock()

	eaddr.pingDelay = time.Since(time.Unix(0, eaddr.pingStartTime))
}

//=============================网络可达================================

func (eaddr *EAddr) Reachable() bool {
	eaddr.reachableLock.RLock()
	defer eaddr.reachableLock.RUnlock()
	return eaddr.reachable
}

func (eaddr *EAddr) setReachable(reach bool) {
	eaddr.reachableLock.Lock()
	defer eaddr.reachableLock.Unlock()
	eaddr.reachable = reach
}

//=============================节点诚实================================

func (eaddr *EAddr) Honest() bool {
	eaddr.honestLock.RLock()
	defer eaddr.honestLock.RUnlock()
	return eaddr.honest
}

func (eaddr *EAddr) setHonest(honest bool) {
	eaddr.honestLock.Lock()
	defer eaddr.honestLock.Unlock()
	eaddr.honest = honest
}

//=============================信誉积分================================

func (eaddr *EAddr) Credit() int {
	eaddr.creditLock.RUnlock()
	defer eaddr.creditLock.RUnlock()
	return eaddr.credit
}

func (eaddr *EAddr) addCredit(delta int) {
	eaddr.creditLock.Unlock()
	defer eaddr.creditLock.Unlock()
	eaddr.credit += delta
}

//==============================作恶记录===============================

func (eaddr *EAddr) TotalBadNum() int {
	eaddr.badRecordLock.RLock()
	defer eaddr.badRecordLock.RUnlock()
	return eaddr.totalBadNum
}

func (eaddr *EAddr) ContinuousBadNum() int {
	eaddr.badRecordLock.RLock()
	defer eaddr.badRecordLock.RUnlock()
	return eaddr.continuousBadNum
}

func (eaddr *EAddr) BadRecords() *BadRecord {
	// 注意这里其实是写锁
	eaddr.badRecordLock.Lock()
	defer eaddr.badRecordLock.Unlock()
	return eaddr.badRecords
}

func (eaddr *EAddr) insetBadRecords(r *BadRecord) {
	// r.Next已经指向原头部节点
	eaddr.badRecordLock.Lock()
	defer eaddr.badRecordLock.Unlock()
	eaddr.badRecords.Prev = r
	eaddr.badRecords = r
}

func (eaddr *EAddr) Record(behaviour uint8) {
	// 判断behaviour在不在creditPolicy表
	if v, ok := CreditPolicy[behaviour]; !ok { // v是对应的积分增量
		// do nothing
		return
	} else {

		// 合规记录
		if behaviour >= 100 {
			eaddr.addCredit(v)
		} else {	// 作恶记录

			eaddr.addCredit(v)
			// 检查Reachable
			if behaviour == BadConnFail {
				eaddr.setReachable(false)
			}
			// 检查Honest
			if eaddr.Credit() < 0 {
				eaddr.setHonest(false)
			}

			// 清除PingTime
			eaddr.clearPingDelay()

			// 更新badnum
			badRecord := BadRecord{
				Time:    common.TimeStamp(time.Now().Unix()),
				BadType: behaviour,
				Punish:  v,
				Prev:    nil,
				Next:    eaddr.BadRecords(),
			}
			eaddr.BadRecords().Prev = &badRecord
			eaddr.insetBadRecords(&badRecord)
			eaddr.continuousBadNum++
			eaddr.totalBadNum++

		}
	}

}

//==============================序列化===============================

type eAddr struct {
	Addr  Addr
	Alias string
	UserId account.UserId

	PingDelay time.Duration
	PingNum int

	Reachable bool

	Honest    bool

	Credit           int

	ContinuousBadNum int
	TotalBadNum int
	BadRecords  *BadRecord
}

func (eaddr *EAddr) Serialize() ([]byte, error) {

	eaddr.pingLock.RLock()
	eaddr.reachableLock.RLock()
	eaddr.honestLock.RLock()
	eaddr.creditLock.RLock()
	eaddr.badRecordLock.RLock()

	eaddrC := eAddr{
		Addr:             eaddr.Addr,
		Alias:            eaddr.alias,
		UserId:           eaddr.userId,
		PingDelay:        eaddr.pingDelay,
		PingNum:          eaddr.pingNum,
		Reachable:        eaddr.reachable,
		Honest:           eaddr.honest,
		Credit:           eaddr.credit,
		ContinuousBadNum: eaddr.continuousBadNum,
		TotalBadNum:      eaddr.totalBadNum,
		BadRecords:       eaddr.badRecords,
	}

	eaddr.pingLock.RUnlock()
	eaddr.reachableLock.RUnlock()
	eaddr.honestLock.RUnlock()
	eaddr.creditLock.RUnlock()
	eaddr.badRecordLock.RUnlock()

	return utils.GobEncode(eaddrC)
}

func (eaddr *EAddr) Deserialize(data []byte) error {

	// eaddr需要是已经初始化好各个锁的

	eaddrC := &eAddr{}
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(eaddrC); err != nil {
		return err
	}

	eaddr.Addr = eaddrC.Addr
	eaddr.alias = eaddrC.Alias
	eaddr.pingNum = eaddrC.PingNum
	eaddr.pingDelay = eaddrC.PingDelay
	eaddr.reachable = eaddrC.Reachable
	eaddr.honest = eaddrC.Honest
	eaddr.credit = eaddrC.Credit
	eaddr.continuousBadNum = eaddrC.ContinuousBadNum
	eaddr.totalBadNum = eaddrC.TotalBadNum
	eaddr.badRecords = eaddrC.BadRecords

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
func NewEAddr(addr Addr, alias string) *EAddr {
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
