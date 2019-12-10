package p2ptcp

import (
	"github.com/azd1997/ego/enet/etcp"
)

type TCPNode struct {
	Ip string
	Port int
	Name string
}

func NewP2PTCPNode(args *Args) *TCPNode {

}

// Start 启动节点TCP服务器，执行初始化流程，handle连接
func (n *TCPNode) Start() {
	// 创建服务器
	etcpOpts := etcp.DefaultOption().SetHost(n.Ip).SetPort(n.Port).SetName(n.Name)
	server := etcp.NewServer(etcpOpts)

	// 注册连接Hook函数
	server.SetOnConnStart(checkNodeFirst)

	// 注册多路由
	server.AddRouter(UnknownMsg, &unknownMsgRouter{})

	// 启动server
	server.Serve()

}

func checkNodeFirst(conn etcp.IConnection) {

}