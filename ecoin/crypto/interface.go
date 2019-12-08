package crypto

// AsymmetricCrypto AsymmetricCryptography 非对称加密接口
type AsymmetricCrypto interface {
	GenerateKeyPair() (PrivKey, PubKey, error)
	Sign(hash Hash, priv PrivKey) (Signature, error)	// Hash就是签名保护的内容
	VerifySign(hash Hash, sig Signature, pub PubKey) bool
	Encrypt(raw []byte, pub PubKey) (encrypted []byte, err error)
	Decrypt(encrypted []byte, priv PrivKey) (raw []byte, err error)
}
