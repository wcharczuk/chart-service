package equity

import "time"

// Quote is a point in time quote.
type Quote struct {
	Timestamp time.Time
	Ticker    string
	Exchange  string
	Last      float64
	Change    float64 //c_fix
	ChangePCT float64 //cp_fix
}

// IsZero returns if the quote is zero or not.
func (q Quote) IsZero() bool {
	return q.Timestamp.IsZero()
}
