package eaccount

import (
	"fmt"
	"github.com/azd1997/Ecare/ecoin/account"
	"sync"
	"testing"
)

func TestEcoinAccounts(t *testing.T) {

	es := EAccounts{map[string]*EAccount{}, sync.RWMutex{}}

	// 生成一个账户
	acc, err := account.NewAccount(1)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%s\n", acc)

	// 生成账户id
	id, err := acc.UserId()
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("id: %s\n", id.String())

	// 构建一个ecoinaccount
	ea := &EAccount{
		UserId:id,
		PubKeyField:acc.PubKey,
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
	err = es.AddEAccount(ea)
	if err != nil || es.GetEAccount(id.Id).UserId != id {
		t.Errorf("err: %s; id1 = id : %v", err, es.GetEAccount(id.Id).UserId == id)
	}
	fmt.Printf("%v\n", es)
	fmt.Printf("id1: %s\n", es.GetEAccount(id.Id).UserId.String())

	// 存入到文件
	err = es.SaveFileWithJsonMarshal("./test.eaccounts")
	if err != nil {
		t.Error(err)
	}

	// 加载文件
	es1 := &EAccounts{map[string]*EAccount{}, sync.RWMutex{}}
	err = es1.LoadFileWithJsonUnmarshal("./test.eaccounts")
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%v\n", es1)
	// 新加载出来的键值应该是一样的，但是value由于是指针，所以不一定一样

	if _, ok := es1.Map[id.Id]; !ok {
		t.Error("还原失败")
	}

}
