package p2ptcp

const PROTOCOL = "tcp"

const (
	Push byte = '1'
	Pull byte = '0'
)

const (
	UnknownMsg = iota
	PingMsg
	PongMsg
	RegisterMsg
	TxMsg
	BlockMsg
	GetBlockMsg
	PotMsg

)

var MsgType = map[uint32]string{
	UnknownMsg: "unknown",
	PingMsg: "ping",
	PongMsg: "pong",
	RegisterMsg:"register",
	TxMsg: "tx",
	BlockMsg:"block",
	GetBlockMsg:"getblock",
	PotMsg:"pot",

}
