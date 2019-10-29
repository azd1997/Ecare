package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

func main() {
	//testFunction(123, "string", false)
	testInterface()
}

func testFunction(args ...interface{}) {
	for i, v := range args {
		switch v.(type) {
		case int:
			fmt.Println(i, ": ", v, "int")
		case string:
			fmt.Println(i, ": ", v, "string")
		case bool:
			fmt.Println(i, ": ", v, "bool")
		default:
			fmt.Println(i, ": ", v, "unknown")
		}
	}
}

func testInterface() {
	t1 := &TestStruct1{
		name: "eiger",
		age:  23,
	}
	t1Bytes := t1.Serialize()

	t1Copy := &TestStruct1{}
	t1Copy.Deserialize(t1Bytes)
	// t1Copy应该和t1一样

	var t1CC TestInterface
	t1CC.Deserialize(t1Bytes)
}

type TestInterface interface {
	Serialize() []byte
	Deserialize(data []byte)
}

type TestStruct1 struct {
	name string
	age int
}

func (t *TestStruct1) Serialize() []byte {
	var buf bytes.Buffer

	gob.NewEncoder(&buf).Encode(t)
	return buf.Bytes()
}

func (t *TestStruct1) Deserialize(data []byte) {

	var buf bytes.Buffer
	buf.Write(data)
	gob.NewDecoder(&buf).Decode(t)
}

type TestStruct2 struct {
	name string
	age int
}

func (t *TestStruct2) Serialize() []byte {
	var buf bytes.Buffer

	gob.NewEncoder(&buf).Encode(t)
	return buf.Bytes()
}

func (t *TestStruct2) Deserialize(data []byte) {

	var buf bytes.Buffer
	buf.Write(data)
	gob.NewDecoder(&buf).Decode(t)
}


