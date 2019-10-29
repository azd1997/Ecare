package ecoinlib

import (
	"bytes"
	"fmt"
	"github.com/dgraph-io/badger"
	"log"
)

type Chain struct {
	LastHash []byte
	Db       *badger.DB
}

func InitChain(userID UserID, port string, dbPathTemp string, genesisMsg string) (c *Chain, err error) {
	var lastHash []byte

	path := fmt.Sprintf(dbPathTemp, port)
	if DbExists(path) {
		// 如果数据库已经存在，在上一级的调用函数中应该处理这个错误并退出软件runtime.Goexit()
		return nil, fmt.Errorf("InitChain: %s", ErrChainAlreadyExists)
	}

	// 打开数据库
	opts := badger.DefaultOptions(path)
	db, err := openDB(opts)
	if err != nil {
		return nil, fmt.Errorf("InitChain: %s", err)
	}

	// 存入创世区块和lastHash
	err = db.Update(func(txn *badger.Txn) error {
		cbTx, err := CoinbaseTx(userID, genesisMsg)
		if err != nil {
			return err // 传回调用者，即db.Update
		}
		genesisBlock, err := GenesisBlock(cbTx, userID)
		if err != nil {
			return err
		}
		log.Println("Genesis created...")
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
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("InitChain: %s", err)
	}

	return &Chain{
		LastHash: lastHash,
		Db:       db, // 注意它存了一个badger.DB指针，其生命周期为程序运行时间
	}, nil
}

func ContinueChain(nodeId string, dbPathTemp string) (c *Chain, err error) {
	path := fmt.Sprintf(dbPathTemp, nodeId)
	if !DbExists(path) {
		return nil, fmt.Errorf("ContinueChain: %s", ErrChainNotExists)
	}

	var lastHash []byte

	opts := badger.DefaultOptions(path)
	db, err := openDB(opts)
	if err != nil {
		return nil, fmt.Errorf("ContinueChain: %s", err)
	}

	err = db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			// 避免直接赋值
			lastHash = append(lastHash, val...)
			return nil
		}) // 由于明确知道这里返回的err=nil，所以不处理
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("ContinueChain: %s", err)
	}

	return &Chain{
		LastHash: lastHash,
		Db:       db,
	}, nil
}

func (c *Chain) MineBlock(txs []*Transaction, createBy UserID, checksumLength int) (b *Block, err error) {
	// 验证交易有效性
	var valid bool
	for i, tx := range txs {
		if valid, _ = tx.Verify(checksumLength); !valid {
			return nil, fmt.Errorf("Chain_MineBlock: txs[%d]: %s", i, ErrInvalidTransaction)
		}
	}

	// 获取当前最新区块信息
	var lastHash []byte
	var lastId int
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
			var lastBlock *Block
			lastBlockBytes = append(lastBlockBytes, val...)
			if lastBlock, err = Deserialize(lastBlockBytes); err != nil {
				return err
			}
			lastId = lastBlock.Id
			return nil
		}) // err = nil
		return err
	}); err != nil {
		return nil, fmt.Errorf("Chain_MineBlock: %s", err)
	}

	// 构造新区块
	var newBlock *Block
	var newBlockBytes []byte
	newBlock = NewBlock(txs, lastHash, lastId+1, createBy)
	if newBlockBytes, err = newBlock.Serialize(); err != nil {
		return newBlock, fmt.Errorf("Chain_MineBlock: %s", err)
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
		return newBlock, fmt.Errorf("Chain_MineBlock: %s", err)
	}

	return newBlock, nil
}

// 从别的节点接收到区块后添加入本地的区块链
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
		// 下边直接存的原因是因为是不断同步的，所以如果区块链中没有这个区块，那么一定是紧接着的区块
		if err = txn.Set(b.Hash, blockBytes); err != nil {
			return err
		}

		// 获取原先最后一个区块
		var item *badger.Item
		var lastBlockBytes []byte
		var lastBlock *Block
		if item, err = txn.Get([]byte("lh")); err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			lastBlockBytes = append(lastBlockBytes, val...)
			return nil
		}) // err = nil
		if lastBlock, err = Deserialize(lastBlockBytes); err != nil {
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
		return fmt.Errorf("Chain_AddBlock: %s", err)
	}
	return nil
}

func (c *Chain) GetBlockByHash(blockHash []byte) (b *Block, err error) {
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

		b, err = Deserialize(blockBytes)
		return err
	}); err != nil {
		return nil, fmt.Errorf("Chain_GetBlockByHash: %s", err)
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
		return nil, fmt.Errorf("Chain_GetBlockById: %s", err)
	}

	// 检查id参数合理与否
	if id > maxId || id < (-maxId - 1) {
		return nil, fmt.Errorf("Chain_GetBlockById: %s", ErrOutOfChainRange)
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
	for i:=0; i<iterNum; i++ {
		if b, err = iter.Prev(); err != nil {
			return nil, fmt.Errorf("Chain_GetBlockById: %s", err)
		}
	}
	// 迭代完成后b指针就是所要的区块指针
	return b, nil
}

// 构造遍历用的迭代器
func (c *Chain) Iterator() *Iterator {
	return &Iterator{
		CurrentHash: c.LastHash,
		Db:          c.Db,
	}
}

// 获取所有区块的哈希集合，用于快速比较两个节点间区块链的一致性
func (c *Chain) GetBlockHashes() (blockHashes [][]byte, err error) {
	iter := c.Iterator()

	var block *Block
	for {
		if block, err = iter.Prev(); err != nil {
			// 因为是用来比较区块链一致性，所以如果中间出错了，那么比较结果没有意义，所以直接退出
			return nil, fmt.Errorf("Chain_GetBlockHashes: %s", err)
		}
		blockHashes = append(blockHashes, block.Hash)
		if len(block.PrevHash) == 0 {
			break
		}
	}
	return blockHashes, nil
}

// 正序打印区块链，所以可能需要栈（翻转输出），这里先直接用切片存储，因为懒得再去实现链表栈了
func (c *Chain) PrintBlockHeaders(start, end int) error {
	// start = 2,  end = 9

	// 检查输入参数
	if start > end {
		return fmt.Errorf("Chain_PrintBlockHeaders: %s", ErrWrongArguments)
	}

	// 检查end是否越界
	maxId, err := c.GetMaxId()
	if err != nil {
		return fmt.Errorf("Chain_PrintBlockHeaders: %s", err)
	}
	if maxId < end {
		return fmt.Errorf("Chain_PrintBlockHeaders: %s", ErrOutOfChainRange)
	}

	var currentBlock *Block
	iter := c.Iterator()	// iter.CurrentHash指向最后一个区块
	// 先遍历到end后一个区块
	for {
		if currentBlock, err = iter.Prev(); err != nil {
			return fmt.Errorf("Chain_PrintBlockHeaders: %s", err)
		}
		if currentBlock.Id == end + 1 || len(currentBlock.PrevHash) == 0 {
			break
		}
	}
	// 现在currentBlock就是end+1的区块，iter.CurrentHash指向end区块。

	// 继续向前遍历
	var resBlocks = make([]*Block, end - start + 1)
	for currentBlock.Id > start {
		// 从end区块开始
		if currentBlock, err = iter.Prev(); err != nil {
			return fmt.Errorf("Chain_PrintBlockHeaders: %s", err)
		}
		resBlocks = append(resBlocks, currentBlock)
	}

	// 倒序遍历resBlocks
	for i:=len(resBlocks); i>0; i-- {
		log.Println(resBlocks[i-1].String())
	}
	return nil
}

// 获取本地区块链中最大Id
func (c *Chain) GetMaxId() (maxId int, err error) {
	var lastBlock *Block

	if err = c.Db.View(func(txn *badger.Txn) error {
		var item *badger.Item
		var err error
		if item, err = txn.Get([]byte("lh")); err != nil {
			return err
		}

		var lastHash []byte
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
		if lastBlock, err = Deserialize(lastBlockBytes); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return -1, fmt.Errorf("Chain_GetMaxId: %s", err)
	}
	return lastBlock.Id, nil
}

func (c *Chain) FindTransaction(txId []byte) (tx *Transaction, err error) {
	iter := c.Iterator()

	var block *Block
	for {
		if block, err = iter.Prev(); err != nil {
			// 如果中间出错了，block可能就会空，后边就会空指针调用
			return nil, fmt.Errorf("Chain_FindTransaction: %s", err)
		}
		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, txId) == 0 {
				return tx, nil
			}
		}
		if len(block.PrevHash) == 0 {
			break
		}
	}
	return nil, ErrTransactionNotExists
}

type Iterator struct {
	CurrentHash []byte
	Db          *badger.DB
}

// 返回当前区块哈希对应的区块，并将哈希指针前移
func (iter *Iterator) Prev() (b *Block, err error) {
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
		if b, err = Deserialize(blockBytes); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("Iterator_Prev: %s", err)
	}

	// 更新current
	iter.CurrentHash = b.PrevHash
	return b, nil
}
