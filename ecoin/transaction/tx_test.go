package tx

import "fmt"

// TxCoinbase在测试区块等时已测试过

// TODO: 带有gsm参数的方法待测试

func accountAndUserIDForTest() (*Account, *UserID) {
	// 构造交易
	acc, err := NewAccount(1)
	if err != nil {
		panic(err)
	}

	userid, err := acc.UserID()
	if err != nil {
		panic(err)
	}

	return acc, &userid
}

func TestNewTxGeneral(t *testing.T) {
	acc, userID := accountAndUserIDForTest()
	_, userIDTo := accountAndUserIDForTest()
	fmt.Printf("from: %s\n", userID)
	fmt.Printf("to: %s\n", userIDTo)

	// 构造参数
	args := &GeneralArgs{
		From:        acc,
		FromID:      *userID,
		To:          *userIDTo,
		Amount:      99,
		Description: "txGeneral",
	}

	// 生成交易
	tx, err := newTxGeneral(args)	// new过程已测试Hash/Sign方法
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("tx: %s\n", tx)  // 测试String方法

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
	fmt.Printf("arbitrator: %s\n", userID)

	// 构造参数
	args := &TxArbitrateArgs{
		Arbitrator:        acc,
		ArbitratorID:*userID,
		TargetTXBytes:[]byte("这是一个待仲裁的交易"),
		TargetTXComplete:false,
		Description: "txArbitrate",
	}

	// 生成交易
	tx, err := newTxArbitrate(args)	// new过程已测试Hash方法
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("tx: %s\n", tx)  // 测试String方法

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
	fmt.Printf("from: %s\n", userID)

	// 构造参数
	args := &TxD2PArgs{
		From:        acc,
		FromID:*userID,
		P2DBytes:[]byte("这是一个待仲裁的交易"),
		Response:[]byte("这是一个回应"),
		Description: "txD2P",
	}

	// 生成交易
	tx, err := newTxD2P(args)	// new过程已测试Hash方法
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("tx: %s\n", tx)  // 测试String方法

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

