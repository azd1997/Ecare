package common

const NODE_VERSION = 1

type PingMsg struct {
	AddrFrom string
}

type PongMsg struct {
	AddrFrom string
}

type GetAddrsMsg struct {
	AddrFrom string
}

type AddrsMsg struct {
	AddrFrom string
	LocalAddrs []string
}

type VersionMsg struct {
	NodeVer uint8
	ChainHeight int
	AddrsFrom string
}


