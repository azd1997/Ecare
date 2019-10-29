package ecoin

// SendPing 发送Ping消息。用来测通信延迟和目标可用性
func (gsm *GlobalStateMachine) SendPing(to string) error {
	request := append(CmdToBytes("ping"), []byte(gsm.Opts().NodeAddress().String())...)
	if err := gsm.SendMsg(to, request); err != nil {
		return err
	}

	// TODO: 开启协程通知计时器
	// map[addrString]AddrData
	// type AddrData struct {
	// 		Metadata
	//      clock
	//      chan
	//}

	return nil
}

// HandlePing 处理
func (gsm *GlobalStateMachine) HandlePing(request []byte) {
	addrFrom := string(request[COMMAD_LENGTH:])
	response := append(CmdToBytes("pong"), []byte(gsm.Opts().NodeAddress().String())...)
	if err := gsm.SendMsg(addrFrom, response); err != nil {
		return
	}



	return
}

// HandlePong 处理pong
func (gsm *GlobalStateMachine) HandlePong(request []byte) {
	addrFrom := string(request[COMMAD_LENGTH:])

	// TODO: 通知计时协程，将addrFrom传过去
	addrFrom = addrFrom
}

// Clock 计时协程函数，在启动节点函数中调用
func (gsm *GlobalStateMachine) Clock() {

	go func() {

	}()


}


// TODO: 在广播发ping的时候就另起协程开始计时，接收每一个pong消息，计时并记录。根据结果对L1重排序

// TODO: PingPong 啥时候进行？每个节点都需要对自己存储的转发节点集合进行延迟排序，优先向前三者发起同步数据请求。

// PingPOng与addr关联。对于新上线节点，先向种子节点或者叫中间节点请求节点列表，与本地进行合并，合并后广播ping。收集到返回的延迟信息后