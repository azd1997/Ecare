package types

import (
	"bytes"
	"encoding/gob"
	"github.com/azd1997/Ecare/common/ecoinlib/log"
	"github.com/azd1997/Ecare/common/ecoinlib/utils"
	"time"
)

// BLockHeader 区块头
type BlockHeader struct {
	Id       uint // 从0开始，很多称为height
	Time     UnixTimeStamp
	PrevHash Hash
	Hash     Hash   // 当前区块哈希，实际是区块体内交易列表组成的MerkleTree的根哈希
	CreateBy UserID // 由哪个账户创建
}

// BlockBody 区块体内容为交易列表
type BlockBody struct {
	Transactions []TX
}

// Block 区块
type Block struct {
	BlockHeader
	BlockBody
}

// BlockForPrint 用于打印的区块结构
type BlockForPrint struct {
	Id       uint    `json:"id"` // 从0开始，很多称为height
	Time     string `json:"time"`
	PrevHash Hash   `json:"prevHash, omitempty"` // 对于创世区块，没有prevHash，零值是nil则会在json转换时忽略该项
	Hash     Hash   `json:"hash"`                // 当前区块哈希，实际是区块体内交易列表组成的MerkleTree的根哈希
	CreateBy string `json:"createBy"`            // 由哪个账户创建
	Txs      []Hash `json:"txs"`
}

// String 转换字符串（DEPRECATED）
//func (b *Block) String() string {
//	return fmt.Sprintf(
//		`{
//	id: 		%d
//	time:       %s
//	prevHash: 	%s
//	hash: 		%s
//	createBy: 	%s
//}`,
//		b.Id,
//		time.Unix(int64(b.Time), 0).Format("2006/01/02 15:04:05"),
//		b.PrevHash,
//		b.Hash,
//		b.CreateBy.Id)
//}

// String 转换字符串
func (b *Block) String() string {
	var txs []Hash
	for _, tx := range b.Transactions {
		txId, _ := tx.Hash() // 因为这只是用于打印展示，所以出错了也不管，跳过
		txs = append(txs, txId)
	}
	blockForPrint := BlockForPrint{
		Id:       b.Id,
		Time:     time.Unix(int64(b.Time), 0).Format("2006/01/02 15:04:05"),
		PrevHash: b.PrevHash,
		Hash:     b.Hash,
		CreateBy: b.CreateBy.ID,
		Txs:      txs,
	}
	return utils.JsonMarshalIndent(&blockForPrint)
}

// NewBlock 新建区块。在这里传入的txs是全部交易列表，包括自己的coinbase交易
func NewBlock(txs []TX, prevHash Hash, id uint, createBy UserID) *Block {
	l := len(txs)
	var txsBytes = make([][]byte, l)
	var err error
	for i, tx := range txs {
		txsBytes[i], err = tx.Serialize()
		if err != nil {
			// 记录下错误并继续
			log.Error("NewBlock: %s\n", err)
			continue
		}
	}
	// merkle.RootNode.Data 根哈希值是一个[]byte且长度为32。只是在其实现处不方便将其改为Hash类型，所以在这里要手动转一下
	merkle := NewMerkleTree(txsBytes)
	merkleRoot := BytesToHash(merkle.RootNode.Data)
	// 注意，这里是确定根哈希是32位的才这么直接赋值过来，不要随便这么写，很容易出问题

	return &Block{
		BlockHeader: BlockHeader{
			Id:       id,
			Time:     UnixTimeStamp(time.Now().Unix()),
			PrevHash: prevHash,
			Hash:     merkleRoot,
			CreateBy: createBy,
		},
		BlockBody: BlockBody{
			Transactions: txs,
		},
	}
}

// GenesisBlock 创世区块。传入创世区块中的coinbase交易
func GenesisBlock(coinbase *TxCoinbase) (gb *Block, err error) {

	// 创世区块就一个交易懒得构造merkle
	coinbaseTxBytes, err := coinbase.Serialize()
	if err != nil {
		return nil, utils.WrapError("GenesisBlock", err)
	}
	merkle := NewMerkleTree([][]byte{coinbaseTxBytes})
	merkleRoot := BytesToHash(merkle.RootNode.Data)

	return &Block{
		BlockHeader: BlockHeader{
			Id:       0,
			Time:     UnixTimeStamp(time.Now().Unix()),
			PrevHash:Hash{},
			Hash:     merkleRoot,
			CreateBy: coinbase.To,
		},
		BlockBody: BlockBody{
			Transactions: []TX{coinbase},
		},
	}, nil
}

// Serialize 序列化为字节切片
func (b *Block) Serialize() (res []byte, err error) {
	return utils.GobEncode(b)
}

// Deserialize 反序列化到给定零值区块 b := &Block{}, b.Deserialize(blockBytes)
func (b *Block) Deserialize(blockBytes []byte) (err error) {
	// 防止非空TxR2P调用该方法改变了自身内容
	b1 := &Block{}
	if b != b1 {
		return utils.WrapError("Block_Deserialize", ErrDeserializeRequireEmptyReceiver)
	}
	// gob解码
	decoder := gob.NewDecoder(bytes.NewReader(blockBytes))
	if err = decoder.Decode(b); err != nil {
		return utils.WrapError("Block_Deserialize", err)
	}
	return nil
}

// PrintTransactions 打印区块内交易列表
func (b *Block) PrintTransactions() {
	for _, tx := range b.Transactions {
		log.Info("%s", tx.String())
	}
}

// VerifyTXs 验证区块内所有交易。收到区块检查时调用
func (b *Block) VerifyTXs(gsm *GlobalStateMachine) (txTypeNoArray [][]uint, err error) {

	// txTypeNoArray [][]uint指的是这个区块内所有交易各种类型的交易其在区块交易列表中的索引

	/*txTypes := []TX{
		&TxCoinbase{},
		&TxGeneral{},
		&TxR2P{},
		&TxP2R{},
		&TxP2H{},
		&TxH2P{},
		&TxP2D{},
		&TxD2P{},
		&TxArbitrate{},
	}*/

	txTypeNoArray = [][]uint{}

	// 1.检查coinbase交易和区块制造者是否吻合
	if b.CreateBy != b.Transactions[0].(*TxCoinbase).To {
		return nil, ErrInvalidCoinbaseTX
	}

	// 使用TX.IsValid检查所有交易
	for i, tx := range b.Transactions {
		if err = tx.IsValid(gsm); err != nil {
			return nil, ErrBlockContainsInvalidTX
		}
		switch tx.TypeNo() {
		case TX_COINBASE:	// TX_COINBASE = 1
			txTypeNoArray[0] = append(txTypeNoArray[0], uint(i))
		case TX_GENERAL:	// 2
			txTypeNoArray[1] = append(txTypeNoArray[1], uint(i))
		case TX_R2P:	// 3
			txTypeNoArray[2] = append(txTypeNoArray[2], uint(i))
		case TX_P2R:	// 4
			txTypeNoArray[3] = append(txTypeNoArray[3], uint(i))
		case TX_P2H:	// 5
			txTypeNoArray[4] = append(txTypeNoArray[4], uint(i))
		case TX_H2P:	// 6
			txTypeNoArray[5] = append(txTypeNoArray[5], uint(i))
		case TX_P2D:	// 7
			txTypeNoArray[6] = append(txTypeNoArray[6], uint(i))
		case TX_D2P:	// 8
			txTypeNoArray[7] = append(txTypeNoArray[7], uint(i))
		case TX_ARBITRATE:	// 9
			txTypeNoArray[8] = append(txTypeNoArray[8], uint(i))
			// 仲裁交易还需要检查仲裁者是否是出块者
			if b.CreateBy != tx.(*TxArbitrate).Arbitrator {
				return nil, ErrInvalidArbitrateTX
			}
		}
	}

	// 验证通过，并且返回各类交易在区块交易列表索引情况
	return txTypeNoArray, nil
}

// IsValid 区块是否有效。这仅用于gsm.ledger不为空的情况，调用之前需要判断这个情况。
func (b *Block) IsValid(gsm *GlobalStateMachine) (err error) {

	/*&Block{
		BlockHeader: BlockHeader{
			Id:       id,
			Time:     UnixTimeStamp(time.Now().Unix()),
			PrevHash: prevHash,
			Hash:     merkleRoot,
			CreateBy: createBy,
		},
		BlockBody: BlockBody{
			Transactions: txs,
		},
	}*/

	// 区块的检验一定是发生在出块者发出区块后其他节点收到来进行检验。传入的gsm也就是检验者本地的gsm
	// 检验区块的基本格式内容
	// 1. 检查时间
	if b.Time >= UnixTimeStamp(time.Now().Unix()) {
		return utils.WrapError("Block_IsValid", ErrWrongTimeBlock)
	}
	// 2. 检查Id和PrevHash值
	localLatestBlock, err := gsm.Ledger.GetBlockByHash(gsm.Ledger.LastHash)
	if err != nil {
		return utils.WrapError("Block_IsValid", err)
	}
	if b.PrevHash != localLatestBlock.Hash || b.Id != localLatestBlock.Id + 1 {
		return utils.WrapError("Block_IsValid", ErrNotNextBlock)
	}
	// 3. 检查出块者权限
	if b.CreateBy.RoleNo >= 10 || !gsm.Accounts.Map[b.CreateBy.ID].Available() {
		return utils.WrapError("Block_IsValid", ErrWrongRoleUserID)
	}
	// 4. 检查所有交易
	if _, err = b.VerifyTXs(gsm); err != nil {
		return utils.WrapError("Block_IsValid", ErrBlockContainsInvalidTX)
	}
	// 5. 检查区块哈希也就是交易默克树根哈希
	var txsBytes [][]byte
	var txBytes []byte
	for _, tx := range b.Transactions {
		txBytes, err = tx.Serialize()
		if err != nil {
			return utils.WrapError("Block_IsValid", err)	// 一个出错，就不可能还原出默克尔根
		}
		txsBytes = append(txsBytes, txBytes)
	}
	merkle := NewMerkleTree(txsBytes)
	merkleRoot := BytesToHash(merkle.RootNode.Data)
	if merkleRoot != b.Hash {
		return utils.WrapError("Block_IsValid", ErrInconsistentMerkleRoot)
	}

	// 检查区块是否是获得了POT。这发生在正常大家竞争POT时然后收到新区块时需要比较新区块内交易数，要是自己的pot表中最大值或更大，否则就不合规
	// TODO: 但这应该放到网络同步中去做。pot只存在与转发节点之中
	// 因为区块的检验还有可能发生在新上线节点或者普通节点中。

	return nil
}