package helper

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"io"
)

func EncryptToken(encryptionSecret, plainText string) (string, error) {
	key := sha256.Sum256([]byte(encryptionSecret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}

	encryptedBytes := aesgcm.Seal(nil, nonce, []byte(plainText), nil)
	cipherTextWithNonce := bytes.Join([][]byte{nonce, encryptedBytes}, []byte{})
	encryptedStr := base64.RawURLEncoding.EncodeToString(cipherTextWithNonce)

	return encryptedStr, nil
}

func DecryptToken(encryptionSecret, encryptedStr string) (string, error) {
	key := sha256.Sum256([]byte(encryptionSecret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}

	unencodedBytes, err := base64.RawURLEncoding.DecodeString(encryptedStr)
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	nonceSize := aesgcm.NonceSize()
	nonce := unencodedBytes[:nonceSize]

	decryptedBytes, err := aesgcm.Open(nil, nonce, unencodedBytes[nonceSize:], nil)
	if err != nil {
		panic(err.Error())
	}

	return string(decryptedBytes), nil
}
