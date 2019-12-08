package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"math/big"

	"github.com/azd1997/Ecare/ecoin/utils"
)

// ECDSA 椭圆曲线签名。注意，由于所使用的go/crypto/ecdsa只是签名算法，所以这里并没有加密解密实现
type ECDSA struct {}

func (e ECDSA) GenerateKeyPair() (PrivKey, PubKey, error) {
	// 椭圆曲线
	curve := elliptic.P256()
	// 生成私钥
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, nil, utils.WrapError("ECDSA_GenerateKeyPair", err)
	}
	pubKey := append(privateKey.PublicKey.X.Bytes(), privateKey.PublicKey.Y.Bytes()...)
	privKey := append(pubKey, privateKey.D.Bytes()...)

	return privKey, pubKey, nil
}

func (e ECDSA) Sign(hash Hash, priv PrivKey) (Signature, error) {
	privateKey := e.privateKey(priv)
	r, s, err := ecdsa.Sign(rand.Reader, &privateKey, hash)
	if err != nil {
		return nil, utils.WrapError("ECDSA_Sign", err)
	}
	sig := append(r.Bytes(), s.Bytes()...)
	return sig, nil
}

func (e ECDSA)  VerifySign(hash Hash, sig Signature, pub PubKey) bool {
	// 从sig还原出r,s两个大数
	sigLen := len(sig)
	r, s := &big.Int{}, &big.Int{}
	r, s = r.SetBytes(sig[:sigLen/2]), s.SetBytes(sig[sigLen/2:]) // 基于下标范围创建新切片的时候下标范围是半开区间 [start, end)

	// 还原ecdsa.PublicKey
	publicKey := e.publicKey(pub)

	// 验证签名
	return ecdsa.Verify(&publicKey, hash, r, s)
}

func (e ECDSA)  Encrypt(raw []byte, pub PubKey) (encrypted []byte, err error) {
	return nil, nil
}

func (e ECDSA)  Decrypt(encrypted []byte, priv PrivKey) (raw []byte, err error) {
	return nil, nil
}

func (e ECDSA) privateKey(priv PrivKey) ecdsa.PrivateKey {
	l := len(priv)	// l=96
	pub := priv[:2*l/3]
	publicKey := e.publicKey(PubKey(pub))
	d := &big.Int{}
	d = d.SetBytes(priv[2*l/3:])

	return ecdsa.PrivateKey{
		PublicKey: publicKey,
		D: d,
	}
}

func (e ECDSA) publicKey(pub PubKey) ecdsa.PublicKey {
	pubKeyLen := len(pub)
	x, y := &big.Int{}, &big.Int{}
	x, y = x.SetBytes(pub[: pubKeyLen/2]), y.SetBytes(pub[pubKeyLen/2 :])
	return ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}
}