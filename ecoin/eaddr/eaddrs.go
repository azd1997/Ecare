package eaddr

import (
	"github.com/azd1997/Ecare/ecoin/utils"
	"github.com/azd1997/ego/edatabase"
	"sync"
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
	eaddrs.RLock()
	defer eaddrs.RUnlock()
	eaddrs.m[addr.String()].PingStart()
}

func (eaddrs *EAddrs) EAddrPingStop(addr Addr) {
	eaddrs.RLock()
	defer eaddrs.RUnlock()
	eaddrs.m[addr.String()].PingStop()
}

func (eaddrs *EAddrs) Record(a Addr, behaviour uint8) {
	eaddrs.RLock()
	defer eaddrs.RUnlock()
	eaddrs.m[a.String()].Record(behaviour)
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
