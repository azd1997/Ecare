package net

import (
	"github.com/azd1997/Ecare/common/ecoinlib/types"
	"github.com/azd1997/Ecare/common/ecoinlib/utils"
	"net"
)

// network

// StartServer 启动节点
func (e *ecoin) StartServer() {
	// 开启监听
	ln, err := net.Listen(e.Opts().Protocol(), e.Opts().NodeAddress().NodeAddr())
	if err != nil {
		utils.LogErr("StartServer", err)
	}
	defer ln.Close()

	// 打开本地数据库获取区块链
	e.Ledger, err = types.ContinueChain(&types.ContinueChainArgs{Opts: e.Opts()})
}

// startServer 启动节点服务器
func startServer(protocol, nodeAddr, minerUserID string) (err error) {
	// 开启监听
	ln, err := net.Listen(protocol, nodeAddr)
	if err != nil {
		return err
	}
	defer ln.Close()

	// 打开本地数据库获取区块链

}




