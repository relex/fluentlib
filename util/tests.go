package util

import (
	"os"

	"golang.org/x/exp/slices"
)

// IsTestGenerationMode returns true if we're in test-gen mode while running tests
func IsTestGenerationMode() bool {
	return slices.Contains(os.Args, "gen")
}
