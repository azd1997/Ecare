package net

const (
	CmdUnknown byte = iota
	CmdPing
	CmdGetAddrs
	CmdAddrs
	CmdVersion
	CmdGetInventory
	CmdInventory
	CmdBlock
	CmdGetBlocks

	CmdTx

)

var CmdMap = map[byte]string{
	CmdUnknown: "unknown",
}
