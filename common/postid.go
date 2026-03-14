package common

import (
	"crypto/rand"
	"math/big"
)

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func GeneratePostId() string {
	b := make([]byte, 10)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(base62Chars))))
		if err != nil {
			b[i] = base62Chars[0]
			continue
		}
		b[i] = base62Chars[n.Int64()]
	}
	return string(b)
}
