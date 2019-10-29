package ecoin

// TODO: 转发节点对某个账户所作的封禁等操作，这些将封装为一类新的交易，通过区块来同步到整个网络
// TODO: 由账户持有者本人进行的信息更新(name,phone,institution...)。怎么处理？
// TODO: 可以这么做：注册账户时，把注册信息存到ecoinaccout中，更新注册信息时往AccountMsg更改注册信息，但其余置空，但留下userid string用来做标识

// AccountMsg 新建账户时需要发送此消息。
type AccountMsg struct {
	EA *EcoinAccount
	AddrFrom string
}

// SendAccountMsg 发送账户消息
func (gsm *GlobalStateMachine) SendAccountMsg(to string, ea *EcoinAccount) error {

}

// BroadcastAccountMsg 广播账户消息
func (gsm *GlobalStateMachine) BroadcastAccountMsg(addrs []string, ea *EcoinAccount) error {

}

// HandleAccountMsg 处理账户消息。本地accounts表若没有则添加
func (gsm *GlobalStateMachine) HandleAccountMsg(request []byte) {

}
