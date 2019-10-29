package ecoin

// Pot 交易量证明
type Pot struct {
	Num  uint // 交易量
	Hash Hash // 交易总验证码
}

/*********************************************************************************************************************
                                                    Workers相关
*********************************************************************************************************************/

// Worker 工人，指转发节点。这里维护转发节点地址所关联的一些信息
type Worker struct {
	addr Address
	pot Pot
	userID UserID
}

// WorkersMetaData 转发节点元信息集合
type WorkersMetaData struct {
	wmap map[string]*Worker
}

// Set 设置键值对
func (wm *WorkersMetaData) SetPot(addr string, pot Pot) {
	wm.wmap[addr].pot = pot
}

// Del 删除。 如果某个节点不诚实了，其余节点在这个表里删除其键
func (wm *WorkersMetaData) DelWorker(addr string) {
	delete(wm.wmap, addr)
}

// Reset 键不动，值全部重置为0
func (wm *WorkersMetaData) ResetPot() {
	for k := range wm.wmap {
		wm.wmap[k].pot = Pot{}
	}
}

// AddWorker 添加worker
func (wm *WorkersMetaData) AddWorker(worker *Worker) {
	wm.wmap[worker.addr.String()] = worker
}

//// 临时关闭某节点Pot权利，仅用于当该节点本来获取到POT但意外掉线，其他节点需
//func (wmap *WorkersMetaData) DisableWorkerPot(addr string) {
//
//}
// TODO: 不需要这个功能，period结束之后period/2节点仍未接收到新区块，则广播其他节点询问inv看其他人是否收到。如果没收到则大家重新POT
// 这有个问题，因为出块节点是维护时间线的重要角色。一旦他挂掉，其他节点谁来决定重新POT呢？
// 为了解决这个问题，修改WinPOT，返回Pot

// WinPOT 查看指定转发节点地址是否赢得POT。赢得则会返回最大交易数，没有则会返回0
// 存在情况：所有worker只收集到0个交易，也不会打包区块，因为所有人都没有POT
// 因为是不是赢得POT只看交易数了而不去看节点地址。所以对于本来应该获得POT的节点，其应该是尽量维护这个规则，不然其他人不会承认
func (wm *WorkersMetaData) WinPOT(addr string) uint {
	// addr如果不存在，则对应的*Worker会是nil，如果再去查
	return 0
}