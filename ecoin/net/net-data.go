package net

import (
	"bytes"
	"fmt"
	"github.com/azd1997/Ecare/ecoin/eaddr"
	"github.com/azd1997/Ecare/ecoin/utils"
	"github.com/azd1997/Ecare/ecoinlib/log"
	"github.com/azd1997/ego/enet/etcp"
	"io"
	"net"
)

// Send 应该提供两种机制，一种是作为客户端尚未建立连接；一种是作为服务端/客户端已经建立连接直接写回数据

// SendData 发送数据。
// 调用之前先将数据封包好
func (n *TCPNode) SendData(to eaddr.Addr, data []byte) error {

	// to应当之前就在localNodeList中。在发送数据时检查是不是已有节点并且找到其位置
	self := eaddr.Addr{
		Ip:   n.Server.Option().Host,
		Port: n.Server.Option().TcpPort,
	}
	if to == self {
		return utils.WrapError("SendData", ErrSendToSelf)
	}

	// 检查节点可用与否
	if !to.IsValid(n.EAddrs) {
		return utils.WrapError("SendData", ErrInvalidAddr)
	}

	//向addr发起tcp连接
	conn, err := net.Dial(PROTOCOL, to.String())

	//连接不可用, 成为作恶情况之一（断网，情况轻微）， 记录作恶历史，连续三次不可用，标记为unreachable
	if err != nil {
		// 连接不可用时，调用EAddrs的方法记录作恶时间作恶类型等
		n.EAddrs.Record(to, eaddr.BadConnFail)

		return fmt.Errorf("SendData: %s: %s", ErrUnreachableNode, to)
	}

	defer conn.Close()

	//将data []byte复制一份通过conn发给对方
	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		return utils.WrapError("SendData", err)
	}
	return nil
}

// SendDataBack 回写数据
func (n *TCPNode) SendDataBack(conn net.TCPConn, data []byte) error {
	_, err := conn.Write(data)
	return err
}

// Broadcast 广播
func (n *TCPNode) Broadcast(addrs []eaddr.Addr, data []byte) {

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

// SendMsg 单独发送消息
func (n *TCPNode) SendMsg(to eaddr.Addr, data []byte) error {

	// 建立连接
	conn, err := net.Dial(PROTOCOL, to.String())
	if err != nil {
		return err
	}
	defer conn.Close()

	// 写数据
	if _, err = io.Copy(conn, bytes.NewReader(data)); err != nil {
		return err
	}

	return nil
}

// 一次读取全部数据。
func (n *TCPNode) UnpackData(conn *net.TCPConn) (cmd uint32, data []byte, err error) {
	dp := etcp.DataPack{MaxSize:1<<24}

	// 读head部分
	headData := make([]byte, dp.GetHeadLen())
	_, err = io.ReadFull(conn, headData)
	if err != nil {
		return CmdUnknown, nil, err
	}

	// 将head拆包到msg中
	msgHead, err := dp.Unpack(headData)
	if err != nil {
		return CmdUnknown, nil, err
	}

	// 继续向后处理
	if msgHead.GetDataLen() > 0 {
		msg := msgHead.(*etcp.Message)
		msg.Data = make([]byte, msg.GetDataLen())
		_, err = io.ReadFull(conn, msg.Data)
		if err != nil {
			return msg.Id, nil, err
		}
		return msg.Id, msg.Data, nil
	}

	return msgHead.GetMsgId(), nil, nil
}

// 打包数据
func (n *TCPNode) PackData(cmd uint32, data []byte) ([]byte, error) {
	dp := etcp.DataPack{MaxSize:1<<24}
	return dp.Pack(etcp.NewMessage(cmd, data))
}