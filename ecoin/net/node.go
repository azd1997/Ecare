package net

import (
	"bytes"
	"github.com/azd1997/Ecare/ecoin/account"
	"github.com/azd1997/Ecare/ecoin/blockchain/singlechain"
	eaccount "github.com/azd1997/Ecare/ecoin/ecoinaccount"
	"github.com/azd1997/Ecare/ecoin/utils"
	"github.com/azd1997/ego/enet/etcp"
	"io"
	"net"
)


type TCPNode struct {
	Server etcp.IServer

	Account account.Account

	Chain singlechain.Chain
	EAccounts eaccount.IEcoinAccounts
	EAddrs EAddrs
}

func NewTCPNode(args *Args) *TCPNode {
	return &TCPNode{
		Server:etcp.NewServer(
			etcp.DefaultOption().SetHost(args.Ip).SetPort(args.Port).SetName(args.Name).SetMaxPacketSize(1 << 24)),
																							// 2^23 = 8388608 大约是8M，够用了

		Account:args.Account,

		Chain:     args.Chain,
		EAccounts: args.EAccouts,
	}
}

// Start 启动节点TCP服务器，执行初始化流程，handle连接
func (n *TCPNode) Start(args *Args) {

	// 注册连接Hook函数
	n.Server.SetOnConnStart(args.OnConnStartHook)
	n.Server.SetOnConnStop(args.OnConnStopHook)

	// 注册多路由
	n.Server.AddRouters(args.Routers)

	// 启动server
	n.Server.Serve()

}

func checkNodeFirst(conn etcp.IConnection) {

}

// 客户端方法

// Push 将本地数据推送给远程节点。 key相当于cmd，value是目标数据， version用于索引。例如获取区块数据, key=block, value=blockdata, version=blocknum
func (n *TCPNode) Push(remote *Address, key string, value []byte, version []byte)  error {


	// 连接远程节点服务端
	conn, err := net.Dial(PROTOCOL, remote.Ipv4Port)
	if err != nil {
		return err
	}
	defer conn.Close()

	// 格式 op(push/pull) + key + versionLength + version + valuelength + value
	// (op占1字节 + key占8B，
	// versionLength 使用uint8，占1B， version为哈希（区块哈希、交易哈希）,
	// valuelength使用int，占4B/8B，具体使用时需要先判断下)
	data := bytes.Join([][]byte{
		[]byte{Push},	// 1B
		[]byte(key),	// 8B
		[]byte{uint8(len(version))},	// 1B
		version,		// len(Hash)
		utils.Uint32ToBytes(uint32(len(value))),
	}, nil)

	//将data []byte复制一份通过conn发给对方
	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		return utils.WrapError("Push", err)
	}
	return nil

}

// Pull 从远程拉取数据
func (n *TCPNode) Pull(remote Address, key string, version []byte) error {

	// 连接远程节点服务端
	conn, err := net.Dial(PROTOCOL, remote.Ipv4Port)
	if err != nil {
		return err
	}
	defer conn.Close()

	// 格式 op(push/pull) + key + versionLength + version + valuelength + value
	// (op占1字节 + key占8B，
	// versionLength 使用uint8，占1B， version为哈希（区块哈希、交易哈希）,
	// valuelength使用int，占4B/8B，具体使用时需要先判断下)
	data := bytes.Join([][]byte{
		[]byte{Pull},	// 1B
		[]byte(key),	// 8B
		[]byte{uint8(len(version))},	// 1B
		version,		// len(Hash)
	}, nil)

	//将data []byte复制一份通过conn发给对方
	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		return utils.WrapError("Pull", err)
	}
	return nil
}

// PullNPush 向远程拉取数据，并与本地比较，更新本地或者将本地的更新数据推给远程。 这种方式收敛最快。
// 具体实现由调用方组合Push&Pull，这里不作具体实现或函数传递
func (n *TCPNode) PullAndPush(key string, version []byte) {}


