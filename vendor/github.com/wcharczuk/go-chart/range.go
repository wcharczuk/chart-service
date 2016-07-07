package chart

import (
	"math"
	"time"
)

// NewRange returns a new Range
func NewRange(domain int, values ...float64) *Range {
	min, max := MinAndMax(values...)
	return &Range{
		MinValue:    min,
		MaxValue:    max,
		MinMaxDelta: max - min,
		Domain:      domain,
	}
}

// Range represents a continuous range
// of float64 values mapped to a [0...WindowMaxValue]
// interval.
type Range struct {
	MinValue    float64
	MaxValue    float64
	MinMaxDelta float64
	Domain      int
}

// Translate maps a given value into the range space.
// An example would be a 600 px image, with a min of 10 and a max of 100.
// Translate(50) would yield (50.0/90.0)*600 ~= 333.33
func (r Range) Translate(value float64) int {
	finalValue := ((r.MaxValue - value) / r.MinMaxDelta) * float64(r.Domain)
	return int(math.Floor(finalValue))
}

// NewRangeOfTime makes a new range of time with the given time values.
func NewRangeOfTime(domain int, values ...time.Time) *RangeOfTime {
	min, max := MinAndMaxOfTime(values...)
	r := &RangeOfTime{
		MinValue:    min,
		MaxValue:    max,
		MinMaxDelta: max.Unix() - min.Unix(),
		Domain:      domain,
	}
	return r
}

// RangeOfTime represents a timeseries.
type RangeOfTime struct {
	MinValue    time.Time
	MaxValue    time.Time
	MinMaxDelta int64 //unix time difference
	Domain      int
}

// Translate maps a given value into the range space (of time).
// An example would be a 600 px image, with a min of jan-01-2016 and a max of jun-01-2016.
// Translate(may-01-2016) would yield ... something.
func (r RangeOfTime) Translate(value time.Time) int {
	valueDelta := r.MaxValue.Unix() - value.Unix()
	finalValue := (float64(valueDelta) / float64(r.MinMaxDelta)) * float64(r.Domain)
	return int(math.Floor(finalValue))
}
