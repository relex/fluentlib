package util

import (
	"time"
)

// TimeToUnixFloat creates Unix epoch seconds from a Time structure
func TimeToUnixFloat(tm time.Time) float64 { // xx:inline
	return float64(tm.UnixNano()) / 1000000000.0
}
