package ecoinlib

import (
	"fmt"
	"strconv"
	"time"
)

type Option struct {
	ChainOption
	LoggerOption
	NetworkOption
	AccountOption
}

func DefaultOption() *Option {
	return &Option{
		ChainOption:   ChainOption{
			GenesisMsg:fmt.Sprintf("[%s] | someone created the blockchain", time.Now().Format("2006/01/02 15:04:05")),
			dbPathTemp:"./tmp/blocks/blocks_%s",
		},
		LoggerOption:  LoggerOption{},
		NetworkOption:NetworkOption{
			commandLength: 12,
			nodeVersion:   1,
			protocol:      "tcp",
			ipv4:"localhost",
			port:"3997",
			maxConnNum:100,
		},
		AccountOption: AccountOption{
			userID:"",
			checksumLength:4,
			accountVersion:byte(0x00),
			accountFilePathTemp:"./tmp/accounts/accounts_%s.data",
		},
	}
}

type ChainOption struct {
	GenesisMsg string
	dbPathTemp string
}

func (op *ChainOption) DbPathTemp() string {
	return op.dbPathTemp
}

type LoggerOption struct {

}

type NetworkOption struct {
	commandLength int
	nodeVersion int
	protocol	string
	ipv4 string
	port string
	maxConnNum int
}

func (op *NetworkOption) Port() int {
	i, _ := strconv.Atoi(op.port)
	return i
}



type AccountOption struct {
	userID UserID	// 运行节点的账户地址，必需项
	checksumLength int
	accountVersion byte	// uint8 version = 0
	accountFilePathTemp	string	// "./tmp/accounts/accounts_%s.data"
}

func (op *AccountOption) ChecksumLength() int {
	return op.checksumLength
}

func (op *AccountOption) Version() byte {
	return op.accountVersion
}

func (op *AccountOption) UserID() UserID {
	return op.userID
}

func (op *AccountOption) SetUserID(newUserID UserID) {
	op.userID = newUserID
}

func (op *AccountOption) AccountFilePathTemp() string {
	return op.accountFilePathTemp
}

