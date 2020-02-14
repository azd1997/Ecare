package net

import (
	"bytes"
	"github.com/azd1997/Ecare/ecoin/account"
	"github.com/azd1997/Ecare/ecoin/blockchain/singlechain"
	"github.com/azd1997/Ecare/ecoin/eaddr"
	eaccount "github.com/azd1997/Ecare/ecoin/ecoinaccount"
	"github.com/azd1997/Ecare/ecoin/log"
	"github.com/azd1997/Ecare/ecoin/utils"
	"github.com/azd1997/ego/enet/etcp"
	"io"
	"io/ioutil"
	"net"
)


type TCPNode struct {

	// TCPServer相关配置
	Listener net.Listener
	Addr     eaddr.Addr
	Name     string


	// 其他
	Account   account.Account
	Chain     singlechain.Chain
	EAccounts eaccount.IEcoinAccounts
	EAddrs    eaddr.EAddrs
}

func NewTCPNode(args *Args) *TCPNode {
	return &TCPNode{
		Addr: eaddr.Addr{
			Ip:   args.Ip,
			Port: args.Port,
		},
		Name:args.Name,

		Account:args.Account,

		Chain:     args.Chain,
		EAccounts: args.EAccouts,
		EAddrs:args.EAddrs,
	}
}

// Start 启动节点TCP服务器，执行初始化流程，handle连接
// 在调用TCPNode.Start之前，将其他全局变量准备好再启动
// 注意：这里是即作为服务端启动，也作为客户端。作为客户端时主动向其他节点拉取数据
func (n *TCPNode) Start() {

	// 开始监听
	listener, err := net.Listen(PROTOCOL, n.Addr.String())
	if err != nil {
		log.Error("StartServer", err)
	}
	defer listener.Close()
	n.Listener = listener

	// 节点启动之初，需要保证自身的节点集合更新到最新、区块链更新到最新、账户集合更新到最新。不更新到最新

	// 首先加载本地节点集合看是否存在，存在则向列表所有集合进行Pull操作拉取最新信息。
	// 若不存在，则向配置文件传入的的一个seed地址Pull数据，都没有，则退出，启动失败
	// 这些操作在调用本方法之前准备好


	// 循环监听请求并尝试接受连接
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go n.HandleConnection(conn)
	}
}

func (n *TCPNode) HandleConnection(conn net.Conn) {
	defer conn.Close()

	// 读取请求
	request, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Error("TCPNode_HandleConn", err)
	}

	// 获取command
	cmd := uint8(request[0])
	log.Info("Received %s command\n", CmdMap[cmd])

	// 根据命令交由不同handler去处理
	switch cmd {
	case CmdPing:
		n.HandlePing(conn)
	case CmdPong:
		n.HandlePong()
	case CmdAddrs:
		n.HandleAddrs(conn, request[1:])
	case CmdInv:
		n.HandleInv(conn, request[1:])
	case CmdVersion:
		n.HandleVersion(conn, request[1:])
	case CmdBlock:
		n.HandleBlock(conn, request[1:])
	}
}

//====================================Ping-Pong=======================================

func (n *TCPNode) Ping(to eaddr.Addr) {
	req := []byte{CmdPing}
	conn, _ := net.Dial(PROTOCOL, to.String())
	_, _ = conn.Write(req)
}

func (n *TCPNode) HandlePing(conn net.Conn) {
	response := []byte{CmdPong}
	_, _ = conn.Write(response)
}

func (n *TCPNode) HandlePong(conn net.Conn) {
	// 获取远程主机地址。 （事实上，在Ping-Pong过程中返回Pong的一定是你主动发起连接的目标主机。
	// 可以通过让返回的Pong包含这个地址信息，也可以像这样获取出来，都行）
	tcpaddr, _ := net.ResolveTCPAddr(conn.RemoteAddr().Network(), conn.RemoteAddr().String())
	addr := eaddr.Addr{Ip: tcpaddr.IP.String(), Port: tcpaddr.Port}

	// Ping stop， 处理总耗时
	n.EAddrs.EAddrPingStop(addr)
}


//====================================Addrs=======================================

// Addrs的场景
// 1. 节点上线时向seed节点(可能是写死的或者自选的)请求节点列表，seed节点随之将其节点列表打包发回
// 2. 节点注册时，藉由注册命令，seed节点会将该节点地址向其他节点扩散。

func (n *TCPNode) SendGetAddrs(to eaddr.Addr) {
	req := []byte{CmdGetAddrs}
	conn, _ := net.Dial(PROTOCOL, to.String())
	_, _ = conn.Write(req)
}

func (n *TCPNode) HandleGetAddrs(conn net.Conn) {
	addrs := n.EAddrs.ValidAddrs()
	payload, _ := utils.GobEncode(addrs)
	resp := append([]byte{CmdAddrs}, payload...)
	_, _ = conn.Write(resp)
}

func (n *TCPNode) HandleAddrs(conn net.Conn) {
	addrs := n.EAddrs.ValidAddrs()
	payload, _ := utils.GobEncode(addrs)
	resp := append([]byte{CmdGetAddrs}, payload...)
	_, _ = conn.Write(resp)
}

//====================================EAddrs=======================================

//====================================Version=======================================

type Version struct {
	NodeVer uint8
	ChainHeight int

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


