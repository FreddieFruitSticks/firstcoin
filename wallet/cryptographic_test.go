package wallet_test

import (
	"blockchain/wallet"
	"testing"
)

func TestVerify(t *testing.T) {
	crypt := wallet.NewCryptographic()
	crypt.GenerateKeyPair()

	message := []byte{12, 23}

	signature := crypt.GenerateSignature(message)

	if err := wallet.VerifySignature(signature, crypt.PublicKey, message); err != nil {
		t.Fatalf("signature not confirmed: %+v", err)
	}
}
