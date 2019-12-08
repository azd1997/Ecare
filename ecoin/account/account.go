package account

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/json"
	"github.com/azd1997/Ecare/ecoin/crypto"
	"io/ioutil"
	"os"

	"github.com/mr-tron/base58"
	"golang.org/x/crypto/ripemd160"

	"github.com/azd1997/Ecare/ecoin/log"
	"github.com/azd1997/Ecare/ecoin/utils"
)

// Account 账户，包含私钥和公钥，标志唯一身份。UserID是外部可见的标志
type Account struct {
	PrivKey crypto.PrivKey `json:"privKey"`
	PubKey  crypto.PubKey  `json:"pubKey"`
	RoleNo  uint           `json:"roleNo"`
}

// NewAccount 新建账户
// TODO: 注意：新建账户时需添加入本地gsm.accounts并向外广播.
func NewAccount(roleNo uint) (*Account, error) {
	privKey, pubKey, err := newKeyPair()
	if err != nil {
		return nil, utils.WrapError("NewAccount", err)
	}
	return &Account{
		PrivKey: privKey,
		PubKey:  pubKey,
		RoleNo:  roleNo,
	}, nil
}

// LoadOrCreateAccount 从指定路径加载账户，加载不到就新建
func LoadOrCreateAccount(accountFile string) (account *Account, err error) {
	account = &Account{}
	// 1.1. 加载selfAccount文件，取账户文件
	// 1.1.1 检查是否存在账户文件
	exists, err := utils.FileExists(accountFile)
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
		err = account.SaveFileWithGobEncode(accountFile)
		if err != nil {
			return nil, err
		}
		id, err := account.UserId()
		if err != nil {
			return nil, err
		}
		log.Success("新账户创建成功并保存至默认路径， 账户ID: %s", id.String())
	} else {
		// 若存在，则从这个文件读取account
		log.Info("%s", "默认路径下发现账户文件， 准备加载......")
		err = account.LoadFileWithGobDecode(accountFile)
		if err != nil {
			return nil, err
		}
		id, err := account.UserId()
		if err != nil {
			return nil, err
		}
		log.Success("账户加载成功， 账户ID: %s", id.String())
	}

	return account, nil
}

// TODO: 待解决的问题：多个账户文件在同一个目录下怎么去选取。目前的做法是只读取指定文件名的账户文件。但如果要考虑多个账户呢？

/*******************************************************实现接口*********************************************************/

// String 打印字符串
func (a *Account) String() string {
	return utils.JsonMarshalIndentToString(a)
}

// UserId publicKeyHashRipemd160 + checksum + version -> base58 -> userID
func (a *Account) UserId() (UserId, error) {
	aCopy := a
	pubHash, err := pubKeyHash(aCopy.PubKey)
	if err != nil {
		return UserId{}, utils.WrapError("Account_UserId", err)
	}
	versionedHash := append([]byte{ACCOUNT_VERSION}, pubHash...)
	checksum := checksum(versionedHash)
	fullHash := append(versionedHash, checksum...)

	userId := base58.Encode(fullHash)
	return UserId{userId, aCopy.RoleNo}, nil
}

// Sign 使用该账号对目标数据作签名。目标数据只能是基础类型、结构体、切片、表等，必须提前转为[]byte
func (a *Account) Sign(target []byte) (sig crypto.Signature, err error) {
	return ACrypto.Sign(target, a.PrivKey, a.PubKey)
}

// VerifySign 验证签名; 这个pubKey不一定是本账户的PubKey
func (a *Account) VerifySign(target []byte, sig []byte, pubKey []byte) bool {
	return ACrypto.VerifySign(target, sig, pubKey)
}

// NewTX 该账户作为主体，构造新交易
// NewTX交给其他地方做
//func (a *Account) NewTX(typ uint, args ArgsOfNewTX) (tx TX, err error) {
//	// TODO: 根据账户类型不同来处理
//	return newTransaction(typ, args)
//	// TODO： 这层只是简单调用，参数检查交给tx自己去做。
//}

// SaveFileWithGobEncode 保存到文件
func (a *Account) SaveFileWithGobEncode(file string) (err error) {
	// 由于Account中只是用了字节切片，所以可以直接编码
	if err = utils.SaveFileWithGobEncode(file, a); err != nil {
		return utils.WrapError("Account_SaveFile", err)
	}
	return nil
}

// LoadFileWithGobDecode 从本地文件中读取自己账户表（用于加载）
func (a *Account) LoadFileWithGobDecode(file string) (err error) {
	if _, err = os.Stat(file); os.IsNotExist(err) {
		return utils.WrapError("Account_LoadFile", err)
	}

	var account Account

	fileContent, err := ioutil.ReadFile(file)
	if err != nil {
		return utils.WrapError("Account_LoadFile", err)
	}

	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	if err = decoder.Decode(&account); err != nil {
		return utils.WrapError("Account_LoadFile", err)
	}

	a.PrivKey = account.PrivKey
	a.PubKey = account.PubKey
	a.RoleNo = account.RoleNo

	return nil
}

// SaveFileWithJsonEncode 保存到文件
func (a *Account) SaveFileWithJsonEncode(file string) (err error) {
	if err = utils.SaveFileWithJsonMarshal(file, a); err != nil {
		return utils.WrapError("Account_SaveFile", err)
	}
	return nil
}

// LoadFileWithJsonDecode 从本地文件中读取自己账户表（用于加载）
func (a *Account) LoadFileWithJsonDecode(file string) (err error) {
	if _, err = os.Stat(file); os.IsNotExist(err) {
		return utils.WrapError("Account_LoadFile", err)
	}

	var account Account

	fileContent, err := ioutil.ReadFile(file)
	if err != nil {
		return utils.WrapError("Account_LoadFile", err)
	}

	err = json.Unmarshal(fileContent, account)
	if err != nil {
		return utils.WrapError("Account_LoadFile", err)
	}

	a.PrivKey = account.PrivKey
	a.PubKey = account.PubKey
	a.RoleNo = account.RoleNo

	return nil
}

/*******************************************************实现接口*********************************************************/

// newKeyPair 创造新的公私钥对
func newKeyPair() (crypto.PrivKey, crypto.PubKey, error) {
	return ACrypto.GenerateKeyPair()
}

// pubKeyHash publicKey -> sha256 -> publicKeyHash -> ripemd160 -> publicKeyHashRipemd160
func pubKeyHash(pubKey []byte) ([]byte, error) {
	pubHash := sha256.Sum256(pubKey)

	hasher := ripemd160.New()
	if _, err := hasher.Write(pubHash[:]); err != nil {
		return nil, utils.WrapError("PubKeyHash", err)
	}
	return hasher.Sum(nil), nil
}

// checksum publicKeyHashRipemd160 -> sha256 -> sha256 -> [:checksumLength] -> checksum
func checksum(payload []byte) []byte {
	firstHash := sha256.Sum256(payload)
	secondHash := sha256.Sum256(firstHash[:])
	return secondHash[:CHECKSUM_LENGTH]
}
