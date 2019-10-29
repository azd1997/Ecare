package ecoinlib

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
	"time"
)

type BlockHeader struct {
	Id         int // 从0开始，很多称为height
	CreateTime int64
	CommitTime int64
	PassTime   int64
	PrevHash   []byte
	Hash       []byte
	CreateBy   UserID // 由哪个账户创建
}

type BlockBody struct {
	Transactions []*Transaction
}

type Block struct {
	BlockHeader
	BlockBody
}

func (b *Block) String() string {
	return fmt.Sprintf(
		`{
	id: 		%d
	createTime: 	%s
	commitTime: 	%s
	prevHash: 	%s
	hash: 		%s
	createBy: 	%s
}`,
		b.Id,
		time.Unix(b.CreateTime, 0).Format("2006/01/02 15:04:05"),
		time.Unix(b.CommitTime, 0).Format("2006/01/02 15:04:05"),
		b.PrevHash,
		b.Hash,
		b.CreateBy)
}

func NewBlock(txs []*Transaction, prevHash []byte, id int, createBy UserID) *Block {
	l := len(txs)
	var txsBytes = make([][]byte, l)
	var err error
	for i, tx := range txs {
		txsBytes[i], err = tx.Serialize()
		if err != nil {
			// 记录下错误并继续
			log.Printf("NewBlock: %s\n", err)
			continue
		}
	}
	merkle := NewMerkleTree(txsBytes)

	return &Block{
		BlockHeader: BlockHeader{
			Id:         id,
			CreateTime: time.Now().Unix(),
			PrevHash:   prevHash,
			Hash:       merkle.RootNode.Data,
			CreateBy:   createBy,
		},
		BlockBody: BlockBody{
			Transactions: txs,
		},
	}
}

func GenesisBlock(coinbase *Transaction, createBy UserID) (gb *Block, err error) {

	// 创世区块就一个交易懒得用merkle
	coinbaseTxBytes, err := coinbase.Serialize()
	if err != nil {
		return nil, fmt.Errorf("GenesisBlock: %s", err)
	}
	hash := sha256.Sum256(coinbaseTxBytes)

	return &Block{
		BlockHeader: BlockHeader{
			Id:         0,
			CreateTime: time.Now().Unix(),
			Hash:       hash[:],
			CreateBy:   createBy,
		},
		BlockBody: BlockBody{
			Transactions: []*Transaction{coinbase},
		},
	}, nil
}

func (b *Block) Serialize() (res []byte, err error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err = encoder.Encode(b); err != nil {
		return nil, fmt.Errorf("Block_Serialize: %s", err)
	}
	return buf.Bytes(), nil
}

func Deserialize(blockBytes []byte) (b *Block, err error) {
	decoder := gob.NewDecoder(bytes.NewReader(blockBytes))
	if err = decoder.Decode(b); err != nil {
		return nil, fmt.Errorf("Deserialize_Block: %s", err)
	}
	return b, nil
}

func (b *Block) PrintTransactions() {
	for _, tx := range b.Transactions {
		Info("%s", tx.String())
	}
}