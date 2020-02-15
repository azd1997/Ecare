package net

import (
	"bytes"
	"encoding/gob"
	"github.com/azd1997/Ecare/ecoin/account"
	"github.com/azd1997/Ecare/ecoin/blockchain/singlechain"
	"github.com/azd1997/Ecare/ecoin/common"
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

	// NodeVersion
	Version uint8

	// TCPServer相关配置
	Listener net.Listener
	Addr     eaddr.Addr
	Name     string


	// 其他
	Account   *account.Account
	Chain     *singlechain.Chain
	EAccounts eaccount.IEcoinAccounts
	EAddrs    *eaddr.EAddrs
}

func NewTCPNode(args *Args) *TCPNode {
	return &TCPNode{
		Version:args.NodeVersion,

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

	// 把req全部读出读取出来(各种消息)
	req, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Error("TCPNode_HandleConn", err)
	}

	log.Info("Received %s command\n", CmdMap[req[0]])

	// 根据命令交由不同handler去处理
	switch req[0] {
	case CmdPing:
		n.HandlePing(req[1:])
	case CmdPong:
		n.HandlePong(req[1:])
	case CmdGetAddrs:
		n.HandleGetAddrs(req[1:])
	case CmdAddrs:
		n.HandleAddrs(req[1:])
	case CmdVersion:
		n.HandleVersion(req[1:])
	case CmdInv:
		n.HandleInventory(req[1:])
	case CmdBlock:
		n.HandleBlock(conn, request[1:])
	}
}

//====================================Ping-Pong=======================================

// TODO: 还有个问题待解决：何时Ping，如何触发

func (n *TCPNode) Ping(to string) {
	// 开始Ping
	n.EAddrs.EAddrPingStartStr(to)

	pingmsg := common.PingMsg{AddrFrom: n.Addr.String()}
	payload, _ := utils.GobEncode(pingmsg)
	req := append([]byte{CmdPing}, payload...)
	_ = n.SendData(to, req)
}

func (n *TCPNode) HandlePing(req []byte) {
	// 解析req
	pingmsg := &common.PingMsg{}
	_ = gob.NewDecoder(bytes.NewReader(req)).Decode(pingmsg)
	pingfrom := pingmsg.AddrFrom
	// 返回resp
	pongmsg := &common.PongMsg{AddrFrom:n.Addr.String()}
	payload, _ := utils.GobEncode(pongmsg)
	resp := append([]byte{CmdPong}, payload...)
	_ = n.SendData(pingfrom, resp)
}

func (n *TCPNode) HandlePong(req []byte) {
	// 解析req
	pongmsg := &common.PongMsg{}
	_ = gob.NewDecoder(bytes.NewReader(req)).Decode(pongmsg)
	pongfrom := pongmsg.AddrFrom
	// 处理pong
	n.EAddrs.EAddrPingStopStr(pongfrom)
}


//====================================Addrs=======================================

// Addrs的场景
// 1. 节点上线时向seed节点(可能是写死的或者自选的)请求节点列表，seed节点随之将其节点列表打包发回
// 2. 节点注册时，藉由注册命令，seed节点会将该节点地址向其他节点扩散。
// TODO: 尽管注册行为使用msg更方便，但使用交易实现会更体现行为历史可追溯的特性
// TODO: 注册流程：本地新建账号，本地监听端口(仅H/R需要)，将账号信息、节点地址信息、身份信息
//  使用注册医院的公钥进行加密（患者可加密，其他角色需要公示），构建交易，发往工人组，此称为注册交易。
//  工人节点接收到注册交易时不作其他检查，只检查是否来自该地址IP是否过于频繁的注册
//  （注册交易也用于更新账户信息）。
//  当注册医院收到这个链上确认的注册交易后，如果是患者的注册，它需要解密患者身份信息，记录下来


// 作为客户端发起请求，需要附带自己的节点地址（这和系统提供的客户端随机port不同）
func (n *TCPNode) SendGetAddrs(to string) {
	payload, _ := utils.GobEncode(common.GetAddrsMsg{AddrFrom:n.Addr.String()})
	req := append([]byte{CmdGetAddrs}, payload...)
	_ = n.SendData(to, req)
	log.Info("TCPNode_SendGetAddrs: send GetAddrsMsg to %s", to)
}

func (n *TCPNode) HandleGetAddrs(req []byte) {
	// 检查GetAddrMsg
	getAddrsMsg := &common.GetAddrsMsg{}
	_ = gob.NewDecoder(bytes.NewReader(req)).Decode(getAddrsMsg)
	gamFrom := getAddrsMsg.AddrFrom

	// 1. 检查来者的节点地址(不是client地址)
	// 如果from不诚实或者未注册
	if !n.EAddrs.IsAddrStrHonest(gamFrom) {
		return
	}

	// 2. 收集本地存有的节点列表，返回
	n.SendAddrs(gamFrom)
}

// 作为客户端发起请求，需要附带自己的节点地址（这和系统提供的客户端随机port不同）
func (n *TCPNode) SendAddrs(to string) {
	addrs := n.EAddrs.ValidAddrs()		// 发给别人没必要排序，因为不具备参考性
	payload, _ := utils.GobEncode(common.AddrsMsg{AddrFrom:n.Addr.String(), LocalAddrs:addrs})
	req := append([]byte{CmdAddrs}, payload...)
	_ = n.SendData(to, req)
	log.Info("TCPNode_SendAddrs: send addrs %v to %s", addrs, to)
}

// 接收对方发来的Addrs
func (n *TCPNode) HandleAddrs(req []byte) {
	// 检查AddrsMsg
	addrsMsg := &common.AddrsMsg{}
	_ = gob.NewDecoder(bytes.NewReader(req)).Decode(addrsMsg)
	amFrom, addrs := addrsMsg.AddrFrom, addrsMsg.LocalAddrs

	log.Info("TCPNode_HandleAddrs: received addrs %v", addrs)

	// 接下来需要将addrs与本地Addrs合并
	n.EAddrs.MergeAddrs(addrs)

	// TODO: 对方正常响应你的请求，是否考虑记录合规行为
	n.EAddrs.RecordStr(amFrom, eaddr.GoodAddrs)
}

//====================================EAddrs=======================================

//====================================Version=======================================

// VersionMsg的逻辑是：
// 当节点新上线请求同步区块链时，需先发送Verion消息(主要记录当前链的长度)给对方，
// 对方返回Version消息。
// HandleVersion: 如果自己更长，version回发，等待对方请求进一步的数据；如果对方更长，

func (n *TCPNode) SendVersion(to string) {

}

func (n *TCPNode) HandleVersion(req []byte) {

}


func (n *TCPNode) SendInventory(to string) {

}

func (n *TCPNode) HandleInventory(req []byte) {

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


