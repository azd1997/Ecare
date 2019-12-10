package p2p

import (
	"fmt"
	"sort"
	"testing"
)

func TestAddress_String(t *testing.T) {
	addr := Address{
		Ipv4Port: "127.0.0.1:9999",
		Alias:    "self",
		PingTime: 0,
		Honest:   true,
	}
	fmt.Printf("addr: %s\n", addr.String())
}

func TestAddressList(t *testing.T) {
	addr1 := Address{
		Ipv4Port: "127.0.0.1:9991",
		Alias:    "self",
		PingTime: 3,
		Honest:   true,
	}
	addr2 := Address{
		Ipv4Port: "127.0.0.1:9992",
		Alias:    "self",
		PingTime: 9,
		Honest:   false,
	}
	addr3 := Address{
		Ipv4Port: "127.0.0.1:9993",
		Alias:    "self",
		PingTime: 50,
		Honest:   true,
	}
	addr4 := Address{
		Ipv4Port: "127.0.0.1:9994",
		Alias:    "self",
		PingTime: 100,
		Honest:   true,
	}

	addrs := []*Address{&addr1, &addr2, &addr3, &addr4}
	fmt.Println("排序之前：")
	for _, addr := range addrs {
		fmt.Println(addr.String())
	}

	sort.Sort(AddressList{list:addrs, less:AddressListLessFn})
	fmt.Println("排序之后：")
	for _, addr := range addrs {
		fmt.Println(addr.String())
	}
}

func TestAddrLists_L1Ipv4Honest(t *testing.T) {
	addr1 := Address{
		Ipv4Port: "127.0.0.1:9991",
		Alias:    "self",
		PingTime: 3,
		Honest:   true,
	}
	addr2 := Address{
		Ipv4Port: "127.0.0.1:9992",
		Alias:    "self",
		PingTime: 9,
		Honest:   false,
	}
	addr3 := Address{
		Ipv4Port: "127.0.0.1:9993",
		Alias:    "self",
		PingTime: 50,
		Honest:   true,
	}
	addr4 := Address{
		Ipv4Port: "127.0.0.1:9994",
		Alias:    "self",
		PingTime: 100,
		Honest:   true,
	}
	addrs := AddrLists{
		L1: []*Address{&addr1, &addr2, &addr3, &addr4},
		L2: []*Address{&addr1, &addr3, &addr4},
		L3: []*Address{&addr2, &addr3, &addr4},
	}

	fmt.Println("排序之前：")
	fmt.Printf("%#v\n", addrs)
	l1 := addrs.L1Ipv4Honest()
	fmt.Printf("l1: %v\n", l1)

	addrs.Sort()
	fmt.Println("排序之后：")
	fmt.Printf("%#v\n", addrs)
	l1 = addrs.L1Ipv4Honest()
	fmt.Printf("l1: %v\n", l1)
}

// 由于Address里边都是导出字段，所以可以使用json来作编码
func TestAddrLists_SaveFileWithJsonMarshal(t *testing.T) {
	addr1 := Address{
		Ipv4Port: "127.0.0.1:9991",
		Alias:    "self",
		PingTime: 3,
		Honest:   true,
	}
	addr2 := Address{
		Ipv4Port: "127.0.0.1:9992",
		Alias:    "self",
		PingTime: 9,
		Honest:   false,
	}
	addr3 := Address{
		Ipv4Port: "127.0.0.1:9993",
		Alias:    "self",
		PingTime: 50,
		Honest:   true,
	}
	addr4 := Address{
		Ipv4Port: "127.0.0.1:9994",
		Alias:    "self",
		PingTime: 100,
		Honest:   true,
	}
	addrs := AddrLists{
		L1: []*Address{&addr1, &addr2, &addr3, &addr4},
		L2: []*Address{&addr1, &addr3, &addr4},
		L3: []*Address{&addr2, &addr3, &addr4},
	}

	fmt.Printf("%#v\n", addrs)
	fmt.Printf("l1: %v\n", addrs.L1Ipv4Honest())

	err := addrs.SaveFileWithJsonMarshal("tmp/addrs.json")
	if err != nil {
		t.Error(err)
	}

	addrs1 := AddrLists{}
	err = addrs1.LoadFileWithJsonUnmarshal("tmp/addrs.json")
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("%#v\n", addrs1)
	fmt.Printf("l1: %v\n", addrs1.L1Ipv4Honest())
	// 比较两次打印的l1是否相同
}