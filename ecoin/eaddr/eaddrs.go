package eaddr

import (
	"github.com/azd1997/Ecare/ecoin/common"
	"github.com/azd1997/Ecare/ecoin/net"
	"github.com/azd1997/Ecare/ecoin/utils"
	"github.com/azd1997/ego/edatabase"
	"sync"
	"time"
)

// 每个节点存储的节点集合。普通用户不必存这么详细，有待删减
// 包含自身
type EAddrs struct {
	m map[string]*EAddr
	sync.RWMutex	// 修改map中键值对时不需要外边这把写锁（但需要读），但是增加键或删除键需要上写锁

	// 存储相关
	DbEngine string
	DbPath string
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

func (eaddrs *EAddrs) EAddrPingStart(addr Addr) {
	eaddrs.Lock()
	defer eaddrs.Unlock()

	eaddrs.m[addr.String()].PingStart = time.Now().UnixNano()
}

func (eaddrs *EAddrs) EAddrPingStop(addr Addr) {
	eaddrs.Lock()
	defer eaddrs.Unlock()

	eaddrs.m[addr.String()].PingDelay = time.Since(time.Unix(0, eaddrs.m[addr.String()].PingStart))
}

func (eaddrs *EAddrs) Record(a Addr, behaviour uint8) {
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

func (eaddrs *EAddrs) Save() error {
	// 保存到数据库

	// 拷贝一份，避免冲突
	eaddrs.RLock()
	kList := make([][]byte, len(eaddrs.m))
	vList := make([][]byte, len(eaddrs.m))
	i := 0
	for k, v := range eaddrs.m {
		vBytes, err := v.Serialize()
		if err != nil {continue}
		kList[i], vList[i] = []byte(k), vBytes
		i++
	}
	eaddrs.RUnlock()

	// 打开数据库并保存
	db, err := edatabase.OpenDatabaseWithRetry(eaddrs.DbEngine, eaddrs.DbPath)
	if err != nil {
		return utils.WrapError("EAddrs_Save", err)
	}
	if err := db.BatchSet(kList, vList); err != nil {
		return utils.WrapError("EAddrs_Save", err)
	}

	return nil
}

func (eaddrs *EAddrs) Load() error {
	// 从数据库加载出来

	// eaddrs的m是空的，只有mutex
	
	
	if !edatabase.DbExists(eaddrs.DbEngine, eaddrs.DbPath) {
		return utils.WrapError("EAddrs_Load", ErrDbNotExists)
	}
	db, err := edatabase.OpenDatabaseWithRetry(eaddrs.DbEngine, eaddrs.DbPath)
	if err != nil {
		return utils.WrapError("EAddrs_Load", err)
	}
	db.IterDB(func(k, v []byte) error {
		eaddr := &EAddr{}
		if err = eaddr.Deserialize(v); err != nil {
			return err
		}
		eaddrs.m[string(k)] = eaddr
		return nil
	})

	return nil
}

// TODO: 其他各种方法
