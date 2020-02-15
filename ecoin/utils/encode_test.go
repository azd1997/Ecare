package utils

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"testing"
)

// 测试数组的编码与解码: 测试结果：对于切片类型，在作为gob解码接收者时，需要传入指针
func TestGobEncode(t *testing.T) {
	data := []string{"张三", "李四", "王麻子"}
	enced, err := GobEncode(data)
	if err != nil {
		t.Errorf("Encode failed: %v\n", err)
	}
	var newdata []string
	err = gob.NewDecoder(bytes.NewReader(enced)).Decode(&newdata)
	if err != nil {
		t.Errorf("Decode failed: %v\n", err)
	}

	fmt.Println(newdata)
}
