package helper

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"io"
)

const (
	nonceSize = 12
)

func Encrypt(data []byte, encryptionSecret string) ([]byte, error) {
	aesgcm, err := MakeCipher(encryptionSecret)
	if err != nil {
		return nil, err
	}

	nonce := MakeNonce()
	encryptedBytes := aesgcm.Seal(nil, nonce, data, nil)
	encryptedBytes = JoinBytes(nonce, encryptedBytes)

	return encryptedBytes, nil
}

func Decrypt(encryptedBytes []byte, encryptionSecret string) ([]byte, error) {
	aesgcm, err := MakeCipher(encryptionSecret)
	if err != nil {
		return nil, err
	}

	nonce, tokenBytes := SplitJoinedBytes(encryptedBytes)
	data, err := aesgcm.Open(nil, nonce, tokenBytes, nil)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func MakeCipher(encryptionSecret string) (cipher.AEAD, error) {
	key := sha256.Sum256([]byte(encryptionSecret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return aesgcm, nil
}

func MakeNonce() []byte {
	nonce := make([]byte, nonceSize)
	io.ReadFull(rand.Reader, nonce)
	return nonce
}

func SplitJoinedBytes(joinedBytes []byte) (nonce []byte, encryptedBytes []byte) {
	return joinedBytes[:nonceSize], joinedBytes[nonceSize:]
}

func JoinBytes(nonce []byte, encryptedBytes []byte) []byte {
	return bytes.Join([][]byte{nonce, encryptedBytes}, []byte{})
}
