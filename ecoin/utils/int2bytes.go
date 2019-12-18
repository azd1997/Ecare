package utils

import (
	"bytes"
	"encoding/binary"
)


//整形转换成字节
func Int32ToBytes(n int32) []byte {
	bytesBuffer := bytes.NewBuffer([]byte{})
	_ = binary.Write(bytesBuffer, binary.BigEndian, n)
	return bytesBuffer.Bytes()
}
//字节转换成整形
func BytesToInt32(b []byte) int32 {
	bytesBuffer := bytes.NewBuffer(b)

	var x int32
	_ = binary.Read(bytesBuffer, binary.BigEndian, &x)

	return x
}

//整形转换成字节
func Uint32ToBytes(n uint32) []byte {
	bytesBuffer := bytes.NewBuffer([]byte{})
	_ = binary.Write(bytesBuffer, binary.BigEndian, n)
	return bytesBuffer.Bytes()
}
//字节转换成整形
func BytesToUint32(b []byte) uint32 {
	bytesBuffer := bytes.NewBuffer(b)

	var x uint32
	_ = binary.Read(bytesBuffer, binary.BigEndian, &x)

	return x
}
