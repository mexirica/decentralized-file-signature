package signer

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
)

var PrivateKey *rsa.PrivateKey
var PublicKey *rsa.PublicKey

func InitKeys() error {
	if PrivateKey == nil || PublicKey == nil {
		// Gera um novo par de chaves
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return err
		}
		PrivateKey = privateKey
		PublicKey = &privateKey.PublicKey
	}
	return nil
}

func Sign(data []byte) (string, error) {
	hash := sha256.Sum256(data)
	signature, err := rsa.SignPKCS1v15(rand.Reader, PrivateKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}

func Verify(data []byte, docSignature string) bool {
	hash := sha256.Sum256(data)
	signature, err := base64.StdEncoding.DecodeString(docSignature)
	if err != nil {
		return false
	}
	err = rsa.VerifyPKCS1v15(PublicKey, crypto.SHA256, hash[:], signature)
	return err == nil
}
