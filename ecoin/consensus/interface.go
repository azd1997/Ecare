package consensus

// consensus 包初步设计将会实现 PoT(自研)、PoW、Raft（强一致性共识的代表）、PBFT
// 该包依赖于点对点网络通信包和区块链包

// Consensus 共识协议接口
// 是否是共识节点以及是否拥有权限加入共识由其他包，主要是上层去实现、决定
// 这里只涉及共识节点之间的共识规则。也就是说，默认为共识节点调用此包
type Consensus interface {
	// 竞争出块权。PoT中是先发送
	CompeteForWin() bool
	ValidateWinnerBlock() error
}