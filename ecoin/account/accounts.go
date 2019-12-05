package account

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/azd1997/Ecare/ecoin/utils"
)

// Accounts 自己的账户集合
type Accounts struct {
	Map map[string]*Account	`json:"selfAccounts"`
}

// SaveFileWithGobEncode 将内存维护的自己账户表写入本地指定路径下的文件
func (as *Accounts) SaveFileWithGobEncode(file string) (err error) {
	if err = utils.SaveFileWithGobEncode(file, as); err != nil {
		return utils.WrapError("Accounts_SaveFile", err)
	}
	return nil
}

// SaveFileWithJsonMarshal 将内存维护的自己账户表写入本地指定路径下的文件
func (as *Accounts) SaveFileWithJsonMarshal(file string) (err error) {
	if err = utils.SaveFileWithJsonMarshal(file, as); err != nil {
		return utils.WrapError("Accounts_SaveFile", err)
	}
	return nil
}

// LoadFileWithGobDecode 从本地文件中读取自己账户表（用于加载）
func (as *Accounts) LoadFileWithGobDecode(file string) (err error) {
	if _, err = os.Stat(file); os.IsNotExist(err) {
		return utils.WrapError("Accounts_LoadFile", err)
	}

	var accounts Accounts

	fileContent, err := ioutil.ReadFile(file)
	if err != nil {
		return utils.WrapError("Accounts_LoadFile", err)
	}

	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	if err = decoder.Decode(&accounts); err != nil {
		return utils.WrapError("Accounts_LoadFile", err)
	}

	as.Map = accounts.Map
	return nil
}

// LoadFileWithJsonUnmarshal 从本地文件中读取自己账户表（用于加载）
func (as *Accounts) LoadFileWithJsonUnmarshal(file string) (err error) {
	if _, err = os.Stat(file); os.IsNotExist(err) {
		return utils.WrapError("Accounts_LoadFile", err)
	}

	var accounts Accounts

	fileContent, err := ioutil.ReadFile(file)
	if err != nil {
		return utils.WrapError("Accounts_LoadFile", err)
	}

	if err = json.Unmarshal(fileContent, &accounts); err != nil {
		return utils.WrapError("Accounts_LoadFile", err)
	}

	as.Map = accounts.Map
	return nil
}

// CreateAccounts 从文件创建新的自己账户表(用于在还没有SelfAccounts时的创建)
func CreateAccountsFromJsonFile(file string) (*Accounts, error) {
	accounts := Accounts{}
	accounts.Map = make(map[string]*Account)

	err := accounts.LoadFileWithJsonUnmarshal(file)
	if err != nil {
		return nil, utils.WrapError("CreateAccounts", err)
	}
	return &accounts, nil
}

// GetAccount 根据用户id查看自己的账户
func (as *Accounts) GetAccount(userID string) Account {
	return *as.Map[userID]
}

// GetAllUserID 获取自己所有账户的对应的UserID
func (as *Accounts) GetAllUserId() (userIds []UserId) {
	for userId := range as.Map {
		userIds = append(userIds, UserId{
			Id:     userId,
			RoleNo: as.Map[userId].RoleNo,
		})
	}
	return userIds
}

// AddAccount 向自己账户表添加新账户
func (as *Accounts) AddAccount(roleNo uint) (userId UserId, err error) {
	account, err := NewAccount(roleNo)
	if err != nil {
		return UserId{}, utils.WrapError("Accounts_AddAccount", err)
	}
	userId, err = account.UserId()
	if err != nil {
		return UserId{}, utils.WrapError("Accounts_AddAccount", err)
	}
	as.Map[userId.Id] = account
	return userId, err
}
