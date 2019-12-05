package account

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"math/big"
)

// PubKey 公钥
type PubKey []byte

// PublicKey 将自定义的切片形式的公钥转回ecdsa.PublicKey
func (pub PubKey) PublicKey() ecdsa.PublicKey {
	pubKeyLen := len(pub)
	x, y := &big.Int{}, &big.Int{}
	x, y = x.SetBytes(pub[: pubKeyLen/2]), y.SetBytes(pub[pubKeyLen/2 :])
	return ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}
}

