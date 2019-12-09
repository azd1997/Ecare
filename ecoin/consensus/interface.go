package consensus

// consensus 包初步设计将会实现 PoT(自研)、PoW、Raft（强一致性共识的代表）、PBFT
// 该包依赖于点对点网络通信包和区块链包

// Consensus 共识协议接口
// 是否是共识节点以及是否拥有权限加入共识由其他包，主要是上层去实现、决定
// 这里只涉及共识节点之间的共识规则。也就是说，默认为共识节点调用此包
// PoT中先通过一轮竞争确认自己是不是该出块；而后其余节点等待接收区块并检查区块
// Raft中初始全为follower，申请成为candidate，获过半票者成为Leader，Leader记录，其余节点接收不质疑（在我们这区块链里要改成检查区块，
// 发现Leader非法则主动制造分区把原先Leader排除，重选Leader）
// PoW所有节点都计算哈希难题，计算出来则拥有出块权，但这个出块正常情况下是有多个节点会出块的； 之后其他节点接收区块并检查
// PBFT
type Consensus interface {
	// 竞争出块权。PoT中是先发送
	Compete() bool
	ValidateWinnerBlock() error
}

// Proof 证明接口。证明一个节点拥有合法的出块权
// 不同的共识机制其Proof可能包含了不同的证明信息
// 例如： PoW 会返回Nonce、PoT会返回winnerPoT或者pot列表（未定）
type Proof interface {

}