package utils

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"math/big"
)

// VerifySignature 用公钥验证签名
func VerifySignature(target []byte, sig []byte, pubKey []byte) bool {
	// 从sig还原出r,s两个大数
	sigLen := len(sig)
	r, s := &big.Int{}, &big.Int{}
	r, s = r.SetBytes(sig[:sigLen / 2]), s.SetBytes(sig[sigLen / 2 :])		// 基于下标范围创建新切片的时候下标范围是半开区间 [start, end)

	// 还原ecdsa.PublicKey
	pubKeyLen := len(pubKey)
	x, y := &big.Int{}, &big.Int{}
	x, y = x.SetBytes(pubKey[: pubKeyLen/2]), y.SetBytes(pubKey[pubKeyLen/2 :])
	rawPubKey := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}

	// 验证签名
	return ecdsa.Verify(rawPubKey, target, r, s)
}

// Sign 用私钥对目标进行签名
func Sign(target []byte, privKey *ecdsa.PrivateKey) (sig []byte, err error) {
	r, s, err := ecdsa.Sign(rand.Reader, privKey, target)
	if err != nil {
		return nil, WrapError("Sign", err)
	}
	signature := append(r.Bytes(), s.Bytes()...)
	return signature, nil
}
