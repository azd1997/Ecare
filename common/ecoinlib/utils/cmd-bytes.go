package utils

import "fmt"

func BytesToCmd(cmdBytes []byte) string {
	var cmd []byte
	for _, b := range cmdBytes {
		if b != 0x0 {
			cmd = append(cmd, b)
		}
	}
	return fmt.Sprintf("%s", cmd)
}

func CmdToBytes(cmd string, commandLength int) []byte {
	var cmdBytes []byte = make([]byte, commandLength)
	for i, c := range cmd {
		cmdBytes[i] = byte(c)
	}
	return cmdBytes
}
