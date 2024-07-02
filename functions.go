package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"io"
)

func encrypt(plaintext string, key string) (string, error) {
	key32Byte := make([]byte, 32)
	copy(key32Byte[:], []byte(key))
	c, err := aes.NewCipher(key32Byte)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	return hex.EncodeToString(gcm.Seal(nonce, nonce, []byte(plaintext), nil)), nil
}

/*
func decrypt(ciphertextString string, keyString string) (string, error) {
	ciphertext, err := hex.DecodeString(ciphertextString)
	if err != nil {
		return "", err
	}
	key := make([]byte, 32)
	copy(key[:], []byte(keyString))
	c, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	decodedText, err := gcm.Open(nil, nonce, ciphertext, nil)
	return string(decodedText), nil
}
*/
