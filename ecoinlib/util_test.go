package ecoin

import (
	"fmt"
	"testing"
)

// 经过测试得知，Hash直接json的话会出现数组元素换行打印情况
func TestHashToJson(t *testing.T) {
	type TestHash struct {
		Hash Hash `json:"hash"`
	}
	str := JsonMarshalIndentToString(&TestHash{Hash:RandomHash()})
	fmt.Printf("%s\n", str)
}

// 经过测试得知，[32]byte直接json的话会出现数组元素换行打印情况
func TestHashToJson2(t *testing.T) {
	type TestHash struct {
		Hash []byte `json:"hash"`
	}
	str := JsonMarshalIndentToString(&TestHash{Hash:RandomHash()})
	fmt.Printf("%s\n", str)
}

// 经过测试得知，Hash转为[]byte再去打印会比较好
func TestHashToJson3(t *testing.T) {
	type TestHash struct {
		Hash []byte `json:"hash"`
	}
	hash := RandomHash()
	str := JsonMarshalIndentToString(&TestHash{Hash:hash[:]})
	fmt.Printf("%s\n", str)
}

func TestGobEncode(t *testing.T) {
	b := Block{
		BlockHeader: BlockHeader{
			Id:12,
		},
		BlockBody:   BlockBody{
			Transactions: [][]byte{},
		},
	}
	_, err := GobEncode(b)
	if err != nil {
		t.Error(err)
	}
}

func TestExtractDirFromFilePath(t *testing.T) {
	filePath := "./tmp/ttt/234.txt"
	dir := ExtractDirFromFilePath(filePath)
	if dir != "./tmp/ttt" {
		t.Error("提取错误，得到： ", dir)
	}
	fmt.Println(dir)
}