package net

import (
	"fmt"
)

type Addr struct {
	Ip string
	Port int
}

func (a Addr) String() string {
	return fmt.Sprintf("%s:%d", a.Ip, a.Port)
}

func (a Addr) IsValid(eaddrs EAddrs) bool {
	eaddrs.RLock()
	defer eaddrs.RUnlock()

	// Addr Ip和Port本身格式这里就不判断了，不对的话会连接失败，这里不管
	// 这里只负责不诚信和不可达两种情况的检查

	if !eaddrs.Map[a.String()].Honest || !eaddrs.Map[a.String()].Reachable {
		return false
	}

	return true
}

// 关于Addr的失信应该和账户状态一样也起一个集合存储并更新


type AddrList struct {
	List []Addr
}

