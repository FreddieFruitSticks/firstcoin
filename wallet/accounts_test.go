package wallet

import (
	"testing"
)

func TestVerify(t *testing.T) {
	account := NewAccount()
	account.GenerateKeyPair()

	message := []byte{12, 23}

	signature := account.GenerateSignature(message)

	if err := account.VerifySignature(signature, account.PublicKey, message); err != nil {
		t.Fatalf("signature not confirmed: %+v", err)
	}
}
