package net

import (
	"bytes"
	"fmt"
	"github.com/azd1997/Ecare/ecoin/eaddr"
	"github.com/azd1997/Ecare/ecoin/erro"
	"github.com/azd1997/Ecare/ecoin/utils"
	"github.com/azd1997/Ecare/ecoinlib/log"
	"io"
	"net"
)

// Send 应该提供两种机制，一种是作为客户端尚未建立连接；一种是作为服务端/客户端已经建立连接直接写回数据
// TODO: 目前认知来看，想要通过总分式的Handle模式来处理各项请求，就需要将请求拆分独立

// TODO: 为了方便起见，所有请求均为一次性连接，发完消息就关闭
// 后期再考虑长连接
// 因此这里(net-data.go)的SendDataBack()避免使用


// SendDataWithCheck 一次性连接、附带节点地址检查、发送数据。
// 调用之前先将数据封包好
func (n *TCPNode) SendDataWithCheck(to string, data []byte) error {

	// to应当之前就在localNodeList中。在发送数据时检查是不是已有节点并且找到其位置
	self := n.Addr.String()
	if to == self {
		return utils.WrapError("SendDataWithCheck", erro.ErrSendToSelf)
	}

	// 检查节点可用与否
	if !n.EAddrs.IsAddrStrValid(to) {
		return utils.WrapError("SendDataWithCheck", erro.ErrInvalidAddr)
	}

	//向addr发起tcp连接
	conn, err := net.Dial(PROTOCOL, to)

	//连接不可用, 成为作恶情况之一（断网，情况轻微）， 记录作恶历史，连续三次不可用，标记为unreachable
	if err != nil {
		// 连接不可用时，调用EAddrs的方法记录作恶时间作恶类型等
		n.EAddrs.RecordStr(to, eaddr.BadConnFail)

		return fmt.Errorf("SendDataWithCheck: %s: %s", erro.ErrUnreachableNode, to)
	}

	defer conn.Close()

	//将data []byte复制一份通过conn发给对方
	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		return utils.WrapError("SendData", err)
	}
	return nil
}

// SendData 一次性、单独发送消息，不做检查，但是会对对方的可达情况进行更新
func (n *TCPNode) SendData(to string, data []byte) error {

	// 建立连接
	conn, err := net.Dial(PROTOCOL, to)
	if err != nil {
		// 连接不可用时，调用EAddrs的方法记录作恶时间作恶类型等
		n.EAddrs.RecordStr(to, eaddr.BadConnFail)
		return fmt.Errorf("SendDataWithCheck: %s: %s", erro.ErrUnreachableNode, to)
	}
	defer conn.Close()

	// 写数据
	if _, err = io.Copy(conn, bytes.NewReader(data)); err != nil {

		return err
	}

	return nil
}

// SendDataBack 持续使用连接，回写数据
func (n *TCPNode) SendDataBack(conn net.Conn, data []byte) error {
	_, err := conn.Write(data)
	return err
}

// Broadcast 广播
func (n *TCPNode) Broadcast(addrs []string, data []byte) {

	for _, addr := range addrs {
		// 并发处理 TODO： 连接池等优化 。 每个节点不仅要作为服务端和多个节点连接，还要作为客户端和多个服务端连接。所以客户端也一样要进行连接优化
		go func() {
			err := n.SendData(addr, data)
			if err != nil {
				log.Warn("Broadcast: %s", err)
				// TODO : 日志处理
			}
		}()
	}
}