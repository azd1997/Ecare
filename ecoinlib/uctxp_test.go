package ecoin

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"testing"
)

func TestUCTXP(t *testing.T) {
	// 构造几个交易
	acc, userID := accountAndUserIDForTest()
	_, userIDTo := accountAndUserIDForTest()
	fmt.Printf("from: %s\n", userID)
	fmt.Printf("to: %s\n", userIDTo)

	// tx1
	args1 := &TxGeneralArgs{
		From:        acc,
		FromID:      *userID,
		To:          *userIDTo,
		Amount:      99,
		Description: "txGeneral",
	}
	tx1, err := newTxGeneral(args1)	// new过程已测试Hash/Sign方法
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("tx1: %s\n", tx1)  // 测试String方法
	//tx1Bytes, err := tx1.Serialize()
	//if err != nil {
	//	t.Error(err)
	//}

	// tx2
	args2 := &TxArbitrateArgs{
		Arbitrator:        acc,
		ArbitratorID:*userID,
		TargetTXBytes:[]byte("这是一个待仲裁的交易"),
		TargetTXComplete:false,
		Description: "txArbitrate",
	}
	tx2, err := newTxArbitrate(args2)	// new过程已测试Hash方法
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("tx2: %s\n", tx2)  // 测试String方法
	//tx2Bytes, err := tx2.Serialize()
	//if err != nil {
	//	t.Error(err)
	//}

	// tx3
	args3 := &TxD2PArgs{
		From:        acc,
		FromID:*userID,
		P2DBytes:[]byte("这是一个待仲裁的交易"),
		Response:[]byte("这是一个回应"),
		Description: "txD2P",
	}
	tx3, err := newTxD2P(args3)	// new过程已测试Hash方法
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("tx3: %s\n", tx3)  // 测试String方法
	//tx3Bytes, err := tx3.Serialize()
	//if err != nil {
	//	t.Error(err)
	//}

	// 构造uctxp
	uctxp := &UnCompleteTXPool{Map: map[string]TX{
		string(tx1.ID): tx1,
		string(tx2.ID): tx2,
		string(tx3.ID): tx3}}
	//uctxp := &UnCompleteTXPool{Map: map[string][]byte{
	//	string(tx1.Id): tx1Bytes,
	//	string(tx2.Id): tx2Bytes,
	//	string(tx3.Id): tx3Bytes}}

	// json文件
	err = uctxp.SaveFileWithJsonMarshal(9999)
	if err != nil {
		t.Error(err)
	}

	// unmarshal
	uctxp1 := &UnCompleteTXPool{}
	err = uctxp1.LoadFileWithJsonUnmarshal(9999)
	if err != nil {
		t.Error(err)
	}
	fmt.Println("uctxp1: ")
	fmt.Printf("%s", uctxp1)
}

func TestUCTXPWithGobEncode(t *testing.T) {
	// 构造几个交易
	acc, userID := accountAndUserIDForTest()
	_, userIDTo := accountAndUserIDForTest()
	fmt.Printf("from: %s\n", userID)
	fmt.Printf("to: %s\n", userIDTo)

	// tx1
	args1 := &TxGeneralArgs{
		From:        acc,
		FromID:      *userID,
		To:          *userIDTo,
		Amount:      99,
		Description: "txGeneral",
	}
	tx1, err := newTxGeneral(args1)	// new过程已测试Hash/Sign方法
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("tx1: %s\n", tx1)  // 测试String方法
	//tx1Bytes, err := tx1.Serialize()
	//if err != nil {
	//	t.Error(err)
	//}

	// tx2
	args2 := &TxArbitrateArgs{
		Arbitrator:        acc,
		ArbitratorID:*userID,
		TargetTXBytes:[]byte("这是一个待仲裁的交易"),
		TargetTXComplete:false,
		Description: "txArbitrate",
	}
	tx2, err := newTxArbitrate(args2)	// new过程已测试Hash方法
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("tx2: %s\n", tx2)  // 测试String方法
	//tx2Bytes, err := tx2.Serialize()
	//if err != nil {
	//	t.Error(err)
	//}

	// tx3
	args3 := &TxD2PArgs{
		From:        acc,
		FromID:*userID,
		P2DBytes:[]byte("这是一个待仲裁的交易"),
		Response:[]byte("这是一个回应"),
		Description: "txD2P",
	}
	tx3, err := newTxD2P(args3)	// new过程已测试Hash方法
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("tx3: %s\n", tx3)  // 测试String方法
	//tx3Bytes, err := tx3.Serialize()
	//if err != nil {
	//	t.Error(err)
	//}

	// 构造uctxp
	uctxp := &UnCompleteTXPool{Map: map[string]TX{
		string(tx1.ID): tx1,
		string(tx2.ID): tx2,
		string(tx3.ID): tx3}}
	//uctxp := &UnCompleteTXPool{Map: map[string][]byte{
	//	string(tx1.Id): tx1Bytes,
	//	string(tx2.Id): tx2Bytes,
	//	string(tx3.Id): tx3Bytes}}

	// json文件
	err = uctxp.SaveFileWithGobEncode(9999)
	if err != nil {
		t.Error(err)
	}

	// unmarshal
	uctxp1 := &UnCompleteTXPool{}
	err = uctxp1.LoadFileWithGobDecode(9999)
	if err != nil {
		t.Error(err)
	}

	//for k, v := range uctxp.Map {
	//	var v1 []byte
	//	var ok bool
	//	if v1, ok = uctxp1.Map[k]; !ok {
	//		t.Error("出错")
	//	}
	//	if string(v1) != string(v) {
	//		t.Error("出错了了")
	//	}
	//}
}

// TODO: 有下面这个测试得知，gob可以对包含接口的结构进行编解码（需要提前注册号实现接口的所有类型，更具体一点是想要编码的那个变量中具体有的实现接口的类型），
//  json编码很方便，解码似乎行不通。
//  结论是：可以保存文件时可以同时两种编码，json的用来自己查看。解码只用gob。
//

type Animal interface {
	Speak() string
}
type Dog struct {
	Name string
	Age int
	Color int
}
func (d *Dog) Speak() string {
	return d.Name
}
type Cat struct {
	Name string
	Color int
	Legs int
}
func (c *Cat) Speak() string {
	return c.Name
}
type AnimalList struct {
	Map map[string]Animal
}
func TestGobEncodeInterface(t *testing.T) {
	cat1 := &Cat{"tom", 2, 4}
	cat2 := &Cat{"jack", 3, 4}
	dog := &Dog{"bob", 1, 4}
	list := &AnimalList{map[string]Animal{
		cat1.Speak():cat1,
		cat2.Speak():cat2,
		dog.Speak():dog,
	}}
	fmt.Printf("list: \n%v\n", list)
	//
	//gob.Register(&Cat{})
	//gob.Register(&Dog{})
	GobRegister(&Cat{}, &Dog{})
	listBytes, err := GobEncode(list)
	if err != nil {
		t.Error(err)
	}

	var list1 AnimalList

	err = gob.NewDecoder(bytes.NewReader(listBytes)).Decode(&list1)
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("list1: \n%v\n", list1)

}

func TestJsonMarshalInterface(t *testing.T) {
	cat1 := &Cat{"tom", 2, 4}
	cat2 := &Cat{"jack", 3, 4}
	dog := &Cat{"bob", 1, 4}
	list := &AnimalList{map[string]Animal{
		cat1.Speak():cat1,
		cat2.Speak():cat2,
		dog.Speak():dog,
	}}


	listBytes, err := json.Marshal(list)
	if err != nil {
		t.Error(err)
	}

	var list1 AnimalList
	err = json.Unmarshal(listBytes, &list1)
	if err != nil {
		t.Error(err)
	}

}