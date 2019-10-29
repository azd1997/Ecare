package ecoinlib

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"os"
)

type Accounts struct {
	Map map[UserID]*Account
}

func (as *Accounts) SaveFile(port int, accountFilePathTemp string) (err error) {
	var content bytes.Buffer
	file := fmt.Sprintf(accountFilePathTemp, string(port))

	gob.Register(elliptic.P256())

	encoder := gob.NewEncoder(&content)
	if err = encoder.Encode(as); err != nil {
		return fmt.Errorf("Accounts_SaveFile: Encode: %s", err)
	}
	if err = ioutil.WriteFile(file, content.Bytes(), 0644); err != nil {
		return fmt.Errorf("Accounts_SaveFile: WriteFile: %s", err)
	}
	return nil
}

func (as *Accounts) LoadFile(port int, accountFilePathTemp string) (err error) {
	file := fmt.Sprintf(accountFilePathTemp, string(port))
	if _, err = os.Stat(file); os.IsNotExist(err) {
		return fmt.Errorf("Accounts_LoadFile: os_Stat: %s", err)
	}

	var accounts Accounts

	fileContent, err := ioutil.ReadFile(file)
	if err != nil {
		return fmt.Errorf("Accounts_LoadFile: ReadFile: %s", err)
	}

	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	if err = decoder.Decode(&accounts); err != nil {
		return fmt.Errorf("Accounts_LoadFile: gob_Decode: %s", err)
	}

	as.Map = accounts.Map
	return nil
}

func CreateAccounts(port int, accountFilePathTemp string) (*Accounts, error) {
	accounts := Accounts{}
	accounts.Map = make(map[UserID]*Account)

	err := accounts.LoadFile(port, accountFilePathTemp)
	if err != nil {
		return nil, fmt.Errorf("CreateAccounts: %s", err)
	}
	return &accounts, nil
}

func (as *Accounts) GetAccount(userID UserID) Account {
	return *as.Map[userID]
}

func (as *Accounts) GetAllUserID() (userIDS []UserID) {
	for userID := range as.Map {
		userIDS = append(userIDS, userID)
	}
	return userIDS
}

func (as *Accounts) AddAccount(checksumLength int, version byte) (userID UserID, err error) {
	account, err := NewAccount()
	if err != nil {
		return "", fmt.Errorf("Accounts_AddAccount: %s", err)
	}
	userID, err = account.UserID(checksumLength, version)
	if err != nil {
		return "", fmt.Errorf("Accounts_AddAccount: %s", err)
	}
	as.Map[userID] = account
	return userID, err
}
