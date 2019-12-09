package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"time"
)



func main()  {

	// 创世区块
	blockchain := CreateBlockchainWithGenesisBlock()

	// 新区块
	blockchain.AddBlockToBlockchain("Send 100RMB To tom",blockchain.Blocks[len(blockchain.Blocks) - 1].Height + 1,blockchain.Blocks[len(blockchain.Blocks) - 1].Hash)

	blockchain.AddBlockToBlockchain("Send 200RMB To lily",blockchain.Blocks[len(blockchain.Blocks) - 1].Height + 1,blockchain.Blocks[len(blockchain.Blocks) - 1].Hash)

	blockchain.AddBlockToBlockchain("Send 300RMB To hanmeimei",blockchain.Blocks[len(blockchain.Blocks) - 1].Height + 1,blockchain.Blocks[len(blockchain.Blocks) - 1].Hash)

	blockchain.AddBlockToBlockchain("Send 50RMB To lucy",blockchain.Blocks[len(blockchain.Blocks) - 1].Height + 1,blockchain.Blocks[len(blockchain.Blocks) - 1].Hash)


	fmt.Println(blockchain)
	fmt.Println(blockchain.Blocks)
}


type Block struct {
	//1. 区块高度
	Height int64
	//2. 上一个区块HASH
	PrevBlockHash []byte
	//3. 交易数据
	Data []byte
	//4. 时间戳
	Timestamp int64
	//5. Hash
	Hash []byte
	// 6. Nonce
	Nonce int64
}


//1. 创建新的区块
func NewBlock(data string,height int64,prevBlockHash []byte) *Block {

	//创建区块
	block := &Block{height,prevBlockHash,[]byte(data),time.Now().Unix(),nil,0}

	// 调用工作量证明的方法并且返回有效的Hash和Nonce
	pow := NewProofOfWork(block)

	// 挖矿验证
	hash,nonce := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	fmt.Println("")

	return block

}

//2. 单独写一个方法，生成创世区块

func CreateGenesisBlock(data string) *Block {

	return NewBlock(data,1, []byte{0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0})
}


type Blockchain struct {
	Blocks []*Block  // 存储有序的区块
}


// 增加区块到区块链里面
func (blc *Blockchain) AddBlockToBlockchain(data string,height int64,preHash []byte)  {
	// 创建新区块
	newBlock := NewBlock(data,height,preHash)
	// 往链里面添加区块
	blc.Blocks = append(blc.Blocks,newBlock)
}


//1. 创建带有创世区块的区块链
func CreateBlockchainWithGenesisBlock() *Blockchain {
	// 创建创世区块
	genesisBlock := CreateGenesisBlock("Genesis Data.......")
	// 返回区块链对象
	return &Blockchain{[]*Block{genesisBlock}}
}

// 将int64转换为字节数组
func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}
