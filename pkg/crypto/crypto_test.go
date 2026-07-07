package crypto

import (
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	key := "01234567890123456789012345678901" // 32 bytes
	plaintext := "my super secret api key"

	ciphertext, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	if ciphertext == plaintext {
		t.Fatalf("Ciphertext same as plaintext")
	}

	decrypted, err := Decrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("Expected decrypted text to be '%s', got '%s'", plaintext, decrypted)
	}
}

func TestInvalidKeySize(t *testing.T) {
	key := "short"
	_, err := Encrypt("test", key)
	if err != ErrInvalidKeySize {
		t.Errorf("Expected ErrInvalidKeySize, got %v", err)
	}
}
