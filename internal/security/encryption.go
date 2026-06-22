package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

// Encrypt encrypts a plaintext string using AES-GCM and a hex-encoded key.
func Encrypt(plaintext, keyHex string) (string, error) {
	key, _ := hex.DecodeString(keyHex)
	block, err := aes.NewCipher(key)
	if err != nil { return "", err }

	aesGCM, err := cipher.NewGCM(block)
	if err != nil { return "", err }

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil { return "", err }

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(ciphertext), nil
}

// Decrypt decrypts a hex-encoded ciphertext using AES-GCM.
func Decrypt(ciphertextHex, keyHex string) (string, error) {
	key, _ := hex.DecodeString(keyHex)
	ciphertext, _ := hex.DecodeString(ciphertextHex)

	block, err := aes.NewCipher(key)
	if err != nil { return "", err }

	aesGCM, err := cipher.NewGCM(block)
	if err != nil { return "", err }

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize { return "", fmt.Errorf("ciphertext too short") }

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil { return "", err }

	return string(plaintext), nil
}
