package ecoin

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/azd1997/Ecare/ecoinlib/log"
	"strconv"
)

// 封装结构体还有个好处是，可以隐藏内部实现细节，通过方法访问内部数据，有利于后期优化底层的数据结构



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

	// 全局区块链账本
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

	// 待发送区块哈希列表
	BlockToTransit []Hash

	// 节点地址集合
	Addrs *AddrLists
	// 需要持久化

	// 分析一下这些map：其中uctxp、ubtxp、pot存的都是指针表或者是整型表，而且每张表中的大小基本是固定的，不算太大，用map没大的问题。
	// 但是accounts表是需要不断添加新的账户的，几乎是只增不减，用map占用内存会爆炸。这是一个问题

	// 另一个问题是，这些表都是会不断更新的，直接全放在这个结构体里，就会申请连续空间，这样会使得整体的内存会不断重新分配，不好。

	// 为什么不直接 type EcoinAccountsMap map[string]*EcoinAccount??? 因为map的指针无法按key查值，必须先通过指针把map赋给一个新的变量才行。所以选择将map包装成一个struct

	// TODO: 这里先不管这些

	// 日志记录器
	Logger *Logger

	// 配置，各个节点Opts不同，不过因为不把opts放过来，很多方法实现会很麻烦
	opts *Option
}

// newGSM 获取GSM对象
func newGSM(opts *Option) (*GlobalStateMachine, error) {
	// 1. 获取account
	account, err := loadOrCreateAccount(opts.Port())
	if err != nil {
		return nil, WrapError("StartNode", err)
	}
	// 2. 构造Option
	opts = opts.SetAccount(*account)
	log.Success("主配置构建成功！ 准备继续区块链......")

	// 3. 继续区块链: 先检查本地有没有区块链数据库路径，有则继续区块链，没有则构建一个空的区块链
	var chain *Chain
	dbPath := fmt.Sprintf(CHAIN_DBPATH_TEMP, strconv.Itoa(int(opts.Port())))
	if DbExists(dbPath) {
		chain, err = ContinueChain(&ContinueChainArgs{opts:opts})
		if err != nil {
			return nil, WrapError("StartNode", err)
		}
		log.Success("继续区块链成功！ Chain.LastHash: %s", base64.StdEncoding.EncodeToString(chain.LastHash[:]))
	} else {
		chain, err = NewEmptyChain(dbPath)
		if err != nil {
			return nil, WrapError("StartNode", err)
		}
		log.Info("区块链数据库不存在，创造空区块链！ Chain.LastHash: %s (ZeroHash)", base64.StdEncoding.EncodeToString(chain.LastHash[:]))
	}

	// 4. 从本地读取EcoinAccountMap。如果本地没有则创建空表
	var accounts *EcoinAccounts
	if accounts, err = CreateEcoinAccountsFromJsonFile(opts.Port()); err != nil {
		log.Error("加载EcoinAccounts表失败: %s", err)
		log.Info("创建一个空的EcoinAccounts表")
		// err 不为空， 说明从配置的路径创建失败了。那么accounts就手动建一个空表
		accounts = &EcoinAccounts{Map: map[string]*EcoinAccount{}}
		// 由于是空表，所以不执行保存文件了。运行过程中有每一次有变动就更新完内存中的数据结构再保存到文件
	} else {
		// err 为空说明从文件读取ecoinaccounts成功（尽管里边可能也是一个账户都没有）
		log.Success("加载本地EcoinAccounts表成功")
	}

	// 5. 恢复UCTXP
	uctxp := &UnCompleteTXPool{}
	if err = uctxp.LoadFileWithGobDecode(opts.Port()); err != nil {
		log.Error("加载UCTXP失败: %s", err)
		log.Info("创建一个空的UCTXP")
		uctxp = &UnCompleteTXPool{Map: map[string]TX{}}
		// 由于是空表，所以不执行保存文件了。运行过程中有每一次有变动就更新完内存中的数据结构再保存到文件
	} else {
		log.Success("加载本地UCTXP成功")
	}

	// 6. 恢复addrs
	addrs := &AddrLists{}
	seedAddress := newAddress(opts.SeedNode(), "seed")
	if err = addrs.LoadFileWithJsonUnmarshal(opts.Port()); err != nil {
		log.Error("加载Addrs表失败: %s", err)
		log.Info("创建一个新的的Addrs表，仅含种子节点: [%s]", opts.SeedNode())
		addrs = &AddrLists{
			L1: []*Address{seedAddress},
			L2: []*Address{},
			L3: []*Address{},
		}
		// 保存到本地文件
		err = addrs.SaveFileWithJsonMarshal(opts.Port())
		if err != nil {
			// 这里如果保存出错是可以不管的
			log.Warn("Addrs表持久化失败: %s", err)
		}
	} else {
		log.Success("加载本地Addrs表成功")
		// seed是否在addrs内，不在就加进去
		if !addrs.L1HasAddress(seedAddress) {
			addrs.L1Add(seedAddress)
			log.Info("已将seed节点地址加入Addrs表")
		}
	}

	// 7. 初始化Workers。根据Addrs.L1生成
	workers := &WorkersMetaData{
		map[string]*Worker{},
	}
	var worker *Worker
	if len(addrs.L1) > 0 {
		for _, addr := range addrs.L1 {
			worker = &Worker{
				addr:   *addr,
				pot:    Pot{},
				userID: UserID{},
			}
			workers.AddWorker(worker)
		}
	}
	log.Success("根据Addrs.L1初始化Workers完成")

	// 8. 构造gsm
	gsm := &GlobalStateMachine{
		Accounts:       accounts,
		Workers:        workers,
		Ledger:         chain,
		UCTXP:          uctxp,
		UBTXP:          &UnBlockedTXPool{[]TX{}},
		TBTXP:          &ToBlockedTXPool{[]TX{}},
		BlockToTransit: []Hash{},
		Addrs:          addrs,
		Logger:         &Logger{},
		opts:           opts,
	}
	log.Success("构造本地GSM成功！")
	return gsm, nil
}

// NewGlobalStateMachine 创建新的GSM，根据传入opts来获取GSM作为本地节点的唯一GSM
//func NewGlobalStateMachine(opts *Option) (gsm *GlobalStateMachine, err error) {
//
//	// 1. 从本地读取EcoinAccountMap。如果本地没有则创建空表
//	var accounts *EcoinAccounts
//	if accounts, err = CreateEcoinAccountsFromJsonFile(opts.Port()); err != nil {
//		// err 不为空， 说明从配置的路径创建失败了。那么accounts就手动建一个空表
//		accounts = &EcoinAccounts{Map: map[string]*EcoinAccount{}}
//	}
//	// err 为空说明从文件读取ecoinaccounts成功（尽管里边可能也是一个账户都没有）
//
//	// 2. 从本地数据库继续区块链，如果没有，并不时调用initchain，而是创建一个零值的chain，用来调用方法接受同步而来的区块
//	var ledger *Chain
//	if ledger, err = ContinueChain(&ContinueChainArgs{opts: opts}); err != nil {
//		badgerOpts := badger.DefaultOptions(opts.DbPath())
//		db, err := openDB(badgerOpts)
//		if err != nil {
//			return nil, err
//		}
//		ledger = &Chain{
//			LastHash: Hash{},
//			Db:       db,
//		} // TODO：notice : 这里要修改，Db是需要创建的，根据opts
//	}
//	// err为空说明chain成功获取
//
//	// 3. 恢复UCTXP
//	uctxp := &UnCompleteTXPool{}
//	if err = uctxp.LoadFileWithJsonUnmarshal(opts.Port()); err != nil {
//		uctxp = &UnCompleteTXPool{Map: map[string]TX{}}
//	} // 不管有没有err
//
//	// 4. 恢复addrs
//	addrs := &AddrLists{}
//	if err = addrs.LoadFileWithJsonUnmarshal(opts.Port()); err != nil {
//		addrs = &AddrLists{
//			L1: []*Address{},
//			L2: []*Address{},
//			L3: []*Address{},
//		}
//	}
//
//	// ubtxp、pot不用持久化
//
//	// 5. 初始化workers
//	workers := &WorkersMetaData{
//		map[string]*Worker{},
//	}
//	var worker *Worker
//	if len(addrs.L1) > 0 {
//		for _, addr := range addrs.L1 {
//			worker = &Worker{
//				addr:   *addr,
//				pot:    Pot{},
//				userID: UserID{},
//			}
//			workers.AddWorker(worker)
//		}
//	}
//
//	return &GlobalStateMachine{
//		Accounts: accounts,
//		Workers:workers,
//		Ledger:   ledger,
//		UCTXP:    uctxp,
//		UBTXP:    &UnBlockedTXPool{},
//		TBTXP:&ToBlockedTXPool{},
//		BlockToTransit:[]Hash{},
//		Addrs:    addrs,
//		Logger:&Logger{},
//		opts:opts,
//	}, nil
//}

// todo: 思考一下： 有必要在gsm去创建方法做这些事情吗？

/*balance相关*/
// GetBalanceOfUserID 获取账户地址的余额，不作账户地址存在与否的检查，如不存在，返回默认值0，存在也有可能返回0
func (gsm *GlobalStateMachine) GetBalanceOfUserID(userID string) Coin {
	return gsm.Accounts.Map[userID].Balance()
}

// 更新余额表，单次单个地址
func (gsm *GlobalStateMachine) UpdateBalanceOfUserID(userID string, newBalance Coin) {
	gsm.Accounts.Map[userID].BalanceCoin = newBalance
}

// 批量更新余额表
func (gsm *GlobalStateMachine) UpdateBalanceOfUserIDs(userIDs []string, newBalances []Coin) error {
	if len(userIDs) != len(newBalances) {
		return errors.New("UpdateBalanceOfAddresses: 不等长度的addr切片和balance切片")
	}
	for i, id := range userIDs {
		gsm.Accounts.Map[id].BalanceCoin = newBalances[i]
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
func (gsm *GlobalStateMachine) GetRoleOfUserID(userID string) *Role {
	return gsm.Accounts.Map[userID].Role()
}

/*available相关*/
func (gsm *GlobalStateMachine) IsUserIDAvailable(userID string) bool {
	return gsm.Accounts.Map[userID].Available()
}

func (gsm *GlobalStateMachine) GetPubKeyOfUserID(userID string) []byte {
	return gsm.Accounts.Map[userID].PubKey()
}

/*********************************************************************************************************************
                                                    TX相关
*********************************************************************************************************************/

// NewTX 新建交易
func (gsm *GlobalStateMachine) NewTX(typ uint, args ArgsOfNewTX) (tx TX, err error) {
	// 检验参数。这里已经做了账户角色的验证
	if err = args.CheckArgsValue(gsm); err != nil {
		return nil, WrapError("newTX", err)
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