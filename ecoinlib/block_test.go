package ecoin

import (
	"fmt"
	"testing"
)

func TestBlock(t *testing.T) {

	// 构造交易
	acc, err := NewAccount(1)
	if err != nil {
		t.Error(err)
	}

	userid, err := acc.UserID()
	if err != nil {
		t.Error(err)
	}
	// 这里直接构造交易，由于TXcoinbase参数检查时虽然传入gsm但未使用，所以并不会报错。
	tx1, err := acc.NewTX(TX_COINBASE, &TxCoinbaseArgs{
		To:          userid,
		Amount:      60,
		Description: "hhhh",
	})
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("tx1: %s\n", tx1)
	tx2, err := acc.NewTX(TX_COINBASE, &TxCoinbaseArgs{
		To:          userid,
		Amount:      99,
		Description: "sssssssssssss",
	})
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("tx2: %s\n", tx2)

	// 构造一个区块，尽管是不符合要求的区块
	b := NewBlock([]TX{tx1, tx2}, Hash{}, 0, userid)
	fmt.Printf("block: %s\n", b)

	// 打印所有交易
	b.PrintTransactions()

	// 序列化
	bBytes, err := b.Serialize()
	if err != nil {
		t.Error(err)
	}

	// 反序列化
	b1 := &Block{}
	err = b1.Deserialize(bBytes)
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("block1: %s\n", b1)

}

// TODO: Block_IsValid测试、VerifyTxs