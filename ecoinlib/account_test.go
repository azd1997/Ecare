package ecoin

import (
	"fmt"
	"testing"
)

func TestNewAccount(t *testing.T) {
	acc, err := NewAccount(1)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%s\n", acc)
}

func TestAccount_UserID(t *testing.T) {
	acc, err := NewAccount(1)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%s\n", acc)

	userid, err := acc.UserID()
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%s\n", userid.String())

	valid, err := userid.IsValid()
	if err != nil {
		t.Error(err)
	}
	if !valid {
		t.Error("userid is invalid")
	}
	fmt.Println("userid is valid")
}

func TestAccount_Sign(t *testing.T) {
	acc, err := NewAccount(1)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%s\n", acc)

	targetForSign := RandomHash()
	fmt.Printf("Hash: %v\n", targetForSign)

	sig, err := acc.Sign(targetForSign[:])
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("Sig: %v\n", sig)

	yes := VerifySignature(targetForSign[:], sig, acc.PubKey)
	if !yes {
		t.Error("verify sig failed")
	}
	fmt.Println("verify sig success")
}

func TestAccount_NewTX_Coinbase(t *testing.T) {
	acc, err := NewAccount(1)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%s\n", acc)

	userid, err := acc.UserID()
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%s\n", userid.String())

	// 这里直接构造交易，由于TXcoinbase参数检查时虽然传入gsm但未使用，所以并不会报错。
	tx, err := acc.NewTX(TX_COINBASE, &TxCoinbaseArgs{
		To:          userid,
		Amount:      60,
		Description: "",
	})
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("tx: %s\n", tx)
}

func TestSelfAccounts_AddAccount(t *testing.T) {
	sa := SelfAccounts{Map: map[string]*Account{}}
	fmt.Printf("%v\n", sa)

	userid, err := sa.AddAccount(1)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("userid: %s\n", userid.String())

	fmt.Printf("%v\n", sa)
}

func TestSelfAccounts_SaveFileWithGobEncode(t *testing.T) {
	sa := SelfAccounts{Map: map[string]*Account{}}
	fmt.Printf("%v\n", sa)

	userid, err := sa.AddAccount(1)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("userid: %s\n", userid.String())

	fmt.Printf("%v\n", sa)

	// 保存
	err = sa.SaveFileWithGobEncode(9999)
	if err != nil {
		t.Error(err)
	}
	// 查看是否生成了相应文件。确实是生成了
}

// TODO：JSONmarshal不能解码回账户，因为privKey含有私有变量

func TestSelfAccounts_LoadFileWithGobDecode(t *testing.T) {

	sa := SelfAccounts{Map: map[string]*Account{}}
	fmt.Printf("%v\n", sa)

	userid, err := sa.AddAccount(1)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("userid: %s\n", userid.String())

	fmt.Printf("%v\n", sa)

	// 保存
	err = sa.SaveFileWithGobEncode(9999)
	if err != nil {
		t.Error(err)
	}

	sa1 := SelfAccounts{map[string]*Account{}}
	err = sa1.LoadFileWithGobDecode(9999)
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("%v\n", sa)
	// 看能否恢复出来

}

func TestEcoinAccounts(t *testing.T) {

	es := EcoinAccounts{map[string]*EcoinAccount{}}

	// 生成一个账户
	acc, err := NewAccount(1)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%s\n", acc)

	// 生成账户id
	id, err := acc.UserID()
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("id: %s\n", id.String())

	// 构建一个ecoinaccount
	ea := &EcoinAccount{
		UserID:id,
		PublicKey:acc.PubKey,
		BalanceCoin:0,
		RoleField:Role{
			NoField:             1,
			AliasField:          "hospital",
			InitialField:        0,
			CoinbaseRewardField: 0,
			GenesisRewardField:  0,
			EnableKsEsField:     false,
			KsField:             nil,
			EsField:             nil,
		},
		AvailableField:true,
	}

	// 加入到ecoinaccounts
	id1, err := es.AddAccount(ea)
	if err != nil || id1 != id {
		t.Errorf("err: %s; id1 = id : %v", err, id1 == id)
	}
	fmt.Printf("%v\n", es)
	fmt.Printf("id1: %s\n", id1.String())

	// 存入到文件
	err = es.SaveFileWithJsonMarshal(9999)
	if err != nil {
		t.Error(err)
	}

	// 加载文件
	es1 := &EcoinAccounts{map[string]*EcoinAccount{}}
	err = es1.LoadFileWithJsonUnmarshal(9999)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%v\n", es1)
	// 新加载出来的键值应该是一样的，但是value由于是指针，所以不一定一样

	if _, ok := es1.Map[id.ID]; !ok {
		t.Error("还原失败")
	}

}