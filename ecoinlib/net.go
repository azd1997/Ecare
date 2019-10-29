package ecoin

import (
	"net"
)

// network

// StartServer 启动节点
func (gsm *GlobalStateMachine) StartServer() {
	// 开启监听
	ln, err := net.Listen(PROTOCOL, gsm.Opts().NodeAddress().String())
	if err != nil {
		LogErr("StartServer", err)
	}
	defer ln.Close()

	// 打开本地数据库获取区块链
	gsm.Ledger, err = ContinueChain(&ContinueChainArgs{opts: gsm.Opts()})

	// TODO： 开辟协程，处理请求
	// TODO：（尤其是对于普通用户）开辟携程，隔一段时间查看一次区块链，看有没有与自己有关的交易，收集并处理
}

// HandleConnection 处理连接
func (gsm *GlobalStateMachine) HandleConnection() {

}




