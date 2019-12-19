package net

const (
	CmdUnknown uint32 = iota
	CmdPing

)

var CmdMap = map[uint32]string{
	CmdUnknown: "unknown",
}
