package singlechain

import "C"
import (
	"encoding/base64"
	"fmt"

	"github.com/azd1997/Ecare/ecoin/account"
	"github.com/azd1997/Ecare/ecoin/crypto"
	"github.com/azd1997/Ecare/ecoin/database"
	"github.com/azd1997/Ecare/ecoin/log"
	"github.com/azd1997/Ecare/ecoin/transaction"
	"github.com/azd1997/Ecare/ecoin/utils"
)

// InitChainArgs 初始化区块链的传入参数
type InitChainArgs struct {
	DbPath        string
	CoinbaseArgs  *transaction.CoinbaseArgs
	CheckArgsFunc transaction.CheckArgsFunc
}

// ContinueChainArgs 继续区块链参数
type ContinueChainArgs struct {
	DbPath string
}

// Chain 区块链
type Chain struct {
	LastHash crypto.Hash
	Db       database.Database
}

// TODO: 思考一下，newTxCoinbase时传入gsm，但要注意只用到了gsm.opts。也就是此时gsm.ledger是可以为空的

// InitChain 初始化区块链. port指节点客户端监听端口，这么做是为了规范数据库路径命名
func InitChain(args *InitChainArgs) (c *Chain, err error) {

	// 确保数据库存在
	if database.DbExists(DBEngine, args.DbPath) {
		// 如果数据库已经存在，在上一级的调用函数中应该处理这个错误并退出软件runtime.Goexit()
		return nil, utils.WrapError("InitChain", ErrChainAlreadyExists)
	}

	// 打开数据库
	db, err := database.OpenDatabaseWithRetry(DBEngine, args.DbPath)
	if err != nil {
		return nil, utils.WrapError("InitChain", err)
	}

	// 构建创世区块
	cbTx, err := transaction.NewTXWithArgsCheck(transaction.TX_COINBASE, args.CoinbaseArgs, args.CheckArgsFunc)
	if err != nil {
		return nil, utils.WrapError("InitChain", err)
	}
	genesisBlock, err := GenesisBlock(cbTx.(*transaction.TxCoinbase))
	if err != nil {
		return nil, utils.WrapError("InitChain", err)
	}
	log.Success("Genesis created: %s", base64.StdEncoding.EncodeToString(genesisBlock.Hash[:]))
	genesisBlockBytes, err := genesisBlock.Serialize()
	if err != nil {
		return nil, utils.WrapError("InitChain", err)
	}

	// 存入创世区块和lastHash
	if err = db.Set(genesisBlock.Hash, genesisBlockBytes); err != nil {
		return nil, utils.WrapError("InitChain", err)
	}
	if err = db.Set(LastHashKey, genesisBlock.Hash); err != nil {
		return nil, utils.WrapError("InitChain", err)
	}

	return &Chain{
		LastHash: genesisBlock.Hash,
		Db:       db, // 注意它存了一个badger.DB指针，其生命周期为程序运行时间
	}, nil
}

// ContinueChain 从数据库继续区块链
func ContinueChain(args *ContinueChainArgs) (c *Chain, err error) {

	// 检查数据库是否存在
	if !database.DbExists(DBEngine, args.DbPath) {
		return nil, utils.WrapError("ContinueChain", ErrChainNotExists)
	}

	// 打开数据库
	db, err := database.OpenDatabaseWithRetry(DBEngine, args.DbPath)
	if err != nil {
		return nil, utils.WrapError("ContinueChain", err)
	}

	// 获取lastHash
	lastHash, err := db.Get(LastHashKey)
	if err != nil {
		return nil, utils.WrapError("ContinueChain", err)
	}

	return &Chain{
		LastHash: lastHash,
		Db:       db,
	}, nil
}

// NewEmptyChain 构造一个没有任何区块的区块链
func NewEmptyChain(dbPath string) (c *Chain, err error) {

	// 检查路径是否已经有区块链数据库，谨防不小心清掉了原本的数据
	if database.DbExists(DBEngine, dbPath) {
		return nil, utils.WrapError("NewEmptyChain", ErrChainAlreadyExists)
	}

	// 打开数据库
	db, err := database.OpenDatabaseWithRetry(DBEngine, dbPath)
	if err != nil {
		return nil, utils.WrapError("NewEmptyChain", err)
	}

	// 更新“lh”键对应值为
	// 存入lastHash
	if err = db.Set(LastHashKey, crypto.ZeroHASH); err != nil {
		return nil, utils.WrapError("NewEmptyChain", err)
	}

	return &Chain{
		LastHash: crypto.ZeroHASH,
		Db:       db,
	}, nil
}

// MineBlock 有Chain的前提下进行挖块就是直接打包区块。 在这里传入的txs需要是全部交易列表，包括coinbase交易
func (c *Chain) MineBlock(txs []transaction.TX, txFunc transaction.ValidateTxFunc, userId account.UserId) (b *Block, err error) {
	// 验证交易有效性
	for i, tx := range txs {
		if err = tx.IsValid(txFunc); err != nil {
			return nil, fmt.Errorf("Chain_MineBlock: txs[%d]: %s", i, ErrInvalidTX)
		}
	}

	// 获取lastBlock
	lastBlockBytes, err := c.Db.Get(c.LastHash)
	if err != nil {
		return nil, utils.WrapError("Chain_MineBlock", err)
	}
	lastBlock := &Block{}
	if err = lastBlock.Deserialize(lastBlockBytes); err != nil {
		return nil, utils.WrapError("Chain_MineBlock", err)
	}

	// TODO: 此处： 按照PoT需要进行竞争才能继续下去； 按POW需要计算哈希难题
	// 都可以通过一个管道传信号过来，通知继续下去

	// 构造新区块
	var newBlock *Block
	var newBlockBytes []byte
	newBlock = NewBlock(txs, c.LastHash, lastBlock.Id+1, userId)
	if newBlockBytes, err = newBlock.Serialize(); err != nil {
		return newBlock, utils.WrapError("Chain_MineBlock", err)
	}

	// 存入数据库
	if err = c.Db.Set(newBlock.Hash, newBlockBytes); err != nil {
		return nil, utils.WrapError("Chain_MineBlock", err)
	}
	if err = c.Db.Set(LastHashKey, newBlock.Hash); err != nil {
		return nil, utils.WrapError("Chain_MineBlock", err)
	}

	return newBlock, nil
}

// AddBlock 从别的节点接收到区块检验合格后添加入本地的区块链
func (c *Chain) AddBlock(b *Block) (err error) {

	// 查看区块链中是否已有该区块
	if _, err = c.Db.Get(b.Hash); err == nil {
		return utils.WrapError("Chain_AddBlock", ErrBlockAlreadyExists)
	}

	// TODO： 传入的b是全方位检查过合格的，所以，这里不检查，认为它就是下一个区块

	// 存入区块
	var blockBytes []byte
	if blockBytes, err = b.Serialize(); err != nil {
		return utils.WrapError("Chain_AddBlock", err)
	}
	if err = c.Db.Set(b.Hash, blockBytes); err != nil {
		return utils.WrapError("Chain_AddBlock", err)
	}
	if err = c.Db.Set(LastHashKey, b.Hash); err != nil {
		return utils.WrapError("Chain_AddBlock", err)
	}

	// 更新c.LastHash
	c.LastHash = b.Hash

	// TODO： 特殊情况是，久不上线节点请求增区块

	return nil
}

// GetBlockByHash 根据区块哈希获取区块
func (c *Chain) GetBlockByHash(blockHash crypto.Hash) (b *Block, err error) {

	blockBytes, err := c.Db.Get(blockHash)
	if err != nil {
		return nil, utils.WrapError("Chain_GetBlockByHash", ErrBlockNotExists) // 这里将所有错误笼统的概括为不存在，因为其他细致的错误我们不需要
	}
	b = &Block{}
	if err = b.Deserialize(blockBytes); err != nil {
		return nil, utils.WrapError("Chain_GetBlockByHash", err)
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
		return nil, utils.WrapError("Chain_GetBlockById", err)
	}

	// 检查id参数合理与否
	if id > maxId || id < (-maxId-1) {
		return nil, utils.WrapError("Chain_GetBlockById", ErrIdOutOfChainRange)
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
			return nil, utils.WrapError("Chain_GetBlockById", err)
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
func (c *Chain) GetBlockHashes() (blockHashes []crypto.Hash, err error) {
	iter := c.Iterator()

	var block *Block
	for {
		if block, err = iter.Prev(); err != nil {
			// 因为是用来比较区块链一致性，所以如果中间出错了，那么比较结果没有意义，所以直接退出
			return nil, utils.WrapError("Chain_GetBlockHashes", err)
		}
		blockHashes = append(blockHashes, block.Hash)
		if string(block.PrevHash) == string(crypto.ZeroHASH) {
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
		return utils.WrapError("Chain_PrintBlockHeaders", ErrWrongArguments)
	}

	// 检查end是否越界
	maxId, err := c.GetMaxId()
	if err != nil {
		return utils.WrapError("Chain_PrintBlockHeaders", err)
	}
	if maxId < int(end) {
		return utils.WrapError("Chain_PrintBlockHeaders", ErrIdOutOfChainRange)
	}

	var currentBlock *Block
	iter := c.Iterator() // iter.CurrentHash指向最后一个区块
	// 先遍历到end后一个区块
	for {
		if currentBlock, err = iter.Prev(); err != nil {
			return utils.WrapError("Chain_PrintBlockHeaders", err)
		}
		if currentBlock.Id == end+1 || string(currentBlock.PrevHash) == string(crypto.ZeroHASH) {
			break
		}
	}
	// 现在currentBlock就是end+1的区块，iter.CurrentHash指向end区块。

	// 继续向前遍历
	var resBlocks = make([]*Block, end-start+1)
	for currentBlock.Id > start {
		// 从end区块开始
		if currentBlock, err = iter.Prev(); err != nil {
			return utils.WrapError("Chain_PrintBlockHeaders", err)
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

	lastBlockBytes, err := c.Db.Get(c.LastHash)
	if err != nil {
		return -1, utils.WrapError("Chain_GetMaxId", err)
	}

	lastBlock := &Block{}
	if err = lastBlock.Deserialize(lastBlockBytes); err != nil {
		return -1, utils.WrapError("Chain_GetMaxId", err)
	}

	return int(lastBlock.Id), nil
}

// FindTransaction
func (c *Chain) FindTransaction(txId crypto.Hash) (tx transaction.TX, err error) {
	iter := c.Iterator()

	var block *Block
	var txHash crypto.Hash
	for {
		if block, err = iter.Prev(); err != nil {
			// 如果中间出错了，block可能就会空，后边就会空指针调用
			return nil, utils.WrapError("Chain_FindTransaction", err)
		}
		for i, txBytes := range block.Transactions {
			tx, err = transaction.DeserializeTX(block.TxTypes[i], txBytes)
			if err != nil {
				continue
			}
			txHash, err = tx.Hash()
			if err != nil {
				continue // 跳过
			}
			if string(txHash) == string(txId) {
				return tx, nil
			}
		}
		if string(block.PrevHash) == string(crypto.ZeroHASH) {
			break
		}
	}
	return nil, ErrTransactionNotExists
}

// Iterator 区块迭代器
type Iterator struct {
	CurrentHash crypto.Hash
	Db          database.Database
}

// Prev 返回当前区块哈希对应的区块，并将哈希指针前移
func (iter *Iterator) Prev() (b *Block, err error) {



	blockBytes, err := iter.Db.Get(iter.CurrentHash)
	if err != nil {
		return nil, utils.WrapError("Iterator_Prev", err)
	}

	b = &Block{}
	if err = b.Deserialize(blockBytes); err != nil {
		return nil, utils.WrapError("Iterator_Prev", err)
	}

	// 更新current
	iter.CurrentHash = b.PrevHash
	return b, nil
}

// HasBlock 区块是否存在本地区块链中
func (c *Chain) HasBlock(b *Block) (bool, error) {
	localBlockHashes, err := c.GetBlockHashes()
	if err != nil {
		return false, utils.WrapError("HasBlock", err)
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
