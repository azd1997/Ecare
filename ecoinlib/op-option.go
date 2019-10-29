package ecoin

import (
	"fmt"
	"strconv"
	"strings"
)

// TODO: 将所有不常更改项改为常量定义，降低各个类之间的耦合

// Option 运行区块链节点所需的所有配置项，包括一些常用配置项和许多不常用配置项
type Option struct {
	ChainOption
	LoggerOption
	NetworkOption
	AccountOption
	StorageOption // 特指具体的生产数据的存储相关的设置
	RegisterInfo // 注册信息
}

// DefaultOption 获取默认配置
func DefaultOption() *Option {
	return &Option{
		ChainOption:   ChainOption{
			//GenesisMsg:fmt.Sprintf("[%s] | someone created the blockchain", time.Now().Format("2006/01/02 15:04:05")),
			//dbPathTemp:"./tmp/blocks/blocks_%s",
		},
		LoggerOption:  LoggerOption{},
		NetworkOption:NetworkOption{
			//commandLength: 12,
			//nodeVersion:   1,
			//protocol:      "tcp",
			ipv4:"localhost",
			port:"3997",
			seedNode: "127.0.0.1:8888",
			//maxConnNum:100,
		},
		AccountOption: AccountOption{
			account:Account{},
			userID:UserID{},
			//checksumLength:4,
			//accountVersion:byte(0x00),
			//accountFilePathTemp:"./tmp/accounts/accounts_%s.data",
			//ecoinAccountFilePathTemp:"./tmp/accounts/ecoinAccounts_%s.data",
		},
		StorageOption:StorageOption{
			brokerAddr:"127.0.0.1:7777",
		},
		RegisterInfo:RegisterInfo{},
	}
}

func (op *Option) String() string {
	return strings.Join([]string{
		fmt.Sprintf("nodeAddress: %s", op.NodeAddress().String()),
		fmt.Sprintf("account_privKey: %s", op.Account().PrivKey),
		fmt.Sprintf("account_pubKey: %s", op.Account().PubKey),
		fmt.Sprintf("account_userID: %s", op.UserID().String()),
	}, "\n")
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
	seedNode string
	commandLength uint
	nodeVersion int
	protocol	string
	ipv4 string
	port string
	maxConnNum int
	addrsFilePathTemp string
}

// SeedNode 获取种子节点。种子节点是为了给初上线没有节点列表的情况用的
func (op *Option) SeedNode() string {
	return op.seedNode
}

// SetSeedNode 设置种子节点
func (op *Option) SetSeedNode(seed string) *Option {
	op.seedNode = seed
	return op
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
func (op *Option) Port() uint {
	i, _ := strconv.Atoi(op.port)
	return uint(i)
}

// SetPort 设置通信端口
func (op *Option) SetPort(newPort int) *Option {
	op.port = strconv.Itoa(newPort)
	return op
}

// NodeAddress 根据ipv4地址和端口构建本机节点地址
// Opts.NodeAddress()是本机节点地址。
func (op *Option) NodeAddress() *Address {
	return &Address{
		Ipv4Port:fmt.Sprintf("%s:%s", op.ipv4, op.port),
		Alias:"",
		PingTime:0,
		Honest:true,
	}
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
	account Account // 运行节点的账户，外部传入
	userID UserID	// 运行节点的账户地址，必需项，由account生成
	checksumLength uint
	accountVersion byte	// uint8 version = 0
	accountFilePathTemp	string	// "./tmp/accounts/accounts_%s.data"
	ecoinAccountFilePathTemp string // "./tmp/accounts/ecoinAccounts_%s.data"
}

// Account 获取当前运行时的账号Account
func (op *Option) Account() *Account {
	return &op.account
}

// SetAccount 设置运行时的账户Account，并修改userid
func (op *Option) SetAccount(newAccount Account) *Option {
	op.account = newAccount
	userID, err := op.account.UserID()
	if err != nil {
		op.userID = UserID{}	// 如果设置了account，却发现userid为空值，可以不管
	}
	op.userID = userID
	return op
}

// UserID 获取当前运行时的账号UserID
func (op *Option) UserID() *UserID {
	return &op.userID
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

// StorageOption 生产数据存储的设置
type StorageOption struct {
	brokerAddr string
}

// BrokerAddr 获取
func (op *Option) BrokerAddr() string {
	return op.brokerAddr
}

// SetBrokerAddr 设置
func (op *Option) SetBrokerAddr(broker string) *Option {
	op.brokerAddr = broker
	return op
}

/*********************************************************************************************************************
                                                    RegisterInfo相关
*********************************************************************************************************************/

// RegisterInfo 注册信息
type RegisterInfo struct {
	NameField string	`json:"name"`
	PhoneField string 	`json:"phone"`
	InstitutionField string 	`json:"institution"`
}

func (op *Option) Name() string {
	return op.NameField
}

func (op *Option) SetName(name string) *Option {
	op.NameField = name
	return op
}

func (op *Option) Phone() string {
	return op.PhoneField
}

func (op *Option) SetPhone(phone string) *Option {
	op.PhoneField = phone
	return op
}

func (op *Option) Institution() string {
	return op.InstitutionField
}

func (op *Option) SetInstitution(institution string) *Option {
	op.InstitutionField = institution
	return op
}


