package net

const (
	CmdUnknown byte = iota
	CmdPing
	CmdPong
	CmdGetAddrs
	CmdAddrs
	CmdInv
	CmdVersion
	CmdBlock
	CmdGetBlocks
	CmdTx

)

var CmdMap = map[byte]string{
	CmdUnknown: "unknown",
}
