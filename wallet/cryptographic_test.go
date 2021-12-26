package wallet_test

import (
	"blockchain/wallet"
	"testing"
)

func TestVerify(t *testing.T) {
	crypt := wallet.NewCryptographic()
	crypt.GenerateKeyPair()

	w := wallet.NewWallet(*crypt)
	message := []byte{12, 23}

	scriptSig := w.GenerateTxSigScript(message)

	if err := wallet.VerifySignature(scriptSig, crypt.FirstcoinAddress, message); err != nil {
		t.Fatalf("signature not confirmed: %+v", err)
	}
}
