package ecoinlib

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"github.com/mr-tron/base58"
	"golang.org/x/crypto/ripemd160"
)

// 采取与比特币基本一致的账户生成

type UserID string

type Account struct {
	PrivKey ecdsa.PrivateKey
	PubKey  []byte
}

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

// TODO: 注意：新建账户时需添加入EcoinWorld
func NewAccount() (*Account, error) {
	privKey, pubKey, err := newKeyPair()
	if err != nil {
		return nil, fmt.Errorf("NewAccount: %s", err)
	}
	return &Account{
		PrivKey: privKey,
		PubKey:  pubKey,
	}, nil
}

// publicKey -> sha256 -> publicKeyHash -> ripemd160 -> publicKeyHashRipemd160
func pubKeyHash(pubKey []byte) ([]byte, error) {
	pubHash := sha256.Sum256(pubKey)

	hasher := ripemd160.New()
	if _, err := hasher.Write(pubHash[:]); err != nil {
		return nil, fmt.Errorf("PubKeyHash: %s", err)
	}
	return hasher.Sum(nil), nil
}

// publicKeyHashRipemd160 -> sha256 -> sha256 -> [:checksumLength] -> checksum
func checksum(payload []byte, checksumLength int) []byte {
	firstHash := sha256.Sum256(payload)
	secondHash := sha256.Sum256(firstHash[:])
	return secondHash[:checksumLength]
}

// publicKeyHashRipemd160 + checksum + version -> base58 -> userID
func (a Account) UserID(checksumLength int, version byte) (UserID, error) {
	pubHash, err := pubKeyHash(a.PubKey)
	if err != nil {
		return "", fmt.Errorf("Account_UserID: %s", err)
	}
	versionedHash := append([]byte{version}, pubHash...)
	checksum := checksum(versionedHash, checksumLength)
	fullHash := append(versionedHash, checksum...)

	userId := base58.Encode(fullHash)
	return UserID(userId), nil
}

func ValidateUserID(userId UserID, checksumLength int) (bool, error) {
	fullPubKeyHash, err := base58.Decode(string(userId))
	if err != nil {
		return false, fmt.Errorf("ValidateUserID: base58_Decode: %s", err)
	}
	length := len(fullPubKeyHash)
	actualChecksum := fullPubKeyHash[length-checksumLength:]
	version := fullPubKeyHash[0]
	pubKeyHash := fullPubKeyHash[1 : length-checksumLength]
	targetChecksum := checksum(append([]byte{version}, pubKeyHash...), checksumLength)
	return bytes.Compare(actualChecksum, targetChecksum) == 0, nil
}

