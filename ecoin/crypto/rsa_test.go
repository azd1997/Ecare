package crypto

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

func TestRSA_GenerateKeyPair(t *testing.T) {
	ac := RSA{}
	priv, pub, err := ac.GenerateKeyPair()
	if err != nil {
		t.Error(err)
	}

	fmt.Println(priv)
	fmt.Println(pub)
}

func TestRSA_Sign(t *testing.T) {
	ac := RSA{}
	priv, pub, err := ac.GenerateKeyPair()
	if err != nil {
		t.Error(err)
	}

	str := []byte("Hello Eiger")
	hash256 := sha256.Sum256(str)
	hash := Hash(hash256[:])

	sig, err := ac.Sign(hash, priv)
	if err != nil {
		t.Error(err)
	}

	valid := ac.VerifySign(hash, sig, pub)
	if !valid {
		t.Error("invalid signature")
	}
}

func TestRSA_Encrypt(t *testing.T) {

	ac := RSA{}
	priv, pub, err := ac.GenerateKeyPair()
	if err != nil {
		t.Error(err)
	}

	raw := []byte("Hello Eiger")

	enc, err := ac.Encrypt(raw, pub)
	if err != nil {
		t.Error(err)
	}

	raw2, err := ac.Decrypt(enc, priv)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("raw: ", string(raw))
	fmt.Println("raw2: ", string(raw2))

	if string(raw) != string(raw2) {
		t.Error("failed")
	}
}
