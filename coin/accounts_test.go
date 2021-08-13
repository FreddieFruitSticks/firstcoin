package coin

import (
	"crypto/x509"
	"encoding/pem"
	"testing"
)

func TestVerify(t *testing.T) {
	account := NewAccount()

	account.GenerateKeyPair()
	pemBlock, _ := pem.Decode([]byte(account.PublicKey))

	signature := account.GenerateSignature()

	publicKey, err := x509.ParsePKCS1PublicKey(pemBlock.Bytes)
	if err != nil {
		t.Fatalf("err parsing pk %s", err)
	}

	verify := account.VerifySignature(signature, publicKey)

	if !verify{
		t.Fatalf("signature not confirmed")
	}
}
