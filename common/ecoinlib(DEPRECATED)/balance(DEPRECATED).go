package ecoinlib

//
//import (
//	"errors"
//)
//
//// state-tree的临时替代品，直接使用哈希表来维护账户状态
//
//// 哈希表的特点是：查询更新方便，但是无法提供MerkleProof，无法证明账户余额
//// 而直接使用Merkle tree的问题在于查找更新麻烦
//
//type Balance uint
//
//// 在真正构建区块链节点的程序中，需要构建一个Balance表
//type balanceMap struct {
//	bmap map[Address]balance
//}
//
//func newBalanceMap() *balanceMap {
//	return &balanceMap{bmap: map[Address]balance{}}
//}
//
//// 获取账户地址的余额，不作账户地址存在与否的检查，如不存在，返回默认值0，存在也有可能返回0
//func (b *balanceMap) GetBalanceOfAddress(addr Address) balance {
//	return b.bmap[addr]
//}
//
//// 更新余额表，单次单个地址
//func (b *balanceMap) UpdateBalanceOfAddress(addr Address, newBalance balance) {
//	b.bmap[addr] = newBalance
//}
//
//// 批量更新余额表
//func (b *balanceMap) UpdateBalanceOfAddresses(addrs []Address, newBalances []balance) error {
//	if len(addrs) != len(newBalances) {
//		return errors.New("UpdateBalanceOfAddresses: 不等长度的addr切片和balance切片")
//	}
//	for i, a := range addrs {
//		b.bmap[a] = newBalances[i]
//	}
//	return nil
//}
//
//// 某个账户是否存在
//func (b *balanceMap)
