package net

import (
	"github.com/azd1997/Ecare/ecoin/common"
	"sync"
	"time"
)

// 每个节点存储的节点集合。普通用户不必存这么详细，有待删减
// 包含自身
type EAddrs struct {
	m map[string]*EAddr
	sync.RWMutex	// 修改map中键值对时不需要外边这把写锁（但需要读），但是增加键或删除键需要上写锁
}

func (eaddrs *EAddrs) ValidAddrs() []Addr {
	return nil
}

func (eaddrs *EAddrs) SortedValidAddrs() []Addr {
	return nil
}

func (eaddrs *EAddrs) EAddr(addr Addr) EAddr {
	eaddrs.RLock()
	defer eaddrs.RUnlock()
	return *eaddrs.m[addr.String()]
}

func (eaddrs *EAddrs) SetEAddr(eaddr *EAddr) {
	eaddrs.Lock()
	defer eaddrs.Unlock()
	eaddrs.m[eaddr.Addr.String()] = eaddr
}

func (eaddrs *EAddrs) Record(a Addr, behaviour uint8) {
	// 判断behaviour在不在creditPolicy表
	if v, ok := CreditPolicy[behaviour]; !ok {	// v是对应的积分增量
		// do nothing
		return
	} else {

		aString := a.String()

		// 合规记录
		if behaviour >= 100 {
			eaddrs.RLock()
			eaddrs.m[aString].Lock()
			eaddrs.m[aString].Credit += v
			eaddrs.m[aString].Unlock()
			eaddrs.RUnlock()
		} else {	// 作恶记录
			eaddrs.RLock()
			eaddr := *eaddrs.m[aString]	// 注意把结构体拷贝出来而不是指针
			eaddrs.RUnlock()

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
			eaddr.PotMsg = PotMsg{}
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


			eaddrs.Lock()
			eaddrs.m[aString] = &eaddr
			eaddrs.Unlock()
		}
	}

}

// TODO: 其他各种方法
