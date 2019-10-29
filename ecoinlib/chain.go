package ecoin

import (
	"encoding/base64"
	"fmt"
	"strconv"

	"github.com/azd1997/Ecare/ecoinlib/log"
	"github.com/dgraph-io/badger"
)

// InitChainArgs 初始化区块链的传入参数
type InitChainArgs struct {
	coinbase *TxCoinbaseArgs
	opts *Option
}

// ContinueChainArgs 继续区块链参数
type ContinueChainArgs struct {
	opts *Option
}

// Chain 区块链
type Chain struct {
	LastHash Hash
	Db       *badger.DB
}

// TODO: 思考一下，newTxCoinbase时传入gsm，但要注意只用到了gsm.opts。也就是此时gsm.ledger是可以为空的

// InitChain 初始化区块链. port指节点客户端监听端口，这么做是为了规范数据库路径命名
func InitChain(args *InitChainArgs) (c *Chain, err error) {
	var lastHash Hash

	path := fmt.Sprintf(CHAIN_DBPATH_TEMP, strconv.Itoa(int(args.opts.Port())))
	if DbExists(path) {
		// 如果数据库已经存在，在上一级的调用函数中应该处理这个错误并退出软件runtime.Goexit()
		return nil, WrapError("InitChain", ErrChainAlreadyExists)
	}

	// 打开数据库
	opts := badger.DefaultOptions(path)
	db, err := openDB(opts)
	if err != nil {
		return nil, WrapError("InitChain", err)
	}

	// 存入创世区块和lastHash
	err = db.Update(func(txn *badger.Txn) error {
		cbTx, err := newTxCoinbase(args.coinbase)
		if err != nil {
			return err // 传回调用者，即db.Update
		}
		genesisBlock, err := GenesisBlock(cbTx)
		if err != nil {
			return err
		}
		log.Success("Genesis created: %s", base64.StdEncoding.EncodeToString(genesisBlock.Hash[:]))
		genesisBlockBytes, err := genesisBlock.Serialize()
		if err != nil {
			return err
		}
		err = txn.Set(genesisBlock.Hash, genesisBlockBytes)
		if err != nil {
			return err
		}
		err = txn.Set([]byte("lh"), genesisBlock.Hash)
		if err != nil {
			return err
		}
		// 更新lastHash
		lastHash = genesisBlock.Hash
		return nil
	})
	if err != nil {
		return nil, WrapError("InitChain", err)
	}

	return &Chain{
		LastHash: lastHash,
		Db:       db, // 注意它存了一个badger.DB指针，其生命周期为程序运行时间
	}, nil
}

// ContinueChain 从数据库继续区块链
func ContinueChain(args *ContinueChainArgs) (c *Chain, err error) {
	path := fmt.Sprintf(CHAIN_DBPATH_TEMP, strconv.Itoa(int(args.opts.Port())))
	if !DbExists(path) {
		return nil, WrapError("ContinueChain", ErrChainNotExists)
	}

	var lastHash Hash

	// 打开数据库
	opts := badger.DefaultOptions(path)
	db, err := openDB(opts)
	if err != nil {
		return nil, WrapError("ContinueChain", err)
	}

	err = db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		if err != nil {
			return err
		}
		_ = item.Value(func(val []byte) error {
			// 避免直接赋值
			lastHash = append(lastHash, val...)
			return nil
		}) // 由于明确知道这里返回的err=nil，所以不处理
		return err
	})
	if err != nil {
		return nil, WrapError("ContinueChain", err)
	}

	return &Chain{
		LastHash: lastHash,
		Db:       db,
	}, nil
}

// NewEmptyChain 构造一个没有任何区块的区块链
func NewEmptyChain(dbPath string) (c *Chain, err error) {

	// 检查路径是否已经有区块链数据库，谨防不小心清掉了原本的数据
	if DbExists(dbPath) {
		return nil, WrapError("NewEmptyChain", ErrChainAlreadyExists)
	}

	// 打开数据库
	badgerOpts := badger.DefaultOptions(dbPath)
	db, err := openDB(badgerOpts)
	if err != nil {
		return nil, err
	}

	// 更新“lh”键对应值为
	// 存入lastHash
	err = db.Update(func(txn *badger.Txn) error {
		err = txn.Set([]byte("lh"), ZeroHASH)
		return err
	})
	if err != nil {
		return nil, WrapError("NewEmptyChain", ErrChainAlreadyExists)
	}

	return &Chain{
		LastHash: ZeroHASH,
		Db:       db,
	}, nil // TODO：notice : 这里要修改，Db是需要创建的，根据opts
}

// MineBlock 有Chain的前提下进行挖块就是直接打包区块。 在这里传入的txs需要是全部交易列表，包括coinbase交易
func (c *Chain) MineBlock(txs []TX, gsm *GlobalStateMachine) (b *Block, err error) {
	// 验证交易有效性
	for i, tx := range txs {
		if err = tx.IsValid(gsm); err != nil {
			return nil, fmt.Errorf("Chain_MineBlock: txs[%d]: %s", i, ErrInvalidTransaction)
		}
	}

	// 获取当前最新区块信息
	var lastHash Hash
	var lastId uint
	if err = c.Db.View(func(txn *badger.Txn) error {
		var item *badger.Item
		var err error
		if item, err = txn.Get([]byte("lh")); err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			lastHash = append(lastHash, val...)
			return nil
		}) // err = nil
		if item, err = txn.Get(lastHash); err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			var lastBlockBytes []byte
			lastBlock := &Block{}
			lastBlockBytes = append(lastBlockBytes, val...)
			if err = lastBlock.Deserialize(lastBlockBytes); err != nil {
				return err
			}
			lastId = lastBlock.Id
			return nil
		}) // err = nil
		return err
	}); err != nil {
		return nil, WrapError("Chain_MineBlock", err)
	}

	// 构造新区块
	var newBlock *Block
	var newBlockBytes []byte
	newBlock = NewBlock(txs, lastHash, lastId+1, *gsm.opts.UserID())
	if newBlockBytes, err = newBlock.Serialize(); err != nil {
		return newBlock, WrapError("Chain_MineBlock", err)
	}

	// 存入数据库
	if err = c.Db.Update(func(txn *badger.Txn) error {
		var err error
		if err = txn.Set(newBlock.Hash, newBlockBytes); err != nil {
			return err
		}
		if err = txn.Set([]byte("lh"), newBlock.Hash); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return newBlock, WrapError("Chain_MineBlock", err)
	}

	return newBlock, nil
}

// AddBlock 从别的节点接收到区块后添加入本地的区块链
func (c *Chain) AddBlock(b *Block) (err error) {
	if err = c.Db.Update(func(txn *badger.Txn) error {
		var err error
		// 查看区块链中是否已有该区块
		if _, err = txn.Get(b.Hash); err == nil {
			return ErrBlockAlreadyExists
		}
		// 存入区块
		var blockBytes []byte
		if blockBytes, err = b.Serialize(); err != nil {
			return err
		}
		// TODO: 下边直接存的原因是因为是不断同步的，所以如果区块链中没有这个区块，那么一定是紧接着的区块
		if err = txn.Set(b.Hash, blockBytes); err != nil {
			return err
		}

		// 获取原先最后一个区块
		var item *badger.Item
		var lastBlockBytes []byte
		lastBlock := &Block{}
		if item, err = txn.Get(c.LastHash); err != nil {
			return err
		}
		_ = item.Value(func(val []byte) error {
			lastBlockBytes = append(lastBlockBytes, val...)
			return nil
		}) // err = nil
		if err = lastBlock.Deserialize(lastBlockBytes); err != nil {
			return err
		}

		// 如果新区块刚好紧跟现有最新区块，则修改lastHash
		if b.Id == lastBlock.Id+1 {
			if err = txn.Set([]byte("lh"), b.Hash); err != nil {
				return err
			}
			c.LastHash = b.Hash
		}

		// 如果因为长期掉线导致新区块是本地区块链最新区块的2个及往后，暂时不更新lastHash直至找到最新
		// 例如本地区块链已有区块 0 1 2， 但最新状态已经是 0 1 2 3 4 5
		// 加入在其他人给我发的区块因为某些原因顺序错乱了，我第一个接收的是区块4
		// TODO:有点复杂

		return nil
	}); err != nil {
		return WrapError("Chain_AddBlock", err)
	}
	return nil
}

// GetBlockByHash 根据区块哈希获取区块
func (c *Chain) GetBlockByHash(blockHash Hash) (b *Block, err error) {
	if err = c.Db.View(func(txn *badger.Txn) error {
		var err error
		var item *badger.Item
		var blockBytes []byte
		if item, err = txn.Get(blockHash); err != nil {
			return ErrBlockNotExists // 这里将所有错误笼统的概括为不存在，因为其他细致的错误我们不需要
		}
		_ = item.Value(func(val []byte) error {
			blockBytes = append(blockBytes, val...)
			return nil
		}) // err = nil

		b = &Block{}
		err = b.Deserialize(blockBytes)
		return err
	}); err != nil {
		return nil, WrapError("Chain_GetBlockByHash", err)
	}

	return b, nil
}

func (c *Chain) GetBlockById(id int) (b *Block, err error) {
	// id允许负数索引
	// 0,1,2,...,maxId
	// -maxId-1, ..., -2, -1

	// 先获取最大
	maxId, err := c.GetMaxId()
	if err != nil {
		return nil, WrapError("Chain_GetBlockById", err)
	}

	// 检查id参数合理与否
	if id > maxId || id < (-maxId-1) {
		return nil, WrapError("Chain_GetBlockById", ErrOutOfChainRange)
	}

	// ex 0 1 2 3 4 5
	// input 3  需迭代5-3 + 1 = 3次
	// input -3 需迭代 -（-3）= 3次
	var iterNum int
	if id >= 0 {
		// 正向索引
		iterNum = maxId - id + 1
	} else {
		// 负数索引
		iterNum = -(id)
	}

	var iter = c.Iterator()
	for i := 0; i < iterNum; i++ {
		if b, err = iter.Prev(); err != nil {
			return nil, WrapError("Chain_GetBlockById", err)
		}
	}
	// 迭代完成后b指针就是所要的区块指针
	return b, nil
}

// Iterator 构造遍历用的迭代器
func (c *Chain) Iterator() *Iterator {
	return &Iterator{
		CurrentHash: c.LastHash,
		Db:          c.Db,
	}
}

// GetBlockHashes 获取所有区块的哈希集合(从后到前)，用于快速比较两个节点间区块链的一致性
func (c *Chain) GetBlockHashes() (blockHashes []Hash, err error) {
	iter := c.Iterator()

	var block *Block
	for {
		if block, err = iter.Prev(); err != nil {
			// 因为是用来比较区块链一致性，所以如果中间出错了，那么比较结果没有意义，所以直接退出
			return nil, WrapError("Chain_GetBlockHashes", err)
		}
		blockHashes = append(blockHashes, block.Hash)
		if string(block.PrevHash) == string(Hash{}) {
			break
		}
	}
	return blockHashes, nil
}

// PrintBlockHeaders 正序打印区块链，所以可能需要栈（翻转输出），这里先直接用切片存储，因为懒得再去实现链表栈了
// TODO:待修正
func (c *Chain) PrintBlockHeaders(start, end uint) error {
	// start = 2,  end = 9

	// 检查输入参数
	if start > end {
		return WrapError("Chain_PrintBlockHeaders", ErrWrongArguments)
	}

	// 检查end是否越界
	maxId, err := c.GetMaxId()
	if err != nil {
		return WrapError("Chain_PrintBlockHeaders", err)
	}
	if maxId < int(end) {
		return WrapError("Chain_PrintBlockHeaders", ErrOutOfChainRange)
	}

	var currentBlock *Block
	iter := c.Iterator() // iter.CurrentHash指向最后一个区块
	// 先遍历到end后一个区块
	for {
		if currentBlock, err = iter.Prev(); err != nil {
			return WrapError("Chain_PrintBlockHeaders", err)
		}
		if currentBlock.Id == end+1 || string(currentBlock.PrevHash) == string(Hash{}) {
			break
		}
	}
	// 现在currentBlock就是end+1的区块，iter.CurrentHash指向end区块。

	// 继续向前遍历
	var resBlocks = make([]*Block, end-start+1)
	for currentBlock.Id > start {
		// 从end区块开始
		if currentBlock, err = iter.Prev(); err != nil {
			return WrapError("Chain_PrintBlockHeaders", err)
		}
		resBlocks = append(resBlocks, currentBlock)
	}

	// 倒序遍历resBlocks
	for i := len(resBlocks); i > 0; i-- {
		log.Info("%s", resBlocks[i-1].String())
	}
	return nil
}

// GetMaxId 获取本地区块链中最大Id
func (c *Chain) GetMaxId() (maxId int, err error) {
	lastBlock := &Block{}

	if err = c.Db.View(func(txn *badger.Txn) error {
		var item *badger.Item
		var err error
		if item, err = txn.Get([]byte("lh")); err != nil {
			return err
		}

		var lastHash Hash
		_ = item.Value(func(val []byte) error {
			lastHash = append(lastHash, val...)
			return nil
		})

		if item, err = txn.Get(lastHash); err != nil {
			return err
		}

		var lastBlockBytes []byte

		_ = item.Value(func(val []byte) error {
			lastBlockBytes = append(lastBlockBytes, val...)
			return nil
		})
		if err = lastBlock.Deserialize(lastBlockBytes); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return -1, WrapError("Chain_GetMaxId", err)
	}
	return int(lastBlock.Id), nil
}

// FindTransaction
func (c *Chain) FindTransaction(txId Hash) (tx TX, err error) {
	iter := c.Iterator()

	var block *Block
	var txHash Hash
	for {
		if block, err = iter.Prev(); err != nil {
			// 如果中间出错了，block可能就会空，后边就会空指针调用
			return nil, WrapError("Chain_FindTransaction", err)
		}
		for i, txBytes := range block.Transactions {
			tx, err = DeserializeTX(block.TxTypes[i], txBytes)
			if err != nil {
				continue
			}
			txHash, err = tx.Hash()
			if err != nil {
				continue	// 跳过
			}
			if string(txHash) == string(txId) {
				return tx, nil
			}
		}
		if string(block.PrevHash) == string(Hash{}) {		// TODO： 理解： 这里为什么 不能直接用Hash{} 因为编译器不认识，会将其认为是判断后的执行程序块
			break
		}
	}
	return nil, ErrTransactionNotExists
}

// Iterator 区块迭代器
type Iterator struct {
	CurrentHash Hash
	Db          *badger.DB
}

// Prev 返回当前区块哈希对应的区块，并将哈希指针前移
func (iter *Iterator) Prev() (b *Block, err error) {

	b = &Block{}

	if err = iter.Db.View(func(txn *badger.Txn) error {
		var item *badger.Item
		var err error
		if item, err = txn.Get(iter.CurrentHash); err != nil {
			return err
		}
		var blockBytes []byte
		_ = item.Value(func(val []byte) error {
			blockBytes = append(blockBytes, val...)
			return nil
		})
		if err = b.Deserialize(blockBytes); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, WrapError("Iterator_Prev", err)
	}

	// 更新current
	iter.CurrentHash = b.PrevHash
	return b, nil
}

// HasBlock 区块是否存在本地区块链中
func (c *Chain) HasBlock(b *Block) (bool, error) {
	localBlockHashes, err := c.GetBlockHashes()
	if err != nil {
		return false, WrapError("HasBlock", err)
	}
	for _, localBlockHash := range localBlockHashes {
		if string(localBlockHash) == string(b.Hash) {
			return true, nil
		}
	}
	return false, nil
}

//// 检查交易是否合理有效
//func (c *Chain) VerifyTx(tx *Transaction) (bool, error) {
//	// valid, err :=
//	// TODO: 修改EcoinWorld相关代码
//	return false, nil
//}

// 检查区块是否合法
func (c *Chain) VerifyBlock(b *Block) (bool, error) {
	// TODO
	return false, nil
}
