package utils

import (
	"crypto/rand"
	"math/big"
)

func RandomInt(max int) int64 {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(max)))
	return n.Int64()
}
