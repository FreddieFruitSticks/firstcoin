package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/md5"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

type Cryptographic struct {
	PublicKey        []byte
	PrivateKey       []byte
	PrivateKeyObject *ecdsa.PrivateKey
	PublicKeyObject  *ecdsa.PublicKey
}

func NewCryptographic() *Cryptographic {
	return &Cryptographic{}
}

// TODO: Bitcoin runs on ECDSA not RSA.
func (c *Cryptographic) GenerateKeyPair() {
	curve := elliptic.P256()
	// generate key
	privatekey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		fmt.Printf("Cannot generate RSA key\n")
		os.Exit(1)
	}

	c.PrivateKeyObject = privatekey

	publickey := &privatekey.PublicKey
	c.PublicKeyObject = publickey

	// private key
	privateKeyBytes, err := x509.MarshalECPrivateKey(privatekey)
	if err != nil {
		panic(err)
	}

	privateKeyBlock := &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	pemEncodedPrivKey := pem.EncodeToMemory(privateKeyBlock)
	c.PrivateKey = pemEncodedPrivKey

	// public key
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publickey)
	if err != nil {
		panic(err)
	}

	publicKeyBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}

	pemEncodedPubKey := pem.EncodeToMemory(publicKeyBlock)
	c.PublicKey = pemEncodedPubKey
}

func (c *Cryptographic) GenerateSignature(message []byte) []byte {
	msgHashSum := hashMessage(message)

	signature, err := ecdsa.SignASN1(rand.Reader, c.PrivateKeyObject, msgHashSum)
	if err != nil {
		panic(err)
	}

	return signature
}

func VerifySignature(signature []byte, publicKey []byte, message []byte) error {
	pemBlock, _ := pem.Decode(publicKey)
	if pemBlock == nil {
		return fmt.Errorf("error verifying signature: could not find pemBlock for public key")
	}

	genericPublicKey, err := x509.ParsePKIXPublicKey(pemBlock.Bytes)
	if err != nil {
		return fmt.Errorf("error verifying signature: could not parse public key %+v", err)
	}

	pubKey, ok := genericPublicKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("could not cast pubKey to ecdsa pubKey")
	}

	msgHashSum := hashMessage(message)

	verify := ecdsa.VerifyASN1(pubKey, msgHashSum, signature)
	if !verify {
		return fmt.Errorf("invalid signature")
	}

	return nil
}

func hashMessage(msg []byte) []byte {
	msgHash := md5.New()
	_, err := msgHash.Write(msg)
	if err != nil {
		panic(err)
	}

	return msgHash.Sum(nil)
}
