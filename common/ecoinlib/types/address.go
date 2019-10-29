package types

import (
	"github.com/azd1997/Ecare/common/ecoinlib/utils"
	"io/ioutil"
	"sort"
	"time"
)

// Address 节点地址，需要实现排序接口
type Address struct {
	Ipv4Port string
	Alias string
	PingTime time.Duration	// 通信延迟
	Honest bool
}

// NodeAddr 获取节点地址
func (a *Address) NodeAddr() string {
	return a.Ipv4Port
}

// AddressList 节点地址列表，需要事先排序列表
type AddressList struct {
	list []*Address
	less func(x, y *Address) bool
}

// AddressListLessFn 多层比较，先比较是否是诚实的，再比较通信延迟。使得诚实且通信延迟低的在前。使用时传入AddrList{[]*Address{}, AddressListLessFn}
// 排序时使用
var AddressListLessFn = func(x, y *Address) bool {
	if x.Honest != y.Honest {	// 如果x,y一个是诚实的一个是不诚实的，那么谁不诚实谁小
		return x.Honest == false
	}
	if x.PingTime != y.PingTime {
		return x.PingTime > y.PingTime
	}

	return false
}

// Less 返回下标为i的元素是否比下标为j的元素“更小”
func (al AddressList) Less(i, j int) bool {
	return al.less(al.list[i], al.list[j])		// 这样做，可以更换这个结构体的less方法，实现多层比较
}

// Len 返回切片长度
func (al AddressList) Len() int {
	return len(al.list)
}

// Swap 交换元素位置
func (al AddressList) Swap(i, j int) {
	al.list[i], al.list[j] = al.list[j], al.list[i]
}

// func (addr Address)

// address类
type AddrLists struct {
	L1 []*Address // 可用转发节点集合，都维护
	L2 []*Address // 转发节点维护的连接叶节点集合
	L3 []*Address // 叶节点维护的朋友节点集合
}

// SaveFile 保存到本地
func (a *AddrLists) SaveFile(filePath string) (err error) {
	res, err := utils.GobEncode(a)
	if err != nil {
		return utils.WrapError("AddrList_SaveFile", err)
	}

	if err = ioutil.WriteFile(filePath, res, 0644); err != nil {
		return utils.WrapError("AddrList_SaveFile", err)
	}
	return nil
}

// LoadFile 从本地还原出AddrList。 addrs := &AddrList{}
func (z *AddrLists) LoadFile(filePath string) (err error) {
	z1 := &AddrLists{}
	if z != z1 {
		return ErrLoadFileNeedEmptyReceiver
	}
	// TODO

	return nil
}

// Sort 原地排序
func (a *AddrLists) Sort() {
	if a.L1 != nil && len(a.L1) > 1 {
		sort.Sort(AddressList{list:a.L1, less:AddressListLessFn})
	}
	if a.L2 != nil && len(a.L2) > 1 {
		sort.Sort(AddressList{list:a.L2, less:AddressListLessFn})
	}
	if a.L3 != nil && len(a.L3) > 1 {
		sort.Sort(AddressList{list:a.L3, less:AddressListLessFn})
	}
	// 否则啥也不干
}

// L1Strings 获取L1的ipv4字符串
func (a *AddrLists) L1Ipv4Honest() []string {

	var res []string

	for _, addr := range a.L1 {
		if addr.Honest {
			res = append(res, addr.NodeAddr())
		}
	}

	return res
}

// multi-connection concurrency communication tree 多连接并发通信树
// 这里参考了李皎《区块链数据通信性能优化》一书中提出的考虑节点失效的多连接并发通信树的一些内容，并作一些基于该场景的适应性修改

// 在我的初始设计中，存在四类角色：医院、第三方机构、医生、病人，当然还有一个创建者
// 显然医院和第三方是基本不会下线的稳定节点，为了通信效率：
// 医院、第三方为拥有打包区块权限（这里按照习惯也称作挖矿）
// 在通信树中，出块节点作为通信树源节点，其余医院及第三方作为转发节点，病人和医生的电脑或者手机默认为叶节点
// 无论是哪一类节点新上线时必定向医院和第三方发起同步请求

// 怎么决定谁挖矿？医院和第三方节点中都会不断收集最新得到的交易，并打包区块
// 采用激励措施？	暂时不想
// 随机从健康节点选择？

// 这棵树不是显性的，对于每个节点来说，只需要知道它需要连接的其他节点，并不需要知道整棵树的构造
// 那么在我的设计里，病人节点和医生节点成为叶节点，只需要知道所有医院和第三方节点地址，并进行ping通检测，得到新的可用节点集合，且根据ping通响应速度来排序，优先从最快相应的节点同步，紧接着向第二个节点同步，直至第三个（这个目的是为了防止第一个节点不是最新的）
// 那么对于医院和第三方，他需要维护两个集合，一个是它用于同步的其他转发节点集合，一个是他需要提供同步的（可以省略）
// 等于说所有节点都只需要维护医院及第三方的结点地址集合及可用集合。所以M3cTree结构如图

// 同步策略：
// 医院和第三方之间p2p【多连接并发】同步
// 叶节点上线后向转发节点请求同步，同步后驻留在转发节点维护的叶节点集合中，转发节点在得到新区快后会向叶节点集合发送新区快并重试三次，三次后失败者剔除出本地集合

// M3cTree
type M3cTree struct {
}
