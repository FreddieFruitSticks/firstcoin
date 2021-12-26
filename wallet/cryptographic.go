package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/btcsuite/btcutil/base58"
	"golang.org/x/crypto/ripemd160"
)

type SigHash string
type Base58CheckVersionPrefix byte

const (
	sigHashAll SigHash = "ALL"

	bitcoinAddressVersionPrefix Base58CheckVersionPrefix = 0
)

type Cryptographic struct {
	PublicKey        []byte
	PrivateKey       []byte
	PrivateKeyObject *ecdsa.PrivateKey
	PublicKeyObject  *ecdsa.PublicKey
	FirstcoinAddress []byte
}

func NewCryptographic() *Cryptographic {
	return &Cryptographic{}
}

func (c *Cryptographic) GenerateKeyPair() {
	curve := elliptic.P256()
	// generate key
	privatekey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		fmt.Printf("Cannot generate ECDSA key\n")
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

	hash160 := ConvertPublicKeyToHash160(pemEncodedPubKey)

	c.FirstcoinAddress = []byte(base58.CheckEncode(hash160, 0))
}

func (c *Cryptographic) GenerateSignature(message []byte) []byte {
	msgHashSum := hashMessage(message)

	signature, err := ecdsa.SignASN1(rand.Reader, c.PrivateKeyObject, msgHashSum)
	if err != nil {
		panic(err)
	}

	return signature
}

func splitScriptSig(scriptSig []byte) ([]byte, []byte, error) {
	split := strings.Split(string(scriptSig), fmt.Sprintf("[%s]", sigHashAll))
	if len(split) != 2 {
		return nil, nil, fmt.Errorf("invalid format of scriptSig")
	}

	return []byte(split[0]), []byte(split[1]), nil
}

func verifyPublicKeyIsAddress(address, publicKey []byte) error {
	hash160PublicKey := ConvertPublicKeyToHash160(publicKey)

	result, version, err := base58.CheckDecode(string(address))
	if err != nil {
		return err
	}

	if version != byte(bitcoinAddressVersionPrefix) {
		return fmt.Errorf("Unsupported base58 check version prefix %d", version)
	}

	if !reflect.DeepEqual(result, hash160PublicKey) {
		return fmt.Errorf("sigScript does not unlock scriptPubKey")
	}

	return nil
}

func VerifySignature(scriptSig []byte, address []byte, message []byte) error {
	signature, publicKey, err := splitScriptSig(scriptSig)
	if err != nil {
		return err
	}

	if err := verifyPublicKeyIsAddress(address, publicKey); err != nil {
		return err
	}

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

	verify := ecdsa.VerifyASN1(pubKey, msgHashSum, []byte(signature))
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

func ConvertPublicKeyToHash160(pubKey []byte) []byte {
	sha := sha256.New()
	sha.Write(pubKey)

	hashedMessage := sha.Sum(nil)

	ripe := ripemd160.New()
	ripe.Write(hashedMessage)

	firstcoinAddress := ripe.Sum(nil)

	return firstcoinAddress
}
