package ecoinlib

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"math/big"
)

// 交易要检查：
// 1.转账者、接收者是否存在
// 2.转账金额非负为整
// 3.转账者余额是否足够

// 一笔交易由转账者构建，A当然可以创建这个交易，但这个问题在于怎么确保其他人无法创建以A的地址和签名的交易

type Transaction struct {
	ID          []byte
	CreateTime  int64 // 创建时间戳
	SubmitTime  int64 // 提交时间戳
	PassTime    int64 // 生效时间戳
	From, To    UserID
	Amount      uint
	Description string
	Signature   []byte // 转账者签名
}

func NewTransaction(from *Account, to UserID, amount uint, description string, checksumLength int, version byte) (tx *Transaction, err error) {
	// 余额不足报错
	fromUserID, err := from.UserID(checksumLength, version)
	if err != nil {
		return nil, fmt.Errorf("NewTransaction: %s", err)
	}
	if uint(EcoinWorld.GetBalanceOfUserID(fromUserID)) < amount {
		return nil, fmt.Errorf("NewTransaction: %s", ErrNotSufficientBalance)
	}
	// 构造交易
	tx = &Transaction{
		From:        fromUserID,
		To:          to,
		Amount:      amount,
		Description: description,
	}
	// 获取并设置ID
	id, err := tx.Hash()
	if err != nil {
		return nil, fmt.Errorf("NewTransaction: %s", err)
	}
	tx.ID = id
	// 签名
	if err = tx.Sign(from.PrivKey); err != nil {
		return nil, fmt.Errorf("NewTransaction: %s", err)
	}

	return tx, nil
}

func CoinbaseTx(to UserID, description string) (tx *Transaction, err error) {
	// coinbase交易只允许role0(定义为创始者)构建

	// 检查to是否为role0创始者
	if EcoinWorld.accounts[to].role.No() != 0 {
		return nil, ErrCoinbaseTxRequireRole0
	}

	// 构造tx
	tx = &Transaction{
		To:          to,
		Amount:      uint(EcoinWorld.accounts[to].role.InitialBalance()),
		Description: description,
	}

	// 设置Id
	id, err := tx.Hash()
	if err != nil {
		return nil, fmt.Errorf("CoinbaseTx: %s", err)
	}
	tx.ID = id
	return tx, nil
}

func (tx *Transaction) Verify(checksumLength int) (valid bool, err error) {

	// 1. 验证转账者，接收者地址是否合法，是否存在，是否可用
	if addrValid, _ := ValidateUserID(tx.From, checksumLength); !addrValid {
		return false, fmt.Errorf("Transaction_Verify: %s: tx.From", ErrInvalidUserID)
	}
	if addrValid, _ := ValidateUserID(tx.To, checksumLength); !addrValid {
		return false, fmt.Errorf("Transaction_Verify: %s: tx.To", ErrInvalidUserID)
	}
	// 是否存在
	if !EcoinWorld.HasUserID(tx.From) {
		return false, fmt.Errorf("Transaction_Verify: %s: tx.From", ErrNonexistentUserID)
	}
	if !EcoinWorld.HasUserID(tx.To) {
		return false, fmt.Errorf("Transaction_Verify: %s: tx.To", ErrNonexistentUserID)
	}
	// 是否可用
	if !EcoinWorld.IsUserIDAvailable(tx.From) {
		return false, fmt.Errorf("Transaction_Verify: %s: tx.From", ErrUnavailableUserID)
	}
	if !EcoinWorld.IsUserIDAvailable(tx.To) {
		return false, fmt.Errorf("Transaction_Verify: %s: tx.To", ErrUnavailableUserID)
	}

	// 2. 转账金额非负为整
	if tx.Amount < 0 {
		return false, fmt.Errorf("Transaction_Verify: %s", ErrNegativeTransferAmount)
	}

	// 3. 转账者余额足够
	if uint(EcoinWorld.GetBalanceOfUserID(tx.From)) < tx.Amount {
		return false, fmt.Errorf("Transaction_Verify: %s", ErrNotSufficientBalance)
	}

	// 4. 验证交易签名，确保是转账者本人操作
	// 复制一份tx.Id
	var hash []byte
	hash = tx.ID
	// 还原r,s
	r, s := big.Int{}, big.Int{}
	length := len(tx.Signature)
	r.SetBytes(tx.Signature[:(length / 2)])
	s.SetBytes(tx.Signature[(length / 2):])
	// 还原x,y
	x, y := big.Int{}, big.Int{}
	pubKey := EcoinWorld.GetPubKeyOfUserID(tx.From)
	length = len(pubKey)
	x.SetBytes(pubKey[:(length / 2)])
	y.SetBytes(pubKey[(length / 2):])
	// 还原原始publicKey
	rawPubKey := ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     &x,
		Y:     &y,
	}
	// 验证
	if ecdsa.Verify(&rawPubKey, hash, &r, &s) == false {
		return false, nil
	}

	return true, nil
}

func (tx *Transaction) IsCoinbase() bool {

	// 检查coinbase to是否为role0，同时满足from为空
	return EcoinWorld.accounts[tx.To].role.No() == 0 && tx.From == ""
}

func (tx *Transaction) String() string {
	return fmt.Sprintf(
`{
	id: 		%s
	from: 		%s
	to:   		%s
	amount: 	%d
	description: 	%s
	signature: 		%s
}`,
		tx.ID, tx.From, tx.To, tx.Amount, tx.Description, tx.Signature)
}

func (tx *Transaction) Serialize() (result []byte, err error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err = encoder.Encode(tx); err != nil {
		return nil, fmt.Errorf("Transaction_Serialize: %s", err)
	}
	return buf.Bytes(), nil
}

func (tx *Transaction) Hash() (id []byte, err error) {
	txCopy := *tx
	txCopy.ID = []byte{}
	txCopy.Signature = []byte{}
	var res []byte
	if res, err = txCopy.Serialize(); err != nil {
		return nil, fmt.Errorf("Transaction_Hash: %s", err)
	}
	hash := sha256.Sum256(res)
	return hash[:], nil
}

func (tx *Transaction) Sign(privKey ecdsa.PrivateKey) (err error) {
	var hash []byte
	hash = tx.ID
	r, s, err := ecdsa.Sign(rand.Reader, &privKey, hash)
	if err != nil {
		return fmt.Errorf("Transaction_Sign: %s", err)
	}
	signature := append(r.Bytes(), s.Bytes()...)
	tx.Signature = signature
	return nil
}
