package internal

import (
	"math/rand"
	"time"
)

const (
	// rand charset
	randCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

var (
	// define seeded random.
	seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))
)

// GenerateRandomString returns random string having `size` string length.
func GenerateRandomString(size int) string {
	b := make([]byte, size)

	for i, _ := range b {
		b[i] = randCharset[seededRand.Intn(len(randCharset))]
	}

	return string(b)
}
