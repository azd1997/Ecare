package transaction

import (
	"fmt"
	"testing"

	"github.com/azd1997/Ecare/ecoin/account"
)

func accountAndUserIDForTest() (*account.Account, account.UserId) {
	// 构造交易
	acc, err := account.NewAccount(1)
	if err != nil {
		panic(err)
	}

	userid, err := acc.UserId()
	if err != nil {
		panic(err)
	}

	return acc, userid
}

func TestNewTxGeneral(t *testing.T) {
	acc, userID := accountAndUserIDForTest()
	_, userIDTo := accountAndUserIDForTest()
	fmt.Printf("from: %v\n", userID)
	fmt.Printf("to: %v\n", userIDTo)

	// 构造参数
	args := &GeneralArgs{
		FromAccount: *acc,
		From:        userID,
		To:          userIDTo,
		Amount:      99,
		Description: "txGeneral",
	}

	// 生成交易
	tx, err := NewTXWithArgsCheck(TX_GENERAL, args) // new过程已测试Hash/Sign方法
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("tx: %s\n", tx) // 测试String方法

	// 测试序列化与反序列化
	txBytes, err := tx.Serialize()
	if err != nil {
		t.Error(err)
	}
	tx1 := &TxGeneral{}
	err = tx1.Deserialize(txBytes)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("tx1: %s\n", tx1)
}

func TestNewTxArbitrate(t *testing.T) {
	acc, userID := accountAndUserIDForTest()
	fmt.Printf("arbitrator: %v\n", userID)

	// 构造参数
	args := &ArbitrateArgs{
		ArbitratorAccount: *acc,
		Arbitrator:        userID,
		TargetTX:          &TxR2P{},
		Arbitration:       BuyerFaultLevel3,
		Description:       "txArbitrate",
	}

	// 生成交易
	tx, err := NewTXWithArgsCheck(TX_ARBITRATE, args) // new过程已测试Hash方法
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("tx: %s\n", tx) // 测试String方法

	// 测试序列化与反序列化
	txBytes, err := tx.Serialize()
	if err != nil {
		t.Error(err)
	}
	tx1 := &TxArbitrate{}
	err = tx1.Deserialize(txBytes)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("tx1: %s\n", tx1)
}

func TestNewTxD2P(t *testing.T) {
	acc, userID := accountAndUserIDForTest()
	fmt.Printf("from: %v\n", userID)

	// 构造参数
	args := &D2PArgs{
		FromAccount: *acc,
		From:        userID,
		P2D:         &TxP2D{},
		Response:    []byte("这是一个回应"),
		Description: "txD2P",
	}

	// 生成交易
	tx, err := newTxD2P(args) // new过程已测试Hash方法
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("tx: %s\n", tx) // 测试String方法

	// 测试序列化与反序列化
	txBytes, err := tx.Serialize()
	if err != nil {
		t.Error(err)
	}
	tx1 := &TxD2P{}
	err = tx1.Deserialize(txBytes)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("tx1: %s\n", tx1)
}


func TestValidateFunc(t *testing.T) {
	g := &GSM{}

	tx := new(TxBase)
	if err := tx.IsValid(g.Validate); err != nil {

	}
}