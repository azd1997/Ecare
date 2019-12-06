package common

import "fmt"

// Signature 签名
type Signature []byte

// String 打印方法
func (s Signature) String() string {
	// 直接转为string会出现乱码，所以还是转为16进制打印
	return fmt.Sprintf("%x", string(s))
}