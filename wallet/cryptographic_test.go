package wallet

import (
	"testing"
)

func TestVerify(t *testing.T) {
	crypt := NewCryptographic()
	crypt.GenerateKeyPair()

	message := []byte{12, 23}

	signature := crypt.GenerateSignature(message)

	if err := crypt.VerifySignature(signature, crypt.PublicKey, message); err != nil {
		t.Fatalf("signature not confirmed: %+v", err)
	}
}
