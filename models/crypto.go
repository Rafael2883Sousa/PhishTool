package models

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"os"
)

func appKey() ([]byte, error) {
	k := os.Getenv("APP_ENCRYPTION_KEY") // 32 bytes (Base64)
	if k == "" {
		return nil, errors.New("APP_ENCRYPTION_KEY not set")
	}
	raw, err := base64.StdEncoding.DecodeString(k)
	if err != nil {
		return nil, err
	}
	if len(raw) != 32 {
		return nil, errors.New("APP_ENCRYPTION_KEY must decode to 32 bytes")
	}
	return raw, nil
}

func EncryptString(plain string) (string, error) {
	key, err := appKey()
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil { return "", err }
	aead, err := cipher.NewGCM(block)
	if err != nil { return "", err }
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil { return "", err }
	ct := aead.Seal(nil, nonce, []byte(plain), nil)
	buf := append(nonce, ct...)
	return "enc:v1:" + base64.StdEncoding.EncodeToString(buf), nil
}

func DecryptString(enc string) (string, error) {
	const pfx = "enc:v1:"
	if len(enc) < len(pfx) || enc[:len(pfx)] != pfx {
		return "", errors.New("invalid cipher text")
	}
	raw, err := base64.StdEncoding.DecodeString(enc[len(pfx):])
	if err != nil { return "", err }
	key, err := appKey()
	if err != nil { return "", err }
	block, err := aes.NewCipher(key)
	if err != nil { return "", err }
	aead, err := cipher.NewGCM(block)
	if err != nil { return "", err }
	if len(raw) < aead.NonceSize() { return "", errors.New("cipher too short") }
	nonce, ct := raw[:aead.NonceSize()], raw[aead.NonceSize():]
	pt, err := aead.Open(nil, nonce, ct, nil)
	if err != nil { return "", err }
	return string(pt), nil
}
