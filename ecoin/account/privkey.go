package account

import (
	"crypto/ecdsa"
	"math/big"
)

// PrivKey 私钥
type PrivKey []byte

// PrivateKey 转换回ecdsa.PrivateKey
func (priv PrivKey) PrivateKey(publicKey ecdsa.PublicKey) ecdsa.PrivateKey {
	d := &big.Int{}
	d = d.SetBytes(priv)
	return ecdsa.PrivateKey{
		PublicKey: publicKey,
		D: d,
	}
}
