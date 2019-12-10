package account

import (
	"github.com/azd1997/ego/ecrypto"
	"github.com/azd1997/ego/ecrypto/ecdsa"
)

// ACrypto AsymmetricCryptography
var ACrypto ecrypto.AsymmetricCrypto

func init()  {
	// 想要更改签名算法，只需要在此修改即可
	ACrypto = ecdsa.ECDSA{}
	// ACrypto = rsa.RSA{}
}
