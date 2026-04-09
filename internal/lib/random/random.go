package random

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
)

var ErrRandomGeneration = errors.New("failed to generate random number")

func NewRandomString(size int) (string, error) {
	charset := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	result := make([]byte, size)
	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", fmt.Errorf("%w: %w", ErrRandomGeneration, err)
		}
		result[i] = charset[n.Uint64()]
	}
	return string(result), nil
}
