package types

import (
	"fmt"
	"strconv"
)

// Option 运行区块链节点所需的所有配置项，包括一些常用配置项和许多不常用配置项
type Option struct {
	ChainOption
	LoggerOption
	NetworkOption
	AccountOption
}

// DefaultOption 获取默认配置
func DefaultOption() *Option {
	return &Option{
		ChainOption:   ChainOption{
			//GenesisMsg:fmt.Sprintf("[%s] | someone created the blockchain", time.Now().Format("2006/01/02 15:04:05")),
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
			userID:UserID{},
			checksumLength:4,
			accountVersion:byte(0x00),
			accountFilePathTemp:"./tmp/accounts/accounts_%s.data",
			ecoinAccountFilePathTemp:"./tmp/accounts/ecoinAccounts_%s.data",
		},
	}
}

// ChainOption 和区块链数协议有关的配置项
type ChainOption struct {
	//GenesisMsg string TODO: 好像不需要这项 genesismsg实际是第一个区块的coinbasetx，所以genesismsg当做常规的coinbase description就好
	dbPathTemp string
	uctxpFilePathTemp string
}

// DbPathTemp 获取默认的区块链数据库路径模板
func (op *Option) DbPathTemp() string {
	return op.dbPathTemp
}

// SetDbPathTemp 设置区块链数据库路径模板
func (op *Option) SetDbPathTemp(newDbPathTemp string) *Option {
	op.dbPathTemp = newDbPathTemp
	return op
}

// DbPath 获取区块链数据库路径
func (op *Option) DbPath() string {
	return fmt.Sprintf(op.dbPathTemp, op.port)
}

// UctxpFilePathTemp 获取uctxp文件存储路径模板
func (op *Option) UctxpFilePathTemp() string {
	return op.uctxpFilePathTemp
}

// SetUctxpFilePathTemp 获取uctxp文件存储路径模板
func (op *Option) SetUctxpFilePathTemp(newTemp string) *Option {
	op.uctxpFilePathTemp = newTemp
	return op
}

// UctxpFilePath 获取uctxp文件路径
func (op *Option) UctxpFilePath() string {
	return fmt.Sprintf(op.uctxpFilePathTemp, op.port)
}

// LoggerOption 日志记录器配置项
type LoggerOption struct {

}

// NetworkOption 区块链网络配置项
type NetworkOption struct {
	commandLength uint
	nodeVersion int
	protocol	string
	ipv4 string
	port string
	maxConnNum int
	addrsFilePathTemp string
}

// CommandLength 网络通信中命令字节长度，区块链一旦开始运行，不可修改，若修改无法通信
func (op *Option) CommandLength() uint {
	return op.commandLength
}

// NodeVersion 节点版本，也就是软件版本
func (op *Option) NodeVersion() int {
	return op.nodeVersion
}

// Protocol 传输协议，目前只支持tcp。
func (op *Option) Protocol() string {
	return op.protocol
}

// Ipv4 获取ipv4地址
func (op *Option) Ipv4() string {
	return op.ipv4
}

// SetIpv4 设置Ipv4
func (op *Option) SetIpv4(newIpv4 string) *Option {
	// TODO: 检查ipv4格式

	op.ipv4 = newIpv4
	return op
}

// Port 通信端口
func (op *Option) Port() int {
	i, _ := strconv.Atoi(op.port)
	return i
}

// SetPort 设置通信端口
func (op *Option) SetPort(newPort int) *Option {
	op.port = strconv.Itoa(newPort)
	return op
}

// NodeAddress 根据ipv4地址和端口构建本机节点地址
func (op *Option) NodeAddress() Address {
	return Address(fmt.Sprintf("%s:%s", op.ipv4, op.port))
}

// MaxConnNum 最大并发连接数
func (op *Option) MaxConnNum() int {
	return op.maxConnNum
}

// SetMaxConnNum 设置最大并发连接数
func (op *Option) SetMaxConnNum(newNum uint) *Option {
	op.maxConnNum = int(newNum)
	return op
}

// AddrsFilePathTemp 获取addrs文件存储路径模板
func (op *Option) AddrsFilePathTemp() string {
	return op.addrsFilePathTemp
}

// SetAddrsFilePathTemp 获取addrs文件存储路径模板
func (op *Option) SetAddrsFilePathTemp(newTemp string) *Option {
	op.addrsFilePathTemp = newTemp
	return op
}

// AddrsFilePath 获取Addrs文件路径
func (op *Option) AddrsFilePath() string {
	return fmt.Sprintf(op.addrsFilePathTemp, op.port)
}

// AccountOption 账户相关的配置
type AccountOption struct {
	userID UserID	// 运行节点的账户地址，必需项
	checksumLength uint
	accountVersion byte	// uint8 version = 0
	accountFilePathTemp	string	// "./tmp/accounts/accounts_%s.data"
	ecoinAccountFilePathTemp string // "./tmp/accounts/ecoinAccounts_%s.data"
}

// UserID 获取当前运行时的账号UserID
func (op *Option) UserID() UserID {
	return op.userID
}

// SetUserID 设置运行时的账户UserID
func (op *Option) SetUserID(newUserID UserID) *Option {
	op.userID = newUserID
	return op
}

// ChecksumLength 校验码长度，默认为
func (op *Option) ChecksumLength() uint {
	return op.checksumLength
}

// Version 账户版本号，默认为byte(0x00)
func (op *Option) Version() byte {
	return op.accountVersion
}

// AccountFilePathTemp 获取账户文件存储路径模板
func (op *Option) AccountFilePathTemp() string {
	return op.accountFilePathTemp
}

// SetAccountFilePathTemp 获取账户文件存储路径模板
func (op *Option) SetAccountFilePathTemp(newTemp string) *Option {
	op.accountFilePathTemp = newTemp
	return op
}

// EcoinAccountFilePathTemp 获取账户文件存储路径模板
func (op *Option) EcoinAccountFilePathTemp() string {
	return op.ecoinAccountFilePathTemp
}

// SetEcoinAccountFilePathTemp 获取账户文件存储路径模板
func (op *Option) SetEcoinAccountFilePathTemp(newTemp string) *Option {
	op.ecoinAccountFilePathTemp = newTemp
	return op
}


