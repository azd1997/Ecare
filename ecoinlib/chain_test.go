package ecoin

import (
	"encoding/base64"
	"fmt"
	"testing"
)

// TODO: MineBlock和VerifyBlock暂不测试

func TestInitChain(t *testing.T) {
	// 构造option
	acc, err := NewAccount(1)
	if err != nil {
		t.Error(err)
	}
	opts := DefaultOption().SetAccount(*acc)
	fmt.Printf("%#v\n", opts)

	// 初始化区块链
	chain, err := InitChain(&InitChainArgs{
		coinbase: &TxCoinbaseArgs{
			To:          *opts.UserID(),
			Amount:      77,
			Description: "wwwwww",
		},
		opts:     opts,
	})
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("chain: %v\n", chain)

	// 构造两个交易
	tx1, err := acc.NewTX(TX_COINBASE, &TxCoinbaseArgs{
		To:          *opts.UserID(),
		Amount:      60,
		Description: "hhhh",
	})
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("tx1: %s\n", tx1)
	tx2, err := acc.NewTX(TX_COINBASE, &TxCoinbaseArgs{
		To:          *opts.UserID(),
		Amount:      99,
		Description: "sssssssssssss",
	})
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("tx2: %s\n", tx2)

	// 构造新的区块
	lastId, err := chain.GetMaxId()
	if err != nil {
		t.Error(err)
	}
	b1 := NewBlock([]TX{tx1, tx2}, chain.LastHash, uint(lastId) + 1, *opts.UserID())
	fmt.Printf("block: %s\n", b1)

	// 添加新区快
	err = chain.AddBlock(b1)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("chain: %v\n", chain)
	// 这时lasthash已经改变

	// 打印区块链
	//err = chain.PrintBlockHeaders(0, uint(lastId) + 1)
	//if err != nil {
	//	t.Error(err)
	//}

	// 获取全部区块哈希
	blockHashes, err := chain.GetBlockHashes()
	if err != nil {
		t.Error(err)
	}
	fmt.Println("blockHashes: ")
	for i, v := range blockHashes {
		fmt.Printf("block[%d]: %s\n", i, base64.StdEncoding.EncodeToString(v[:]))
	}

	// 根据区块哈希提取区块
	b1Copy, err := chain.GetBlockByHash(b1.Hash)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("b1Copy: %s\n", b1Copy)
	if string(b1Copy.Hash) != string(b1.Hash) {
		t.Error("哈希不一致")
	}

	// 根据区块id提取区块
	b0Copy, err := chain.GetBlockById(0)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("b0Copy: %s\n", b0Copy)

	// 判断是否有区块
	has, err := chain.HasBlock(b1Copy)
	if err != nil || has == false {
		t.Error(err)
	}
	fmt.Println("b1Copy is in the chain")

	// 寻找交易
	tx2Copy, err := chain.FindTransaction(tx2.Id())
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("tx2Copy: %s\n", tx2Copy)


}

func TestContinueChain(t *testing.T) {

	// TODO： 这里就不管opts中账户发生变化了，只是为了得到opts.Port()来组装路径

	// 构造option
	acc, err := NewAccount(1)
	if err != nil {
		t.Error(err)
	}
	opts := DefaultOption().SetAccount(*acc)
	fmt.Printf("%#v\n", opts)

	// 初始化区块链
	chain, err := ContinueChain(&ContinueChainArgs{
		opts:     opts,
	})
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("chain: %v\n", chain)

	// 查看一下chain.LastHash是否与TestINitChain一致
	// [77 118 157 201 71 61 108 152 215 86 205 78 122 105 19 153 80 160 57 194 242 81 181 152 232 77 36 181 11 66 66 243]
	// TXadyUc9bJjXVs1OemkTmVCgOcLyUbWY6E0ktQtCQvM=
	fmt.Printf("LastHash: %s\n", base64.StdEncoding.EncodeToString(chain.LastHash[:]))

	b0, err := chain.GetBlockById(0)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("b0: %s\n", b0)
	b1, err := chain.GetBlockById(1)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("b1: %s\n", b1)

	fmt.Printf("b1.Hash: %s\n", base64.StdEncoding.EncodeToString(b1.Hash[:]))

	if string(b1.Hash) != string(chain.LastHash) {
		t.Error("lastHash不一致")
	}
}