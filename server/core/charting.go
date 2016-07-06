package core

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// Points are an array of points.
type Points []Point

func (p Points) String() string {
	var values []string
	for _, v := range p {
		values = append(values, fmt.Sprintf("%d,%d", v.X, v.Y))
	}
	return strings.Join(values, "\n")
}

// Point represents a x,y coordinate.
type Point struct {
	X int
	Y int
}

// NewRange returns a new Range
func NewRange(windowSize, offset int, values ...float64) *Range {
	r := &Range{
		MinValue:       Min(values...),
		MaxValue:       Max(values...),
		WindowMaxValue: windowSize,
		Offset:         offset,
	}
	r.MinMaxDelta = r.MaxValue - r.MinValue
	return r
}

// Range represents a continuous range
// of float64 values mapped to a [0...WindowMaxValue]
// interval.
type Range struct {
	MinValue       float64
	MaxValue       float64
	MinMaxDelta    float64
	WindowMaxValue int
	Offset         int
}

// Translate maps a given value into the range space.
// An example would be a 600 px image, with a min of 10 and a max of 100.
// Translate(50) would yield (50.0/90.0)*600 ~= 333.33
func (r Range) Translate(value float64) int {
	finalValue := ((r.MaxValue - value) / r.MinMaxDelta) * float64(r.WindowMaxValue)
	return int(math.Floor(finalValue)) + r.Offset
}

// NewRangeOfTime makes a new range of time with the given time values.
func NewRangeOfTime(windowSize, offset int, values ...time.Time) *RangeOfTime {
	r := &RangeOfTime{MaxValue: values[0].Unix(), MinValue: values[0].Unix(), WindowMaxValue: windowSize, Offset: offset}
	for _, time := range values {
		unix := time.Unix()
		if r.MinValue > unix {
			r.MinValue = unix
		}
		if r.MaxValue < unix {
			r.MaxValue = unix
		}
	}
	r.MinMaxDelta = r.MaxValue - r.MinValue
	return r
}

// RangeOfTime represents a timeseries.
type RangeOfTime struct {
	MinValue       int64
	MaxValue       int64
	MinMaxDelta    int64 //unix time difference
	WindowMaxValue int
	Offset         int
}

// Translate maps a given value into the range space (of time).
// An example would be a 600 px image, with a min of jan-01-2016 and a max of jun-01-2016.
// Translate(may-01-2016) would yield ... something.
func (r RangeOfTime) Translate(value time.Time) int {
	unixValue := value.Unix()
	finalValue := (float64(r.MaxValue-unixValue) / float64(r.MinMaxDelta)) * float64(r.WindowMaxValue)
	return int(math.Floor(finalValue)) + r.Offset
}
