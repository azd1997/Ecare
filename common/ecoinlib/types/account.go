package types

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"github.com/azd1997/Ecare/common/ecoinlib/utils"
	"github.com/mr-tron/base58"
	"golang.org/x/crypto/ripemd160"
	"io/ioutil"
	"os"
)

// 采取与比特币基本一致的账户生成

/*********************************************************************************************************************
                                                    UserID相关
*********************************************************************************************************************/

// UserID 用户身份标识符，包含标识符和角色编号两个属性
type UserID struct {
	ID     string `json:"id"`
	RoleNo uint   `json:"roleNo"` // 角色编号，参见role.go
}

// String 转换为json字符串
func (userID *UserID) String() string {
	return utils.JsonMarshalIndent(userID)
}

// IsValid 判断UserID.ID是否有效
func (userID *UserID) IsValid(checksumLength uint) (bool, error) {
	fullPubKeyHash, err := base58.Decode(userID.ID)
	if err != nil {
		return false, fmt.Errorf("UserID_IsValid: base58_Decode: %s", err)
	}
	length := uint(len(fullPubKeyHash))
	actualChecksum := fullPubKeyHash[length-checksumLength:]
	version := fullPubKeyHash[0]
	pubKeyHash := fullPubKeyHash[1 : length-checksumLength]
	targetChecksum := checksum(append([]byte{version}, pubKeyHash...), checksumLength)
	return bytes.Compare(actualChecksum, targetChecksum) == 0, nil
}

/*********************************************************************************************************************
                                                    Account相关
*********************************************************************************************************************/

// Account 账户，包含私钥和公钥，标志唯一身份。UserID是外部可见的标志
type Account struct {
	PrivKey ecdsa.PrivateKey
	PubKey  []byte
	RoleNo  uint
}

// newKeyPair 创造新的公私钥对
func newKeyPair() (ecdsa.PrivateKey, []byte, error) {
	// 椭圆曲线
	curve := elliptic.P256()
	// 生成私钥
	privKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return ecdsa.PrivateKey{}, nil, fmt.Errorf("NewKeyPair: %s", err)
	}
	pubKey := append(privKey.PublicKey.X.Bytes(), privKey.PublicKey.Y.Bytes()...)

	return *privKey, pubKey, nil
}

// NewAccount 新建账户
// TODO: 注意：新建账户时需添加入本地gsm.accounts并向外广播.
func NewAccount(roleNo uint) (*Account, error) {
	privKey, pubKey, err := newKeyPair()
	if err != nil {
		return nil, fmt.Errorf("NewAccount: %s", err)
	}
	return &Account{
		PrivKey: privKey,
		PubKey:  pubKey,
		RoleNo:  roleNo,
	}, nil
}

// pubKeyHash publicKey -> sha256 -> publicKeyHash -> ripemd160 -> publicKeyHashRipemd160
func pubKeyHash(pubKey []byte) ([]byte, error) {
	pubHash := sha256.Sum256(pubKey)

	hasher := ripemd160.New()
	if _, err := hasher.Write(pubHash[:]); err != nil {
		return nil, fmt.Errorf("PubKeyHash: %s", err)
	}
	return hasher.Sum(nil), nil
}

// checksum publicKeyHashRipemd160 -> sha256 -> sha256 -> [:checksumLength] -> checksum
func checksum(payload []byte, checksumLength uint) []byte {
	firstHash := sha256.Sum256(payload)
	secondHash := sha256.Sum256(firstHash[:])
	return secondHash[:checksumLength]
}

// UserID publicKeyHashRipemd160 + checksum + version -> base58 -> userID
func (a *Account) UserID(checksumLength uint, version byte) (UserID, error) {
	aCopy := a
	pubHash, err := pubKeyHash(aCopy.PubKey)
	if err != nil {
		return UserID{}, fmt.Errorf("Account_UserID: %s", err)
	}
	versionedHash := append([]byte{version}, pubHash...)
	checksum := checksum(versionedHash, checksumLength)
	fullHash := append(versionedHash, checksum...)

	userId := base58.Encode(fullHash)
	return UserID{userId, aCopy.RoleNo}, nil
}

// Sign 使用该账号对目标数据作签名。目标数据只能是基础类型、结构体、切片、表等，必须提前转为[]byte
func (a *Account) Sign(target []byte) (sig Signature, err error) {
	//r, s, err := ecdsa.Sign(rand.Reader, &a.PrivKey, target)
	//if err != nil {
	//	return nil, fmt.Errorf("Account_Sign: %s", err)
	//}
	//signature := append(r.Bytes(), s.Bytes()...)
	//return signature, nil
	return utils.Sign(target, &a.PrivKey)
}

//// NewTX 该账户作为主体，构造新交易
//func (a *Account) NewTX(typ uint, args ArgsOfNewTX) (tx TX, err error) {
//	// TODO: 根据账户类型不同来处理
//	return newTransaction(typ, args)
//}
// TODO: 备注： 已经交给gsm去做，account不负责

/*********************************************************************************************************************
                                                    SelfAccounts相关
*********************************************************************************************************************/

// SelfAccounts 自己的账户集合
type SelfAccounts struct {
	Map map[string]*Account
}

// SaveFile 将内存维护的自己账户表写入本地指定路径下的文件
func (as *SelfAccounts) SaveFile(port int, accountFilePathTemp string) (err error) {
	var content bytes.Buffer
	file := fmt.Sprintf(accountFilePathTemp, string(port))

	gob.Register(elliptic.P256())

	encoder := gob.NewEncoder(&content)
	if err = encoder.Encode(as); err != nil {
		return fmt.Errorf("SelfAccounts_SaveFile: Encode: %s", err)
	}
	if err = ioutil.WriteFile(file, content.Bytes(), 0644); err != nil {
		return fmt.Errorf("SelfAccounts_SaveFile: WriteFile: %s", err)
	}
	return nil
}

// LoadFile 从本地文件中读取自己账户表（用于加载）
func (as *SelfAccounts) LoadFile(port int, accountFilePathTemp string) (err error) {
	file := fmt.Sprintf(accountFilePathTemp, string(port))
	if _, err = os.Stat(file); os.IsNotExist(err) {
		return fmt.Errorf("SelfAccounts_LoadFile: os_Stat: %s", err)
	}

	var accounts SelfAccounts

	fileContent, err := ioutil.ReadFile(file)
	if err != nil {
		return fmt.Errorf("SelfAccounts_LoadFile: ReadFile: %s", err)
	}

	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	if err = decoder.Decode(&accounts); err != nil {
		return fmt.Errorf("SelfAccounts_LoadFile: gob_Decode: %s", err)
	}

	as.Map = accounts.Map
	return nil
}

// CreateSelfAccounts 从文件创建新的自己账户表(用于在还没有SelfAccounts时的创建)
func CreateSelfAccounts(port int, accountFilePathTemp string) (*SelfAccounts, error) {
	accounts := SelfAccounts{}
	accounts.Map = make(map[string]*Account)

	err := accounts.LoadFile(port, accountFilePathTemp)
	if err != nil {
		return nil, fmt.Errorf("CreateSelfAccounts: %s", err)
	}
	return &accounts, nil
}

// GetAccount 根据用户id查看自己的账户
func (as *SelfAccounts) GetAccount(userID string) Account {
	return *as.Map[userID]
}

// GetAllUserID 获取自己所有账户的对应的UserID
func (as *SelfAccounts) GetAllUserID() (userIDs []UserID) {
	for userID := range as.Map {
		userIDs = append(userIDs, UserID{
			ID:     userID,
			RoleNo: as.Map[userID].RoleNo,
		})
	}
	return userIDs
}

// AddAccount 向自己账户表添加新账户
func (as *SelfAccounts) AddAccount(roleNo uint, checksumLength uint, version byte) (userID UserID, err error) {
	account, err := NewAccount(roleNo)
	if err != nil {
		return UserID{}, fmt.Errorf("Accounts_AddAccount: %s", err)
	}
	userID, err = account.UserID(checksumLength, version)
	if err != nil {
		return UserID{}, fmt.Errorf("Accounts_AddAccount: %s", err)
	}
	as.Map[userID.ID] = account
	return userID, err
}

/*********************************************************************************************************************
                                                    EcoinAccount相关
*********************************************************************************************************************/

// EcoinAccount 每个区块链账户的公开信息。
// account.go中Account指完全权限的账户信息，含有私钥信息，又因为全局状态机中会维护这些公私钥以外的信息，所以Account只有公私钥信息
type EcoinAccount struct {
	UserID    UserID
	pubKey    []byte
	balance   Coin
	role      Role	// 存储角色信息，和UserID中都存了一次RoleNo。必须保证相同
	available bool
}

// PubKey 获取账户公钥
func (a *EcoinAccount) PubKey() []byte {
	return a.pubKey
}

// Balance 获取余额
func (a *EcoinAccount) Balance() Coin {
	return a.balance
}

// Role 获取账户的角色
func (a *EcoinAccount) Role() *Role {
	return &a.role
}

// Available 账户是否可用
func (a *EcoinAccount) Available() bool {
	return a.available
}

/*********************************************************************************************************************
                                                    EcoinAccountMap相关
*********************************************************************************************************************/

// EcoinAccountsMap 系统账户表 存储每个账户的可公开的信息，包括余额、角色（角色中定义了其币相关的规则）、公钥、可用状态。键值为UserID.Id
// TODO: 后期将之改为状态树或者叫账户树。因为现在这么做如果账户很多其实占用很大。
type EcoinAccounts struct {
	Map map[string]*EcoinAccount
}

// todo: 注意： 这张表没有历史记录，也就意味着无法回滚状态。

// SaveFile 将内存维护的自己账户表写入本地指定路径下的文件
func (as *EcoinAccounts) SaveFile(port int, accountFilePathTemp string) (err error) {
	var content bytes.Buffer
	file := fmt.Sprintf(accountFilePathTemp, string(port))

	gob.Register(elliptic.P256())

	encoder := gob.NewEncoder(&content)
	if err = encoder.Encode(as); err != nil {
		return fmt.Errorf("EcoinAccounts_SaveFile: Encode: %s", err)
	}
	if err = ioutil.WriteFile(file, content.Bytes(), 0644); err != nil {
		return fmt.Errorf("EcoinAccounts_SaveFile: WriteFile: %s", err)
	}
	return nil
}

// LoadFile 从本地文件中读取自己账户表（用于加载）
func (as *EcoinAccounts) LoadFile(port int, accountFilePathTemp string) (err error) {
	file := fmt.Sprintf(accountFilePathTemp, string(port))
	if _, err = os.Stat(file); os.IsNotExist(err) {
		return fmt.Errorf("EcoinAccounts_LoadFile: os_Stat: %s", err)
	}

	accounts := EcoinAccounts{}

	fileContent, err := ioutil.ReadFile(file)
	if err != nil {
		return fmt.Errorf("EcoinAccounts_LoadFile: ReadFile: %s", err)
	}

	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	if err = decoder.Decode(&accounts); err != nil {
		return fmt.Errorf("EcoinAccounts_LoadFile: gob_Decode: %s", err)
	}

	as.Map = accounts.Map

	return nil
}

// CreateEcoinAccounts 从文件创建新的自己账户表(用于在还没有EcoinAccounts时的创建)
func CreateEcoinAccounts(port int, accountFilePathTemp string) (*EcoinAccounts, error) {
	accounts := EcoinAccounts{}
	accounts.Map = make(map[string]*EcoinAccount)

	err := accounts.LoadFile(port, accountFilePathTemp)
	if err != nil {
		return nil, fmt.Errorf("CreateEcoinAccounts: %s", err)
	}
	return &accounts, nil
}

// GetAccount 根据用户id查看公开账户
func (as *EcoinAccounts) GetAccount(userID string) EcoinAccount {
	return *as.Map[userID]
}

// GetAllUserID 获取所有账户的对应的UserID
func (as *EcoinAccounts) GetAllUserID() (userIDs []UserID) {
	for userID := range as.Map {
		userIDs = append(userIDs, UserID{
			ID:     userID,
			RoleNo: uint(as.Map[userID].Role().No()),
		})
	}
	return userIDs
}

// AddAccount 向账户表添加新账户
func (as *EcoinAccounts) AddAccount(newEcoinAccount *EcoinAccount) (userID UserID, err error) {

	// TODO: 账户的检验

	as.Map[newEcoinAccount.UserID.ID] = newEcoinAccount
	return userID, err
}