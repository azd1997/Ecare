package types

import (
	"errors"
	"github.com/azd1997/Ecare/common/ecoinlib/utils"
	"github.com/dgraph-io/badger"
)

// 封装结构体还有个好处是，可以隐藏内部实现细节，通过方法访问内部数据，有利于后期优化底层的数据结构

// UnCompleteTXPool 未完成交易池
type UnCompleteTXPool struct {
	Map map[Hash]TX // 一旦前部交易的新进展被加进来了，前部交易被删除
}

// SaveFile 保存到文件
func (uctxp *UnCompleteTXPool) SaveFile(filePath string) (err error) {

	// todo

	return nil
}

// LoadFile 保存到文件
func (uctxp *UnCompleteTXPool) LoadFile(filePath string) (err error) {

	uctxp1 := &UnCompleteTXPool{}
	if uctxp != uctxp1 {
		return ErrLoadFileNeedEmptyReceiver
	}
	// TODO

	return nil
}

/*********************************************************************************************************************
                                                    UBTXP相关
*********************************************************************************************************************/

// UnBlockedTXPool 未打包交易池。检验合格的交易才会存入。但是由于可能存在存入时余额等符合条件，但是出块时（可能不是他第一个出块）又不符合条件
// 所以需要在出块前再检查一次
// 遇上这种情况原本的转账者如何得知？只要原本的转账者余额没发生变化，交易就仍有机会打包，。如果余额变化了，认为交易失效
type UnBlockedTXPool struct {
	list []TX
}

// Add 添加新的待出块交易
func (ubtxp *UnBlockedTXPool) Add(tx TX) {
	ubtxp.list = append(ubtxp.list, tx)
}

// Del 移除交易
// 出块时可能会有池中交易出现失效的情况，此时直接当出完块应该将池置空
func (ubtxp *UnBlockedTXPool) Clear() {
	ubtxp.list = []TX{}
}

// All 取出所有交易
func (ubtxp *UnBlockedTXPool) All() []TX {
	return ubtxp.list
}


/*********************************************************************************************************************
                                                    TBTXP相关
*********************************************************************************************************************/

// UnBlockedTXPool 未打包交易池。检验合格的交易才会存入。但是由于可能存在存入时余额等符合条件，但是出块时（可能不是他第一个出块）又不符合条件
// 所以需要在出块前再检查一次
// 遇上这种情况原本的转账者如何得知？只要原本的转账者余额没发生变化，交易就仍有机会打包，。如果余额变化了，认为交易失效
type ToBlockedTXPool struct {
	list []TX
}

// Add 添加新的待出块交易
func (tbtxp *ToBlockedTXPool) Add(tx TX) {
	tbtxp.list = append(tbtxp.list, tx)
}

// Del 移除交易
// 出块时可能会有池中交易出现失效的情况，此时直接当出完块应该将池置空
func (tbtxp *ToBlockedTXPool) Clear() {
	tbtxp.list = []TX{}
}

// All 取出所有交易
func (tbtxp *ToBlockedTXPool) All() []TX {
	return tbtxp.list
}

/*********************************************************************************************************************
                                                    POT相关
*********************************************************************************************************************/

// Pot 交易量证明
type Pot struct {
	Num  uint // 交易量
	Hash Hash // 交易总验证码
}

// ProofOfTransactionsMap POT证明表
// 出块时和接收到新区快并检验合格后，将之Reset，开启POT竞争后开始更新
type ProofOfTransactionsMap struct {
	pmap map[string]Pot// k - 节点地址； v - 节点收集的交易数
}

// Set 设置键值对
func (potmap *ProofOfTransactionsMap) Set(addr string, pot Pot) {
	potmap.pmap[addr] = pot
}

// Del 删除。 如果某个节点不诚实了，其余节点在这个表里删除其键
func (potmap *ProofOfTransactionsMap) Del(addr string) {
	delete(potmap.pmap, addr)
}

// Reset 键不动，值全部重置为0
func (potmap *ProofOfTransactionsMap) Reset() {
	for k := range potmap.pmap {
		potmap.pmap[k] = Pot{}
	}
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
func (wmap *WorkersMetaData) SetPot(addr string, pot Pot) {
	wmap.wmap[addr].pot = pot
}

// Del 删除。 如果某个节点不诚实了，其余节点在这个表里删除其键
func (wmap *WorkersMetaData) DelWorker(addr string) {
	delete(wmap.wmap, addr)
}

// Reset 键不动，值全部重置为0
func (wmap *WorkersMetaData) ResetPot() {
	for k := range wmap.wmap {
		wmap.wmap[k].pot = Pot{}
	}
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
func (wmap *WorkersMetaData) WinPOT(addr string) uint {
	// addr如果不存在，则对应的*Worker会是nil，如果再去查

}

/*********************************************************************************************************************
                                                    GSM相关
*********************************************************************************************************************/

// GlobalStateMachine 全局状态机，所有节点都需要同步和维护的变量，包括系统账户表、区块链账本
type GlobalStateMachine struct {
	// accounts/ledger/addrlist为区块链系统需要维持同步的变量，是真正的全局状态机

	// 全局账户状态表
	Accounts *EcoinAccounts

	// 全局转发节点信息表
	Workers *WorkersMetaData

	// 区块链账本
	Ledger *Chain

	// 未完成交易池
	UCTXP *UnCompleteTXPool
	// 关于UCTXP。正常运行时接收区块，部分交易进入UCTXP，遇见UCTXP的结束条件就去处理。
	// 加入掉线了，区块不会新接收，UCTXP也不会更新，等到接收新区快时，才会重新更新UCTXP，但是要注意uctxp要持久化到本地，进来一个写一个，
	// 保证一旦掉线重上线后UCTXP能够恢复。

	// 未打包交易池
	UBTXP *UnBlockedTXPool
	// UBTXP不需要持久化，因为如果掉线了，交易早就被别人打包了。

	// POT竞选后的交易池。
	// 出块者出块时清空其，接收区块者需要遍历TBTXP和UBTXP，并作清除
	TBTXP *ToBlockedTXPool

	// POT竞争表
	POT *ProofOfTransactionsMap
	// 不需要持久化

	// 节点地址集合
	Addrs *AddrLists
	// 需要持久化

	// 分析一下这些map：其中uctxp、ubtxp、pot存的都是指针表或者是整型表，而且每张表中的大小基本是固定的，不算太大，用map没大的问题。
	// 但是accounts表是需要不断添加新的账户的，几乎是只增不减，用map占用内存会爆炸。这是一个问题

	// 另一个问题是，这些表都是会不断更新的，直接全放在这个结构体里，就会申请连续空间，这样会使得整体的内存会不断重新分配，不好。

	// 为什么不直接 type EcoinAccountsMap map[string]*EcoinAccount??? 因为map的指针无法按key查值，必须先通过指针把map赋给一个新的变量才行。所以选择将map包装成一个struct

	// TODO: 这里先不管这些

	// 配置，各个节点Opts不同，不过因为不把opts放过来，很多方法实现会很麻烦
	opts *Option
}

// NewGlobalStateMachine 创建新的GSM，根据传入opts来获取GSM作为本地节点的唯一GSM
func NewGlobalStateMachine(opts *Option) (gsm *GlobalStateMachine, err error) {

	// 1. 从本地读取EcoinAccountMap。如果本地没有则创建空表
	var accounts *EcoinAccounts
	if accounts, err = CreateEcoinAccounts(opts.Port(), opts.EcoinAccountFilePathTemp()); err != nil {
		// err 不为空， 说明从配置的路径创建失败了。那么accounts就手动建一个空表
		accounts = &EcoinAccounts{Map: map[string]*EcoinAccount{}}
	}
	// err 为空说明从文件读取ecoinaccounts成功（尽管里边可能也是一个账户都没有）

	// 2. 从本地数据库继续区块链，如果没有，并不时调用initchain，而是创建一个零值的chain，用来调用方法接受同步而来的区块
	var ledger *Chain
	if ledger, err = ContinueChain(&ContinueChainArgs{Opts: opts}); err != nil {
		badgerOpts := badger.DefaultOptions(opts.DbPath())
		db, err := openDB(badgerOpts)
		if err != nil {
			return nil, err
		}
		ledger = &Chain{
			LastHash: Hash{},
			Db:       db,
		} // TODO：notice : 这里要修改，Db是需要创建的，根据opts
	}
	// err为空说明chain成功获取

	// 3. 恢复UCTXP
	uctxp := &UnCompleteTXPool{}
	uctxp.LoadFile(opts.UctxpFilePath()) // 不管有没有err

	// 4. 恢复addrs
	addrs := &AddrLists{}
	addrs.LoadFile(opts.AddrsFilePath())

	// ubtxp、pot不用持久化

	return &GlobalStateMachine{
		Accounts: accounts,
		Ledger:   ledger,
		UCTXP:    uctxp,
		UBTXP:    &UnBlockedTXPool{},
		POT:      &ProofOfTransactionsMap{},
		Addrs:    addrs,
	}, nil
}

// todo: 思考一下： 有必要在gsm去创建方法做这些事情吗？

/*balance相关*/
// GetBalanceOfUserID 获取账户地址的余额，不作账户地址存在与否的检查，如不存在，返回默认值0，存在也有可能返回0
func (gsm *GlobalStateMachine) GetBalanceOfUserID(userID string) Coin {
	return gsm.Accounts.Map[userID].balance
}

// 更新余额表，单次单个地址
func (gsm *GlobalStateMachine) UpdateBalanceOfUserID(userID string, newBalance Coin) {
	gsm.Accounts.Map[userID].balance = newBalance
}

// 批量更新余额表
func (gsm *GlobalStateMachine) UpdateBalanceOfUserIDs(userIDs []string, newBalances []Coin) error {
	if len(userIDs) != len(newBalances) {
		return errors.New("UpdateBalanceOfAddresses: 不等长度的addr切片和balance切片")
	}
	for i, id := range userIDs {
		gsm.Accounts.Map[id].balance = newBalances[i]
	}
	return nil
}

/*检查账户是否存在*/
// 某个账户是否存在
func (gsm *GlobalStateMachine) HasUserID(userID string) bool {
	_, ok := gsm.Accounts.Map[userID]
	return ok
}

/*role相关*/
// 返回role信息(不能返回指针，因为不允许更改)
func (gsm *GlobalStateMachine) GetRoleOfUserID(userID string) Role {
	return gsm.Accounts.Map[userID].role
}

/*available相关*/
func (gsm *GlobalStateMachine) IsUserIDAvailable(userID string) bool {
	return gsm.Accounts.Map[userID].available
}

func (gsm *GlobalStateMachine) GetPubKeyOfUserID(userID string) []byte {
	return gsm.Accounts.Map[userID].pubKey
}

/*********************************************************************************************************************
                                                    TX相关
*********************************************************************************************************************/

// NewTX 新建交易
func (gsm *GlobalStateMachine) NewTX(typ uint, args ArgsOfNewTX) (tx TX, err error) {
	// 检验参数。这里已经做了账户角色的验证
	if err = args.CheckArgsValue(gsm); err != nil {
		return nil, utils.WrapError("newTX", err)
	}

	// 构造交易
	return newTransaction(typ, args)
}

// VerifyTX 验证交易
func (gsm *GlobalStateMachine) VerifyTX(tx TX) (err error) {
	return tx.IsValid(gsm)
}

/*********************************************************************************************************************
                                                    Block相关
*********************************************************************************************************************/

// VerifyBlock 验证区块
func (gsm *GlobalStateMachine) VerifyBlock(block *Block) (err error) {
	return block.IsValid(gsm)
}

// MineBlock 出块
func (gsm *GlobalStateMachine) MineBlock(txs []TX) (b *Block, err error) {
	// 只有竞争到pot后才可以打包交易并出块，继而广播出去
	return gsm.Ledger.MineBlock(txs, gsm)
}

// AddBlock 添加区块
func (gsm *GlobalStateMachine) AddBlock(block *Block) (err error) {
	// 收到新的区块后先去验证区块，再添加到本地
	return gsm.Ledger.AddBlock(block)
}

/*********************************************************************************************************************
                                                    Option相关
*********************************************************************************************************************/

// Opts 获取Option
func (gsm *GlobalStateMachine) Opts() *Option {
	return gsm.opts
}

// SetOpts 设置选项
func (gsm *GlobalStateMachine) SetOpts(opts *Option) {
	gsm.opts = opts
}

/*********************************************************************************************************************
                                                    UBTXP相关
*********************************************************************************************************************/

// CheckUBTXP 检查UBTXP
// 检查发生在MineTX和SendPot时，SendPot发生在period/2时期，MineTX发生在period结束时。
// SendPot检查完之后需要把这部分交易给独立出来。因此UBTXP需要再增加一个交易池，命名为TBUXP表示待出块交易池，用于Pot竞争时存储交易
// 实际上应该是只有Pot时需要检查。
// 所以这个方法检查UBTXP并将有效交易存至TBTXP
func (gsm *GlobalStateMachine) CheckUBTXPAndUpdateTBTXP() {
	// 遍历交易池，将验证合格后的交易存入txs
	for _, tx := range gsm.UBTXP.All() {
		if err := tx.IsValid(gsm); err != nil {
			continue 	// 有问题的跳过
		}
		gsm.TBTXP.list = append(gsm.TBTXP.list, tx)
	}
	// 清空UBTXP
	gsm.UBTXP.Clear()
}