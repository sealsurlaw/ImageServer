package token

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"time"
)

type Tokenizer struct {
	aesgcm cipher.AEAD
}

type TokenData struct {
	ExpiresAt int64  `json:"exp"`
	Filename  string `json:"fnm"`
}

func NewTokenizer(encryptionSecret string) (*Tokenizer, error) {
	key := sha256.Sum256([]byte(encryptionSecret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return &Tokenizer{
		aesgcm: aesgcm,
	}, nil
}

func (t *Tokenizer) CreateToken(filename string, expiresAt *time.Time) (string, error) {
	tokenData := TokenData{
		Filename:  filename,
		ExpiresAt: expiresAt.Unix(),
	}
	tokenBytes := dataToJsonBytes(&tokenData)

	nonce := t.makeNonce()
	encryptedBytes := t.aesgcm.Seal(nil, nonce, tokenBytes, nil)
	encryptedBytes = joinTokenBytes(nonce, encryptedBytes)
	encryptedStr := base64.RawURLEncoding.EncodeToString(encryptedBytes)

	return encryptedStr, nil
}

func (t *Tokenizer) ParseToken(token string) (filename string, expiresAt *time.Time, err error) {
	tokenBytes, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return "", nil, err
	}

	nonce, tokenBytes := t.splitTokenBytes(tokenBytes)

	decryptedBytes, err := t.aesgcm.Open(nil, nonce, tokenBytes, nil)
	if err != nil {
		return "", nil, err
	}

	tokenData := jsonBytesToData(decryptedBytes)

	expires := time.Unix(tokenData.ExpiresAt, 0)

	return tokenData.Filename, &expires, nil
}

func (t *Tokenizer) makeNonce() []byte {
	nonce := make([]byte, t.aesgcm.NonceSize())
	io.ReadFull(rand.Reader, nonce)
	return nonce
}

func (t *Tokenizer) splitTokenBytes(tokenBytes []byte) (nonce []byte, token []byte) {
	return tokenBytes[:t.aesgcm.NonceSize()], tokenBytes[t.aesgcm.NonceSize():]
}

func joinTokenBytes(nonce []byte, tokenBytes []byte) []byte {
	return bytes.Join([][]byte{nonce, tokenBytes}, []byte{})
}

func dataToJsonBytes(tokenData *TokenData) []byte {
	jsonBytes, _ := json.Marshal(tokenData)
	return jsonBytes
}

func jsonBytesToData(jsonBytes []byte) *TokenData {
	tokenData := &TokenData{}
	_ = json.Unmarshal(jsonBytes, tokenData)
	return tokenData
}
