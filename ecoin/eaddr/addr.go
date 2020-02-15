package eaddr

import (
	"fmt"
	"net"
)

type Addr struct {
	Ip string
	Port int
}

// String 输出为ipv4:port字符串
func (a Addr) String() string {
	return fmt.Sprintf("%s:%d", a.Ip, a.Port)
}

// IsValid 判断Addr是否有效
func (a Addr) IsValid(eaddrs *EAddrs) bool {
	eaddrs.RLock()
	defer eaddrs.RUnlock()

	// Addr Ip和Port本身格式这里就不判断了，不对的话会连接失败，这里不管
	// 这里只负责不诚信和不可达两种情况的检查

	if !eaddrs.IsAddrValid(a) {
		return false
	}

	return true
}

// NewAddr 将地址字符串转为Addr结构体
func NewAddr(addr string) Addr {
	tcpaddr, err := net.ResolveTCPAddr("tcp4", addr)
	if err != nil {return Addr{}}
	return Addr{tcpaddr.IP.String(), tcpaddr.Port}
}

// 关于Addr的失信应该和账户状态一样也起一个集合存储并更新


//type AddrList struct {
//	List []Addr
//}

