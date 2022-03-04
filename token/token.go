package token

import (
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/sealsurlaw/gouvre/errs"
	"github.com/sealsurlaw/gouvre/helper"
)

type Tokenizer struct {
	aesgcm cipher.AEAD
}

type TokenData struct {
	Filename         string `json:"f"`
	ExpiresAt        int64  `json:"e,omitempty"`
	EncryptionSecret string `json:"s,omitempty"`
}

func NewTokenizer(encryptionSecret string) (*Tokenizer, error) {
	aesgcm, err := helper.MakeCipher(encryptionSecret)
	if err != nil {
		return nil, err
	}

	return &Tokenizer{
		aesgcm: aesgcm,
	}, nil
}

func (t *Tokenizer) CreateToken(
	filename string,
	expiresAt *time.Time,
	encryptionSecret string,
) (string, error) {
	tokenData := TokenData{
		Filename:         filename,
		EncryptionSecret: encryptionSecret,
	}
	if expiresAt != nil {
		tokenData.ExpiresAt = expiresAt.Unix()
	}
	tokenBytes := dataToJsonBytes(&tokenData)

	nonce := helper.MakeNonce()
	encryptedBytes := t.aesgcm.Seal(nil, nonce, tokenBytes, nil)
	encryptedBytes = helper.JoinBytes(nonce, encryptedBytes)
	encryptedStr := base64.RawURLEncoding.EncodeToString(encryptedBytes)

	return encryptedStr, nil
}

func (t *Tokenizer) ParseToken(
	token string,
) (filename string, expiresAt *time.Time, encryptionSecret string, err error) {
	tokenBytes, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return "", nil, "", err
	}

	nonce, tokenBytes := helper.SplitJoinedBytes(tokenBytes)

	decryptedBytes, err := t.aesgcm.Open(nil, nonce, tokenBytes, nil)
	if err != nil {
		return "", nil, "", err
	}

	tokenData := jsonBytesToData(decryptedBytes)

	var expires time.Time
	if tokenData.ExpiresAt != 0 {
		expires = time.Unix(tokenData.ExpiresAt, 0)
	}

	if time.Now().After(expires) {
		return "", nil, "", errs.ErrTokenExpired
	}

	return tokenData.Filename, &expires, tokenData.EncryptionSecret, nil
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
