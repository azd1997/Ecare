package ecoin

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/azd1997/Ecare/ecoinlib/log"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/mr-tron/base58"
	"golang.org/x/crypto/ripemd160"
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

// newUserID 根据ID生成UserID
func newUserID(id string, gsm *GlobalStateMachine) *UserID {
	ea := gsm.Accounts.GetAccount(id)
	if ea != nil {	// 说明存在
		return &ea.UserID
	}
	return &UserID{
		ID:     id,
		RoleNo: 100,	// 100是默认的roleNo值，可以替代所有的B类角色
	}
}

// String 转换为json字符串
func (userID *UserID) String() string {
	return JsonMarshalIndentToString(userID)
}

// IsValid 判断UserID.ID是否有效
func (userID *UserID) IsValid() (bool, error) {
	fullPubKeyHash, err := base58.Decode(userID.ID)
	if err != nil {
		return false, WrapError("UserID_IsValid", err)
	}
	length := uint(len(fullPubKeyHash))
	actualChecksum := fullPubKeyHash[length-CHECKSUM_LENGTH:]
	version := fullPubKeyHash[0]
	pubKeyHash := fullPubKeyHash[1 : length-CHECKSUM_LENGTH]
	targetChecksum := checksum(append([]byte{version}, pubKeyHash...))
	return bytes.Compare(actualChecksum, targetChecksum) == 0, nil
}

//func (userID *UserID) IsValid(checksumLength uint) (bool, error) {
//	fullPubKeyHash, err := base58.Decode(userID.ID)
//	if err != nil {
//		return false, fmt.Errorf("UserID_IsValid: base58_Decode: %s", err)
//	}
//	length := uint(len(fullPubKeyHash))
//	actualChecksum := fullPubKeyHash[length-checksumLength:]
//	version := fullPubKeyHash[0]
//	pubKeyHash := fullPubKeyHash[1 : length-checksumLength]
//	targetChecksum := checksum(append([]byte{version}, pubKeyHash...), checksumLength)
//	return bytes.Compare(actualChecksum, targetChecksum) == 0, nil
//}

/*********************************************************************************************************************
                                                    Account相关
*********************************************************************************************************************/

// Account 账户，包含私钥和公钥，标志唯一身份。UserID是外部可见的标志
type Account struct {
	PrivKey ecdsa.PrivateKey	`json:"privKey"`
	PubKey  []byte	`json:"pubKey"`
	RoleNo  uint	`json:"roleNo"`
}

// loadOrCreateAccount 从指定路径加载账户，加载不到就新建
func loadOrCreateAccount(port uint) (account *Account, err error) {
	account = &Account{}
	// 1.1. 加载selfAccount文件，取账户文件
	accountFile := fmt.Sprintf(SELFACCOUNT_FILEPATH_TEMP, strconv.Itoa(int(port)))
	// 1.1.1 检查是否存在账户文件
	exists, err := FileExists(accountFile)
	if err != nil {
		return nil, err
	}
	// 若不存在，则需要创建一个账户并保存到这个文件
	if !exists {
		log.Warn("默认路径下找不到指定账户文件: %s", accountFile)
		log.Info("%s", "准备创建新账户......")
		account, err = NewAccount(0)
		if err != nil {
			return nil, err
		}
		err = account.SaveFileWithGobEncode(port)
		if err != nil {
			return nil, err
		}
		id, err := account.UserID()
		if err != nil {
			return nil, err
		}
		log.Success("新账户创建成功并保存至默认路径， 账户ID: %s", id.String())
	} else {
		// 若存在，则从这个文件读取account
		log.Info("%s", "默认路径下发现账户文件， 准备加载......")
		err = account.LoadFileWithGobDecode(port)
		if err != nil {
			return nil, err
		}
		id, err := account.UserID()
		if err != nil {
			return nil, err
		}
		log.Success("账户加载成功， 账户ID: %s", id.String())
	}

	return account, nil
}

// TODO: 待解决的问题：多个账户文件在同一个目录下怎么去选取。目前的做法是只读取指定文件名的账户文件。但如果要考虑多个账户呢？

// SaveFileWithGobEncode 保存到文件
func (a *Account) SaveFileWithGobEncode(port uint) (err error) {
	file := fmt.Sprintf(SELFACCOUNT_FILEPATH_TEMP, strconv.Itoa(int(port)))
	gob.Register(elliptic.P256())
	if err = saveFileWithGobEncode(file, a); err != nil {
		return WrapError("Account_SaveFile", err)
	}
	return nil
}

// LoadFileWithGobDecode 从本地文件中读取自己账户表（用于加载）
func (a *Account) LoadFileWithGobDecode(port uint) (err error) {
	file := fmt.Sprintf(SELFACCOUNT_FILEPATH_TEMP, strconv.Itoa(int(port)))
	if _, err = os.Stat(file); os.IsNotExist(err) {
		return WrapError("Account_LoadFile", err)
	}

	var account Account

	fileContent, err := ioutil.ReadFile(file)
	if err != nil {
		return WrapError("Account_LoadFile", err)
	}

	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	if err = decoder.Decode(&account); err != nil {
		return WrapError("Account_LoadFile", err)
	}

	a.PrivKey = account.PrivKey
	a.PubKey = account.PubKey
	a.RoleNo = account.RoleNo

	return nil
}

// String 打印字符串
func (a *Account) String() string {
	return strings.Join([]string{
		fmt.Sprintf("	PrivKey_D: %s", a.PrivKey.D),
		fmt.Sprintf("	PrivKey_X: %s", a.PrivKey.X),
		fmt.Sprintf("	PrivKey_Y: %s", a.PrivKey.Y),
		fmt.Sprintf("	PubKey: %s", base64.StdEncoding.EncodeToString(a.PubKey)),
		fmt.Sprintf("	RoleNo: %d", a.RoleNo),
	}, "\n")
}

// newKeyPair 创造新的公私钥对
func newKeyPair() (ecdsa.PrivateKey, []byte, error) {
	// 椭圆曲线
	curve := elliptic.P256()
	// 生成私钥
	privKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return ecdsa.PrivateKey{}, nil, WrapError("NewKeyPair", err)
	}
	pubKey := append(privKey.PublicKey.X.Bytes(), privKey.PublicKey.Y.Bytes()...)

	return *privKey, pubKey, nil
}

// NewAccount 新建账户
// TODO: 注意：新建账户时需添加入本地gsm.accounts并向外广播.
func NewAccount(roleNo uint) (*Account, error) {
	privKey, pubKey, err := newKeyPair()
	if err != nil {
		return nil, WrapError("NewAccount", err)
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
		return nil, WrapError("PubKeyHash", err)
	}
	return hasher.Sum(nil), nil
}

// checksum publicKeyHashRipemd160 -> sha256 -> sha256 -> [:checksumLength] -> checksum
func checksum(payload []byte) []byte {
	firstHash := sha256.Sum256(payload)
	secondHash := sha256.Sum256(firstHash[:])
	return secondHash[:CHECKSUM_LENGTH]
}

// UserID publicKeyHashRipemd160 + checksum + version -> base58 -> userID
func (a *Account) UserID() (UserID, error) {
	aCopy := a
	pubHash, err := pubKeyHash(aCopy.PubKey)
	if err != nil {
		return UserID{}, WrapError("Account_UserID", err)
	}
	versionedHash := append([]byte{ACCOUNT_VERSION}, pubHash...)
	checksum := checksum(versionedHash)
	fullHash := append(versionedHash, checksum...)

	userId := base58.Encode(fullHash)
	return UserID{userId, aCopy.RoleNo}, nil
}

// Sign 使用该账号对目标数据作签名。目标数据只能是基础类型、结构体、切片、表等，必须提前转为[]byte
func (a *Account) Sign(target []byte) (sig Signature, err error) {
	return Sign(target, &a.PrivKey)
}

// NewTX 该账户作为主体，构造新交易
func (a *Account) NewTX(typ uint, args ArgsOfNewTX) (tx TX, err error) {
	// TODO: 根据账户类型不同来处理
	return newTransaction(typ, args)
	// TODO： 这层只是简单调用，参数检查交给tx自己去做。
}

/*********************************************************************************************************************
                                                    SelfAccounts相关
*********************************************************************************************************************/

// SelfAccounts 自己的账户集合
type SelfAccounts struct {
	Map map[string]*Account	`json:"selfAccounts"`
}

// SaveFileWithGobEncode 将内存维护的自己账户表写入本地指定路径下的文件
func (as *SelfAccounts) SaveFileWithGobEncode(port uint) (err error) {
	file := fmt.Sprintf(SELFACCOUNTS_FILEPATH_TEMP, strconv.Itoa(int(port)))
	if err = saveFileWithGobEncode(file, as); err != nil {
		return WrapError("SelfAccounts_SaveFile", err)
	}
	return nil
}

// SaveFileWithJsonMarshal 将内存维护的自己账户表写入本地指定路径下的文件
func (as *SelfAccounts) SaveFileWithJsonMarshal(port uint) (err error) {
	file := fmt.Sprintf(SELFACCOUNTS_FILEPATH_TEMP, strconv.Itoa(int(port)))
	if err = saveFileWithJsonMarshal(file, as); err != nil {
		return WrapError("SelfAccounts_SaveFile", err)
	}
	return nil
}

// LoadFileWithGobDecode 从本地文件中读取自己账户表（用于加载）
func (as *SelfAccounts) LoadFileWithGobDecode(port uint) (err error) {
	file := fmt.Sprintf(SELFACCOUNTS_FILEPATH_TEMP, strconv.Itoa(int(port)))
	if _, err = os.Stat(file); os.IsNotExist(err) {
		return WrapError("SelfAccounts_LoadFile", err)
	}

	var accounts SelfAccounts

	fileContent, err := ioutil.ReadFile(file)
	if err != nil {
		return WrapError("SelfAccounts_LoadFile", err)
	}

	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	if err = decoder.Decode(&accounts); err != nil {
		return WrapError("SelfAccounts_LoadFile", err)
	}

	as.Map = accounts.Map
	return nil
}

// LoadFileWithJsonUnmarshal 从本地文件中读取自己账户表（用于加载）
func (as *SelfAccounts) LoadFileWithJsonUnmarshal(port uint) (err error) {
	file := fmt.Sprintf(SELFACCOUNTS_FILEPATH_TEMP, strconv.Itoa(int(port)))
	if _, err = os.Stat(file); os.IsNotExist(err) {
		return WrapError("SelfAccounts_LoadFile", err)
	}

	var accounts SelfAccounts

	fileContent, err := ioutil.ReadFile(file)
	if err != nil {
		return WrapError("SelfAccounts_LoadFile", err)
	}

	if err = json.Unmarshal(fileContent, &accounts); err != nil {
		return WrapError("SelfAccounts_LoadFile", err)
	}

	as.Map = accounts.Map
	return nil
}

// CreateSelfAccounts 从文件创建新的自己账户表(用于在还没有SelfAccounts时的创建)
func CreateSelfAccountsFromJsonFile(port uint) (*SelfAccounts, error) {
	accounts := SelfAccounts{}
	accounts.Map = make(map[string]*Account)

	err := accounts.LoadFileWithJsonUnmarshal(port)
	if err != nil {
		return nil, WrapError("CreateSelfAccounts", err)
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
func (as *SelfAccounts) AddAccount(roleNo uint) (userID UserID, err error) {
	account, err := NewAccount(roleNo)
	if err != nil {
		return UserID{}, WrapError("SelfAccounts_AddAccount", err)
	}
	userID, err = account.UserID()
	if err != nil {
		return UserID{}, WrapError("SelfAccounts_AddAccount", err)
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
	UserID    UserID `json:"userID"`
	PublicKey    []byte `json:"pubKey"`
	BalanceCoin   Coin `json:"balance"`
	RoleField      Role	`json:"role"`// 存储角色信息，和UserID中都存了一次RoleNo。必须保证相同
	AvailableField bool `json:"available"`
	RegisterInfo RegisterInfo `json:"registerInfo"`
}



// PubKey 获取账户公钥
func (a *EcoinAccount) PubKey() []byte {
	return a.PublicKey
}

// Balance 获取余额
func (a *EcoinAccount) Balance() Coin {
	return a.BalanceCoin
}

// Role 获取账户的角色
func (a *EcoinAccount) Role() *Role {
	return &a.RoleField
}

// Available 账户是否可用
func (a *EcoinAccount) Available() bool {
	return a.AvailableField
}

/*********************************************************************************************************************
                                                    EcoinAccountMap相关
*********************************************************************************************************************/

// EcoinAccountsMap 系统账户表 存储每个账户的可公开的信息，包括余额、角色（角色中定义了其币相关的规则）、公钥、可用状态。键值为UserID.ID
// TODO: 后期将之改为状态树或者叫账户树。因为现在这么做如果账户很多其实占用很大。
type EcoinAccounts struct {
	Map map[string]*EcoinAccount 	`json:"ecoinAccounts"`
}

// todo: 注意： 这张表没有历史记录，也就意味着无法回滚状态。

// SaveFileWithGobEncode 将内存维护的自己账户表写入本地指定路径下的文件
func (as *EcoinAccounts) SaveFileWithGobEncode(port uint) (err error) {
	file := fmt.Sprintf(ECOINACCOUNTS_FILEPATH_TEMP, strconv.Itoa(int(port)))

	if err = saveFileWithGobEncode(file, as); err != nil {
		return WrapError("EcoinAccounts_SaveFile", err)
	}
	return nil
}

// SaveFileWithJsonMarshal 将内存维护的自己账户表写入本地指定路径下的文件
func (as *EcoinAccounts) SaveFileWithJsonMarshal(port uint) (err error) {
	file := fmt.Sprintf(ECOINACCOUNTS_FILEPATH_TEMP, strconv.Itoa(int(port)))

	if err = saveFileWithJsonMarshal(file, as); err != nil {
		return WrapError("EcoinAccounts_SaveFile", err)
	}
	return nil
}

// LoadFileWithGobDecode 从本地文件中读取自己账户表（用于加载）
func (as *EcoinAccounts) LoadFileWithGobDecode(port uint) (err error) {
	file := fmt.Sprintf(ECOINACCOUNTS_FILEPATH_TEMP, strconv.Itoa(int(port)))
	if _, err = os.Stat(file); os.IsNotExist(err) {
		return WrapError("EcoinAccounts_LoadFile", err)
	}

	var accounts EcoinAccounts

	fileContent, err := ioutil.ReadFile(file)
	if err != nil {
		return WrapError("EcoinAccounts_LoadFile", err)
	}

	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	if err = decoder.Decode(&accounts); err != nil {
		return WrapError("EcoinAccounts_LoadFile", err)
	}

	as.Map = accounts.Map
	return nil
}

// LoadFileWithJsonUnmarshal 从本地文件中读取自己账户表（用于加载）
func (as *EcoinAccounts) LoadFileWithJsonUnmarshal(port uint) (err error) {
	file := fmt.Sprintf(ECOINACCOUNTS_FILEPATH_TEMP, strconv.Itoa(int(port)))
	if _, err = os.Stat(file); os.IsNotExist(err) {
		return WrapError("EcoinAccounts_LoadFile", err)
	}

	accounts := EcoinAccounts{}

	fileContent, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println("1111111111")
		return WrapError("EcoinAccounts_LoadFile", err)
	}

	if err = json.Unmarshal(fileContent, &accounts); err != nil {
		fmt.Println("2222222222")
		return WrapError("EcoinAccounts_LoadFile", err)
	}

	as.Map = accounts.Map

	return nil
}

// CreateEcoinAccountsFromJsonFile 从文件创建新的自己账户表(用于在还没有EcoinAccounts时的创建)
func CreateEcoinAccountsFromJsonFile(port uint) (*EcoinAccounts, error) {
	accounts := EcoinAccounts{}
	accounts.Map = make(map[string]*EcoinAccount)

	err := accounts.LoadFileWithJsonUnmarshal(port)
	if err != nil {
		return nil, WrapError("CreateEcoinAccounts", err)
	}
	return &accounts, nil
}

// GetAccount 根据用户id查看公开账户
func (as *EcoinAccounts) GetAccount(userID string) *EcoinAccount {
	if v, ok := as.Map[userID]; ok {
		return v
	}
	return nil
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
	return newEcoinAccount.UserID, nil
}