package util

import (
	"os"
)

// IsTestGenerationMode returns true if we're in test-gen mode while running tests
func IsTestGenerationMode() bool {
	return containsString(os.Args, "gen")
}

func containsString(slice []string, target string) bool {
	for _, item := range slice {
		if item == target {
			return true
		}
	}
	return false
}
