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

	// 待传输区块队列
	BlockInTransit []singlechain.Block
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
	case CmdGetAddrs:
		n.HandleGetAddrs(req[1:])
	case CmdAddrs:
		n.HandleAddrs(req[1:])
	case CmdVersion:
		n.HandleVersion(req[1:])
	case CmdGetInventory:
		n.HandleGetInventory(req[1:])
	case CmdInventory:
		n.HandleInventory(req[1:])
	case CmdBlock:
		n.HandleBlock(req[1:])
	}
}

//====================================CheckAddr=======================================

// 检查并处理发信人的地址，返回发信人是否诚实，不诚实，就别往下执行了
func (n *TCPNode) checkAndHandleAddrFrom(addr string) (honest bool) {
	contained, honest, reachable := n.EAddrs.ContainsAndHonestAndReachable(addr)
	if !contained {
		// 加入到本地
		n.EAddrs.AddAddrStr(addr)
		return true	// 初始化为诚实
	}
	if !honest {
		return false
	}
	if !reachable {		// 原本包含该地址但不可达，现在需要修改状态
		n.EAddrs.SetEAddrReachable(addr, true)
	}
	return true
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
	from, pong := pingmsg.AddrFrom, pingmsg.Pong
	// 返回resp
	if pong {
		// 处理pong
		n.EAddrs.EAddrPingStopStr(from)
	} else {
		pongmsg := &common.PingMsg{AddrFrom:n.Addr.String(), Pong:true}
		payload, _ := utils.GobEncode(pongmsg)
		resp := append([]byte{CmdPing}, payload...)
		_ = n.SendData(from, resp)
	}
}


//====================================GetAddrs=======================================

// Addrs的场景
// 1. 节点上线时向seed节点(可能是写死的或者自选的)请求节点列表，seed节点随之将其节点列表打包发回
// 2. 节点注册时，藉由注册命令，seed节点会将该节点地址向其他节点扩散。
// TODO: 尽管注册行为使用msg更方便，但使用交易实现会更体现行为历史可追溯的特性
// TODO: 注册流程：本地新建账号，本地监听端口(仅H/R需要)，将账号信息、节点地址信息、身份信息
//  使用注册医院的公钥进行加密（患者可加密，其他角色需要公示），构建交易，发往工人组，此称为注册交易。
//  工人节点接收到注册交易时不作其他检查，只检查是否来自该地址IP是否过于频繁的注册
//  （注册交易也用于更新账户信息）。
//  当注册医院收到这个链上确认的注册交易后，如果是患者的注册，它需要解密患者身份信息，记录下来


// TODO: 节点地址和账户不应该绑定，节点的问题封节点，账户的问题封账户
// 账户必须注册，注册通过发送注册交易，，而后被所有共识组缓存到本地，这称为注册账户表
// 每次检查交易/区块之前要检查构建者及接收者是否已注册
// 至于节点，节点只会拦截搞了破坏的节点的请求
// 节点不要求注册，账号要求注册

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

	// 1. 判断该地址是否有不良记录
	if honest := n.checkAndHandleAddrFrom(gamFrom); !honest {
		return
	}

	// 2. 收集本地存有的节点列表，返回
	n.SendAddrs(gamFrom)
}

//====================================Addrs=======================================


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

	// 0. 判断该地址是否有不良记录
	if honest := n.checkAndHandleAddrFrom(amFrom); !honest {
		return
	}

	log.Info("TCPNode_HandleAddrs: received addrs %v", addrs)

	// 接下来需要将addrs与本地Addrs合并
	n.EAddrs.MergeAddrs(addrs)

	// TODO: 对方正常响应你的请求，是否考虑记录合规行为
	n.EAddrs.RecordStr(amFrom, eaddr.GoodAddrs)
}


//====================================Version=======================================

// VersionMsg的逻辑是：
// 当节点新上线请求同步区块链时，需先发送Verion消息(主要记录当前链的长度)给对方，
// 对方返回Version消息。
// HandleVersion: 如果自己更长，version回发，等待对方请求进一步的数据；如果对方更长，

func (n *TCPNode) SendVersion(to string) {
	versionMsg := common.VersionMsg{
		AddrFrom:n.Addr.String(),
		NodeVersion:n.Version,
		MaxBlockID:n.Chain.MaxBlockID,
		LatestBlockHash:n.Chain.LastHash,
		SecondLatestBlockHash:n.Chain.SecondLastHash,
	}
	payload, _ := utils.GobEncode(versionMsg)
	req := append([]byte{CmdVersion}, payload...)
	_ = n.SendData(to, req)
}

func (n *TCPNode) HandleVersion(req []byte) {
	// 解码请求
	versionMsg := &common.VersionMsg{}
	_ = gob.NewDecoder(bytes.NewReader(req)).Decode(versionMsg)

	// 处理VersionMsg : B收到A的VersionMsg:
	// B比A长：检查A的Maxid在B本地链上的区块，是否一致，
	// 		一致，则将自身的verion返回，等对方请求自己的inventory
	//		不一致，将自身version返回，并且自己向其他所有节点请求version消息
	//		（这里需要一个临时的集合，用来收集所有回返信息，最后多者作为正确答案）
	// A比B长：B先检查A的两个哈希是否能直接校验，不能再向对方请求inventory

	// 0. 判断该地址是否有不良记录
	if honest := n.checkAndHandleAddrFrom(versionMsg.AddrFrom); !honest {
		return
	}

	// 1. 比较NodeVersion,不等于就是错了，不响应
	if versionMsg.NodeVersion != n.Version {
		return
	}

	// TODO: 暂时不使用VersionMsg的其他字段

	// 2. 比较MaxID
	if versionMsg.MaxBlockID == n.Chain.MaxBlockID {
		// 暂时 do nothing
	} else if versionMsg.MaxBlockID > n.Chain.MaxBlockID {
		// 请求区块存证
		n.SendGetInventory(versionMsg.AddrFrom, common.InvBlock)
	} else {	// 自身比对方链长
		n.SendVersion(versionMsg.AddrFrom)
	}
}

//====================================GetInventory=======================================


// 向目标发送获取区块或存证消息
func (n *TCPNode) SendGetInventory(to string, invType uint8) {
	giMsg := &common.GetInventoryMsg{AddrFrom:n.Addr.String(), InvType:invType}
	payload, _ := utils.GobEncode(giMsg)
	req := append([]byte{CmdGetInventory}, payload...)
	_ = n.SendData(to, req)
}

// 处理GetInventory消息
func (n *TCPNode) HandleGetInventory(req []byte) {
	// 解码
	giMsg := &common.GetInventoryMsg{}
	_ = gob.NewDecoder(bytes.NewReader(req)).Decode(giMsg)

	// 0. 判断该地址是否有不良记录
	if honest := n.checkAndHandleAddrFrom(giMsg.AddrFrom); !honest {
		return
	}

	// 处理
	switch giMsg.InvType {
	case common.InvBlock:
		n.SendInventory(giMsg.AddrFrom, common.InvBlock)
	case common.InvTx:
		// 用于收集到新交易同步
	default:
		//TODO: 错误的存证代号，作出相应惩处
	}
}

//====================================Inventory=======================================

// 发送存证信息
func (n *TCPNode) SendInventory(to string, invType uint8) {
	invMsg := &common.InventoryMsg{AddrFrom:n.Addr.String(), InvType:invType}
	switch invType {
	case common.InvBlock:
		// 暂且一股脑将所有区块哈希发过去
		hashes, _ := n.Chain.GetBlockHashes()
		invMsg.Invs = hashes
	case common.InvTx:
		// TODO
	}

	payload, _ := utils.GobEncode(invMsg)
	req := append([]byte{CmdInventory}, payload...)
	_ = n.SendData(to, req)
}


func (n *TCPNode) HandleInventory(req []byte) {
	// 解码
	invMsg := &common.InventoryMsg{}
	_ = gob.NewDecoder(bytes.NewReader(req)).Decode(invMsg)

	if honest := n.checkAndHandleAddrFrom(invMsg.AddrFrom); !honest {
		return
	}

	// 处理
	switch invMsg.InvType {
	case common.InvBlock:
		blockHashes := invMsg.Invs
		// 暂且不考虑对方与自己不在同一链的情况，收到存证之后，应该将存证加入一个待传输区块队列，
		// 这个队列维护在TCPNode中
		n.BlockInTransit

	case common.InvTx:
		// 待处理
	default:
		// 作恶，待处理
	}
}

//====================================111=======================================












