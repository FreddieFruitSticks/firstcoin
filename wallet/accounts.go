package wallet

import (
	"blockchain/utils"
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

type Account struct {
	PublicKey        []byte
	PrivateKey       []byte
	PrivateKeyObject *rsa.PrivateKey
	PublicKeyObject  *rsa.PublicKey
}

func NewAccount() *Account {
	return &Account{}
}

func (a *Account) GenerateKeyPair() {
	// generate key
	privatekey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Printf("Cannot generate RSA key\n")
		os.Exit(1)
	}
	a.PrivateKeyObject = privatekey
	publickey := &privatekey.PublicKey
	a.PublicKeyObject = publickey

	// private key
	var privateKeyBytes []byte = x509.MarshalPKCS1PrivateKey(privatekey)
	privateKeyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	var b []byte
	buffer := bytes.NewBuffer(b)

	err = pem.Encode(buffer, privateKeyBlock)
	if err != nil {
		fmt.Printf("error when encode private pem: %s \n", err)
		os.Exit(1)
	}

	sk := make([]byte, 2048)

	_, err = buffer.Read(sk)
	utils.CheckError(err)

	a.PrivateKey = sk

	buffer.Reset()

	// public key
	publicKeyBytes := x509.MarshalPKCS1PublicKey(publickey)
	publicKeyBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}

	err = pem.Encode(buffer, publicKeyBlock)
	if err != nil {
		fmt.Printf("error when encode private pem: %s \n", err)
		os.Exit(1)
	}

	pk := make([]byte, 418)

	_, err = buffer.Read(pk)
	utils.CheckError(err)

	a.PublicKey = pk
}

func (a *Account) GenerateSignature(message []byte) []byte {
	msgHashSum := hashMessage(message)

	// In order to generate the signature, we provide a random number generator,
	// our private key, the hashing algorithm that we used, and the hash sum
	// of our message
	signature, err := rsa.SignPSS(rand.Reader, a.PrivateKeyObject, crypto.SHA256, msgHashSum, nil)
	if err != nil {
		panic(err)
	}

	return signature
}

func (a *Account) VerifySignature(signature []byte, publicKey []byte, message []byte) error {
	pemBlock, _ := pem.Decode(publicKey)
	if pemBlock == nil {
		return fmt.Errorf("error verifying signature: could not find pemBlock for public key")
	}

	pk, err := x509.ParsePKCS1PublicKey(pemBlock.Bytes)
	if err != nil {
		return fmt.Errorf("error verifying signature: could not parse public key %+v", err)
	}

	msgHashSum := hashMessage(message)

	err = rsa.VerifyPSS(pk, crypto.SHA256, msgHashSum, signature, nil)
	if err != nil {
		fmt.Println()
		return fmt.Errorf("%+v", err)
	}

	return nil
}

func hashMessage(msg []byte) []byte {
	msgHash := sha256.New()
	_, err := msgHash.Write(msg)
	if err != nil {
		panic(err)
	}

	return msgHash.Sum(nil)
}
