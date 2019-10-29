package utils

import (
	"fmt"
	"testing"
)

type testStruct struct {
	Data []string
}

func TestGob(t *testing.T) {
	var s = testStruct{
		Data:[]string{"eiger", "zr"},
	}
	fmt.Println(s)

	s1, err := GobEncode(s)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(s1)

	var s2 *testStruct
	err = GobDecode(s1, s2)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(s2)
}
