package signer

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/pkg/errors"
)

var PrivateKey *rsa.PrivateKey
var PublicKey *rsa.PublicKey

func InitKeys() error {
	if PrivateKey == nil || PublicKey == nil {
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return errors.Wrap(err, "failed to generate RSA key")
		}

		PrivateKey = privateKey
		PublicKey = &privateKey.PublicKey
	}
	return nil
}

func Sign(data []byte) (string, error) {
	hash := sha256.Sum256(data)
	signature, err := rsa.SignPSS(rand.Reader, PrivateKey, crypto.SHA256, hash[:], &rsa.PSSOptions{
		SaltLength: rsa.PSSSaltLengthAuto,
	})
	if err != nil {
		return "", fmt.Errorf("failed to sign: %v", err)
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}

func Verify(data []byte, docSignature string) bool {
	hash := sha256.Sum256(data)

	decodedSignature, err := base64.StdEncoding.DecodeString(docSignature)
	if err != nil {
		return false
	}

	err = rsa.VerifyPSS(PublicKey, crypto.SHA256, hash[:], decodedSignature, &rsa.PSSOptions{
		SaltLength: rsa.PSSSaltLengthAuto,
	})
	if err != nil {
		return false
	}

	return true
}
