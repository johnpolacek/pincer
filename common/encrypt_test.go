package common

import (
	"testing"
)

func TestEncrypt(t *testing.T) {
	CIPHER_KEY := []byte("0123456789012345")
	msg := "A quick brown fox jumped over the lazy dog."

	encrypted, _ := Encrypt(CIPHER_KEY, msg)
	decrypted, _ := Decrypt(CIPHER_KEY, encrypted)

	if msg != decrypted {
		t.Error("expected decrypted to match got", decrypted)
	}
}
