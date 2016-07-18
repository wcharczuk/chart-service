package core

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/wcharczuk/go-chart"
)

// ParseTimeFrame parses a value timeframe.
// Examples include:
// - LTM : last twelve months
// - 6M : last 6 months
// - 1M : last month
// - 1WK : last week (5 business days).
// The following are to be implemented later:
// - 1D : for the day (hourly).
func ParseTimeFrame(value string) (from time.Time, to time.Time, xvf, yvf chart.ValueFormatter, err error) {
	xvf = DateValueFormatter
	yvf = chart.FloatValueFormatter
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
	case "10d":
		from = time.Now().UTC().AddDate(0, 0, -10)
		to = time.Now().UTC()
		return
	case "3d":
		from = time.Now().UTC().AddDate(0, 0, -3)
		to = time.Now().UTC()
		xvf = DateHourValueFormatter
		return
	case "1d":
		from = time.Now().UTC().AddDate(0, 0, -1)
		to = time.Now().UTC()
		xvf = DateHourValueFormatter
		return
	}
	return time.Time{}, time.Time{}, nil, nil, fmt.Errorf("Invalid timeframe value")
}

// DateValueFormatter is a value formatter that takes a date format.
func DateValueFormatter(v interface{}) string {
	return DateValueFormatterWithFormat(v, chart.DefaultDateFormat)
}

// DateHourValueFormatter is a value formatter that takes a date format.
func DateHourValueFormatter(v interface{}) string {
	return DateValueFormatterWithFormat(v, chart.DefaultDateHourFormat)
}

// DateValueFormatterWithFormat is a value formatter that takes a date format.
func DateValueFormatterWithFormat(v interface{}, dateFormat string) string {
	var d time.Time
	if typed, isTyped := v.(time.Time); isTyped {
		d = typed
	} else if typed, isTyped := v.(int64); isTyped {
		d = time.Unix(0, typed)
	} else if typed, isTyped := v.(float64); isTyped {
		d = time.Unix(0, int64(typed))
	}
	asEastern := d.In(GetEasternTimezone())
	return asEastern.Format(dateFormat)
}

var (
	_easternLock sync.Mutex
	_eastern     *time.Location
)

// GetEasternTimezone gets the eastern timezone.
func GetEasternTimezone() *time.Location {
	if _eastern == nil {
		_easternLock.Lock()
		defer _easternLock.Unlock()
		if _eastern == nil {
			_eastern, _ = time.LoadLocation("America/New_York")
		}
	}
	return _eastern
}
