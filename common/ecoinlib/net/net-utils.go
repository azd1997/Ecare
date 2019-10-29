package net

import (
	"fmt"
	"github.com/azd1997/Ecare/common/ecoinlib/types"
)

func NodeExists(nodeAddr types.Address, KnownNodeAddrList []types.Address) bool {
	for _, node := range KnownNodeAddrList {
		if node == nodeAddr {
			return true
		}
	}
	return false
}

func NodeLocate(nodeAddr types.Address, KnownNodeAddrList []types.Address) int {
	for i, node := range KnownNodeAddrList {
		if node == nodeAddr {
			return i
		}
	}
	return -1
}

func MergeTwoNodeList(l1, l2 []types.Address) (l3 []types.Address) {
	l3 = l1
	for _, v := range l2 {
		if !NodeExists(v, l3) {
			l3 = append(l3, v)
		}
	}
	return l3
}

func BytesToCmd(cmdBytes []byte) string {
	var cmd []byte
	for _, b := range cmdBytes {
		if b != 0x0 {
			cmd = append(cmd, b)
		}
	}
	return fmt.Sprintf("%s", cmd)
}

func CmdToBytes(cmd string, commandLength uint) []byte {
	var cmdBytes []byte = make([]byte, commandLength)
	for i, c := range cmd {
		cmdBytes[i] = byte(c)
	}
	return cmdBytes
}
