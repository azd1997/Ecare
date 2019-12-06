package eaccount

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/azd1997/Ecare/ecoin/account"
	"github.com/azd1997/Ecare/ecoin/utils"
)

// EAccounts 系统账户表 存储每个账户的可公开的信息，包括余额、角色（角色中定义了其币相关的规则）、公钥、可用状态。键值为UserID.ID
// TODO: 后期将之改为状态树或者叫账户树。因为现在这么做如果账户很多其实占用很大。
type EAccounts struct {
	Map map[string]*EAccount 	`json:"eAccounts"`
	Lock sync.RWMutex	// 读写锁，尽管每个共识节点都维护自己的EAccounts集合，但由于各自都是并发多协程的，所以需要进行并发保护
}

// todo: 注意： 这张表没有历史记录，也就意味着无法回滚状态。

// TODO: 由于使用了RWMutex(内含不导出变量)，所以EAccounts没办法直接编码解码。
// TODO： 初步考虑，将这张表放到键值对数据库或者是文档数据库，当然关系型也可以，每次重新加载时将其加载到内存（或不加载到内存）
// TODO：优化查询。这里会是很重要的瓶颈，因为几乎所有操作都需要从这取信息
// TODO: 如果账号数量太大，不适合放在内存，则考虑KV键值对数据库，然后将K存到ElasticSearch便于搜索
// 但是由于EAccount可能存有许多字段，使用结构化的Mysql也不错

// TODO: 这里的存储与加载部分待DB模块实现之后进行修改

// TODO: 直接使用json/gob等编码解码，由于RWMutex内部都是不导出整型数字，所以会被还原成0出来。但在这里其实是可以的，因为文件加载eaccounts
//  只存在于程序初启动时，运行中不会从文件读取。

// CreateEcoinAccountsFromJsonFile 从文件创建新的自己账户表(用于在还没有EcoinAccounts时的创建)
func CreateEcoinAccountsFromJsonFile(file string) (*EAccounts, error) {
	accounts := EAccounts{}
	accounts.Map = make(map[string]*EAccount)

	err := accounts.LoadFileWithJsonUnmarshal(file)
	if err != nil {
		return nil, utils.WrapError("CreateEAccounts", err)
	}
	return &accounts, nil
}



/*******************************************************实现接口*********************************************************/

// AddAccount 向账户表添加新账户
func (as *EAccounts) AddEAccount(new *EAccount) (err error) {

	// TODO: 账户的检验

	as.Map[new.UserId.Id] = new
	return nil
}

// DelEAccount 根据用户Id删除EAccount
func (as *EAccounts) DelEAccount(id string) error {
	delete(as.Map, id)
	return nil
}

// UpdateEAccount 更新某一个账户的状态。为了追求通用，直接传入一个保持该Id不变的新EAccount来进行更新
func (as *EAccounts) UpdateEAccount(new *EAccount) error {
	as.Map[new.UserId.Id] = new
	return nil
}

// GetEAccount 根据用户id查看公开账户
func (as *EAccounts) GetEAccount(id string) *EAccount {
	if v, ok := as.Map[id]; ok {
		return v
	}
	return nil
}

// GetAllUserID 获取所有账户的对应的UserID
func (as *EAccounts) GetAllUserID() (userIds []account.UserId) {
	for userId := range as.Map {
		userIds = append(userIds, account.UserId{
			Id:     userId,
			RoleNo: uint(as.Map[userId].Role().No()),
		})
	}
	return userIds
}

// SaveFileWithGobEncode 将内存维护的自己账户表写入本地指定路径下的文件
func (as *EAccounts) SaveFileWithGobEncode(file string) (err error) {
	if err = utils.SaveFileWithGobEncode(file, as); err != nil {
		return utils.WrapError("EAccounts_SaveFile", err)
	}
	return nil
}

// SaveFileWithJsonMarshal 将内存维护的自己账户表写入本地指定路径下的文件
func (as *EAccounts) SaveFileWithJsonMarshal(file string) (err error) {
	if err = utils.SaveFileWithJsonMarshal(file, as); err != nil {
		return utils.WrapError("EAccounts_SaveFile", err)
	}
	return nil
}

// LoadFileWithGobDecode 从本地文件中读取自己账户表（用于加载）
func (as *EAccounts) LoadFileWithGobDecode(file string) (err error) {
	if _, err = os.Stat(file); os.IsNotExist(err) {
		return utils.WrapError("EAccounts_LoadFile", err)
	}

	var accounts EAccounts

	fileContent, err := ioutil.ReadFile(file)
	if err != nil {
		return utils.WrapError("EAccounts_LoadFile", err)
	}

	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	if err = decoder.Decode(&accounts); err != nil {
		return utils.WrapError("EAccounts_LoadFile", err)
	}

	as.Map = accounts.Map
	return nil
}

// LoadFileWithJsonUnmarshal 从本地文件中读取自己账户表（用于加载）
func (as *EAccounts) LoadFileWithJsonUnmarshal(file string) (err error) {
	if _, err = os.Stat(file); os.IsNotExist(err) {
		return utils.WrapError("EAccounts_LoadFile", err)
	}

	accounts := EAccounts{}

	fileContent, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println("1111111111")
		return utils.WrapError("EAccounts_LoadFile", err)
	}

	if err = json.Unmarshal(fileContent, &accounts); err != nil {
		fmt.Println("2222222222")
		return utils.WrapError("EAccounts_LoadFile", err)
	}

	as.Map = accounts.Map

	return nil
}

/*******************************************************实现接口*********************************************************/