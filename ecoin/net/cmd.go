package net

const (
	CmdUnknown uint8 = iota
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

var CmdMap = map[uint8]string{
	CmdUnknown: "unknown",
}
