package wallet

import (
	"testing"
)

func TestVerify(t *testing.T) {
	account := NewAccount()

	account.GenerateKeyPair()
	// pemBlock, _ := pem.Decode([]byte(account.PublicKey))
	message := []byte{12, 23}

	signature := account.GenerateSignature(message)

	// publicKey, err := x509.ParsePKCS1PublicKey(pemBlock.Bytes)
	// if err != nil {
	// 	t.Fatalf("err parsing pk %s", err)
	// }

	verify, _ := account.VerifySignature(signature, account.PublicKey, message)

	if !verify {
		t.Fatalf("signature not confirmed")
	}
}
