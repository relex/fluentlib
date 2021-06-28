package forwardprotocol

import (
	"crypto/sha512"
	"encoding/hex"

	"github.com/relex/gotils/logger"
)

// sha512ToHexdigest computes SHA512 for given string and returns hex
func sha512ToHexdigest(content string) string {
	hasher := sha512.New()
	if _, err := hasher.Write([]byte(content)); err != nil {
		logger.Panic(err)
	}
	hash := hasher.Sum(nil)
	return hex.EncodeToString(hash)
}
