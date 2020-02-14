package eaddr

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestAddr_EAddr_EAddrs(t *testing.T) {
	a_s := []Addr{
		{Ip:   "127.0.0.1", Port: 8000,},
		{Ip:   "127.0.0.1", Port: 8008,},
		{Ip:   "127.0.0.1", Port: 8009,},
		{Ip:   "127.0.0.1", Port: 8010,}}

	ea_s := make([]EAddr, len(a_s))
	for i:=0; i<len(a_s); i++ {
		ea_s[i] = NewEAddr(a_s[i], "地址1")
	}

	eas := NewEAddrs()
	eas.SetEAddrBatch(ea_s...)

	fmt.Println(eas)

	// 设置所有人诚实可达，再设置ea3不诚实
	for i:=0; i<len(a_s); i++ {
		ea_s[i].setHonest(true)
		ea_s[i].setReachable(true)
		eas.SetEAddrBatch(ea_s...)
	}
	ea_s[2].setHonest(false)
	eas.SetEAddr(ea_s[2])

	// 设置四个地址pingdelay
	for i:=0; i<len(a_s); i++ {
		ea_s[i].PingStart()
		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
		ea_s[i].PingStop()
		fmt.Println(ea_s[i].PingDelay())
	}

	fmt.Println(eas)

	sortedValidAddrs := eas.SortedValidAddrs()
	fmt.Println(sortedValidAddrs)
}
