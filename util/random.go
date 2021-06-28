package util

import (
	crypto_rand "crypto/rand"
	"encoding/binary"
	math_rand "math/rand"

	"github.com/relex/gotils/logger"
)

// SeedRand seeds math/rand with a secure random number
func SeedRand() {
	// from https://stackoverflow.com/a/54491783/3488757
	var b [8]byte
	_, err := crypto_rand.Read(b[:])
	if err != nil {
		logger.Panic("failed to read crypto/rand: ", err)
	}
	math_rand.Seed(int64(binary.LittleEndian.Uint64(b[:])))
}
