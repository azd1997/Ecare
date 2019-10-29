package ecoin

import (
	"bytes"
	"fmt"
	"github.com/azd1997/Ecare/ecoinlib/log"
	"io"
	"net"
)

// SendData 发送数据。
// TODO： 原本在SendData这里是需要将发送失败的那些节点从节点集合中剔除，但是由于加入了周期性的ping，所以把这个工作放到ping中
func (gsm *GlobalStateMachine) SendData(to string, data []byte) error {

	var nodes []*Address
	var index int

	// to应当之前就在localNodeList中。在发送数据时检查是不是已有节点并且找到其位置
	self := gsm.Opts().NodeAddress()
	if to == self.String() {
		return WrapError("SendData", ErrSendToSelf)
	}
	//// 只有转发节点才需要进行检查可用与否
	//index =  NodeLocate(to, e.addrList.L1)
	//if index != -1 {
	//	nodes = e.addrList.L1
	//} else {
	//	return utils.WrapError("SendData", ErrUnKnownNode)
	//}
	// 叶节点集合不可用就不可用，不管

	//向addr发起tcp连接
	conn, err := net.Dial(PROTOCOL, to)

	//连接不可用，则更新已知节点集，将之移除
	if err != nil {
		// log.Error("%s: %s", utils.WrapError("SendData", ErrUnavailableNode), to)
		for i:=index; i<len(nodes)-1; i++ {
			nodes[i] = nodes[i+1]
		}
		// TODO:注意是否要写保护 wg.wait或别的。因为涉及并发连接
		gsm.Addrs.L1 = nodes

		return fmt.Errorf("SendData: %s: %s", ErrUnavailableNode, to)
	}

	defer conn.Close()

	//将data []byte复制一份通过conn发给对方
	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		return WrapError("SendData", err)
	}
	return nil
}

// Broadcast 广播
func (gsm *GlobalStateMachine) Broadcast(addrs []string, data []byte) {

	for _, addr := range addrs {
		// 并发处理 TODO： 连接池等优化
		go func() {
			err := gsm.SendMsg(addr, data)
			if err != nil {
				log.Warn("Broadcast: %s", err)
				// TODO : 日志处理
			}
		}()
	}
}

// SendMsg 单独发送消息
func (gsm *GlobalStateMachine) SendMsg(to string, data []byte) error {

	// 建立连接
	conn, err := net.Dial(PROTOCOL, to)
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