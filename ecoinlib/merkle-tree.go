package ecoin

import "crypto/sha256"

// MerkleTree 默克尔树
type MerkleTree struct {
	RootNode *MerkleNode
}

// MerkleNode 默克尔节点
type MerkleNode struct {
	Left  *MerkleNode
	Right *MerkleNode
	Data  []byte
}

// 根节点、中间节点、叶节点

// 非叶节点 传入left、right均不能为空，data为空
func NewMerkleNode(left, right *MerkleNode) *MerkleNode {
	node := MerkleNode{}
	prevHashes := append(left.Data, right.Data...)
	hash := sha256.Sum256(prevHashes)
	node.Data = hash[:]
	node.Left = left
	node.Right = right
	return &node
}

// 叶节点 对于叶节点，传入left/right为空，data为序列化的交易
func NewMerkleLeafNode(data []byte) *MerkleNode {
	node := MerkleNode{}
	hash := sha256.Sum256(data)
	node.Data = hash[:]
	return &node
}

// 根据交易数组创建MerkleTree
func NewMerkleTree(data [][]byte) *MerkleTree {

	// 更正：merkle根哈希值存储于区块头

	// 区块体中应包含一颗默克尔树和一个交易数组。当然也可以将叶节点Data中存储交易数据
	// 之所以这么做是MerkleRoot作证明，每个节点接收到区块收到区块后根据交易数组重新生成默克尔树，比较Root判断有无篡改，
	// 进一步可以根据中间节点哈希比较找出是哪个地方被修改了。这也是为什么不直接取交易列表哈希的原因

	// 交易数组中交易数为奇数时复制最后一个
	l := len(data)
	if l == 0 {		// 一旦传入长度为0的data, 最后 &nodes[0]会panic
		return nil
	}
	if l%2 == 1 {
		data = append(data, data[l-1])
	}

	// 根据交易列表生成叶节点集合
	var nodes []MerkleNode // 叶节点值切片
	var node *MerkleNode   // 节点指针
	for _, dat := range data {
		node = NewMerkleLeafNode(dat)
		nodes = append(nodes, *node)
	}

	// 根据叶节点集合生成默克尔树
	for i := 0; i < l/2; i++ {
		var level []MerkleNode
		for j := 0; j < len(nodes); j += 2 {
			node = NewMerkleNode(&nodes[j], &nodes[j+1])
			level = append(level, *node)
		}
		nodes = level // nodes 变成其上一层的节点集合，不停迭代，直至nodes中只包含一个节点也就是根节点
	}

	// 构建MerkleTree结构体
	return &MerkleTree{RootNode: &nodes[0]}
}
