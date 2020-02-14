package eaddr

import (
	"github.com/azd1997/Ecare/ecoin/erro"
	"github.com/azd1997/Ecare/ecoin/utils"
	"github.com/azd1997/ego/edatabase"
	"sort"
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

// ValidAddrs 返回所有有效的地址列表
func (eaddrs *EAddrs) ValidAddrs() []string {
	eaddrs.RWMutex.RLock()
	defer eaddrs.RWMutex.RUnlock()

	res := make([]string, 0, len(eaddrs.m))
	for addrStr, eaddr := range eaddrs.m {
		if eaddr.IsValid() {
			res = append(res, addrStr)
		}
	}

	return res
}

// SortedValidAddrs 返回所有有效的地址列表，并且排好序
func (eaddrs *EAddrs) SortedValidAddrs() []string {
	eaddrs.RWMutex.RLock()
	defer eaddrs.RWMutex.RUnlock()

	res := make([]string, 0, len(eaddrs.m))
	for addrStr, eaddr := range eaddrs.m {
		if eaddr.IsValid() {
			res = append(res, addrStr)
		}
	}
	// 按pingtime排序
	sort.Slice(res, func(i, j int) bool {
		return eaddrs.m[res[i]].pingDelay < eaddrs.m[res[j]].pingDelay
	})
	return res
}

// EAddr 根据Addr查看EAddr
func (eaddrs *EAddrs) EAddr(addr Addr) *EAddr {
	eaddrs.RLock()
	defer eaddrs.RUnlock()
	return eaddrs.m[addr.String()]
}

// SetEAddr 设置/更新EAddr
// 修改EAddr时，建议直接读出来拷贝一份再传回
func (eaddrs *EAddrs) SetEAddr(eaddr *EAddr) {
	eaddrs.Lock()
	defer eaddrs.Unlock()
	eaddrs.m[eaddr.Addr.String()] = eaddr
}

// EAddrPingStart 开始ping一个Addr
func (eaddrs *EAddrs) EAddrPingStart(addr Addr) {
	eaddrs.RLock()
	defer eaddrs.RUnlock()
	eaddrs.m[addr.String()].PingStart()
}

// EAddrPingEnd 结束ping一个Addr，将时延记录
func (eaddrs *EAddrs) EAddrPingStop(addr Addr) {
	eaddrs.RLock()
	defer eaddrs.RUnlock()
	eaddrs.m[addr.String()].PingStop()
}

// Record 记录某个Addr的某个行为
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
		return utils.WrapError("EAddrs_Load", erro.ErrDbNotExists)
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

// IsAddrValid 查看Addr是否有效
func (eaddrs *EAddrs) IsAddrValid(addr Addr) bool {
	eaddrs.RWMutex.RLock()
	defer eaddrs.RWMutex.RUnlock()
	eaddr := eaddrs.m[addr.String()]
	return eaddr.honest && eaddr.reachable
}