package core

import (
	"fmt"
	"strings"
	"time"
)

// ParseTimeFrame parses a value timeframe.
// Examples include:
// - LTM : last twelve months
// - 6M : last 6 months
// - 1M : last month
// - 1WK : last week (5 business days).
// The following are to be implemented later:
// - 1D : for the day (hourly).
func ParseTimeFrame(value string) (from time.Time, to time.Time, err error) {
	switch strings.ToLower(value) {
	case "ltm":
		from = time.Now().UTC().AddDate(0, -12, 0)
		to = time.Now().UTC()
		return
	case "6m":
		from = time.Now().UTC().AddDate(0, -6, 0)
		to = time.Now().UTC()
		return
	case "1m":
		from = time.Now().UTC().AddDate(0, -1, 0)
		to = time.Now().UTC()
		return
	case "1wk":
		from = time.Now().UTC().AddDate(0, 0, -7)
		to = time.Now().UTC()
		return
	}
	return time.Time{}, time.Time{}, fmt.Errorf("Invalid timeframe value")
}
