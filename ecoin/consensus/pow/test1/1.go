package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
)

//声明了一个挖矿难度Bits
const N  = 0x0FFFffffffff

//定义一个block结构体
type block struct {
	//区块数据
	Data  []byte
	//随机数
	Nonce int
	//当前块的hash
	hash []byte
	//挖矿目标难度值
	Bits int
}

func main() {
	//初始化一个空块
	var b = block{[]byte("helloworld"), 0,nil,N}

	nonce := 0
	for {

		//拼接block字段内容
		dataBytes := b.PreParae(nonce)
		//矿工计算函数，我只用了一次256hash
		hash := sha256.Sum256(dataBytes)
		//将hash转换成Uint64类型，将与N进行比较大小
		hash1 := BytesToUint64(hash[:])
		//不断显现hash函数后的值
		fmt.Printf("\r%x", hash)
		//hash值与目标值进行大小比较
		if hash1 < uint64(N) {
			//挖矿成功后，给b重新赋值，并跳出循环
			fmt.Println()
			b.Nonce=nonce
			b.hash=hash[:]
			break
		} else {
			nonce++
		}
	}
	fmt.Println(b)
}
//block字段拼接
func (b block) PreParae(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			b.Data,
			IntToHex(int64(nonce)),
			b.hash,
		},
		[]byte{},
	)
	return data
}
//字节转换成64进制
func BytesToUint64(array []byte) uint64 {
	var data uint64 = 0
	for i := 0; i < len(array); i++ {
		data = data + uint64(uint(array[i])<<uint(8*i))
	}

	return data
}
//将64进制数字转换成字节数组
func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}
