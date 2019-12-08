package account

import "github.com/azd1997/Ecare/ecoin/crypto"

// ACrypto AsymmetricCryptography
var ACrypto crypto.AsymmetricCrypto

func init()  {
	// 想要更改签名算法，只需要在此修改即可
	ACrypto = crypto.ECDSA{}
}
