package chart

import "time"

// Series is a entity data set.
type Series interface {
	GetName() string
	GetStyle() Style
	Len() int
	GetValue(index int) Point
}

// TimeSeries is a line on a chart.
type TimeSeries struct {
	Name  string
	Style Style

	XValues []time.Time
	YValues []float64
}

// GetName returns the name of the time series.
func (ts TimeSeries) GetName() string {
	return ts.Name
}

// GetStyle returns the line style.
func (ts TimeSeries) GetStyle() Style {
	return ts.Style
}

// Len returns the number of elements in the series.
func (ts TimeSeries) Len() int {
	return len(ts.XValues)
}

// GetValue gets a value at a given index.
func (ts TimeSeries) GetValue(index int) Point {
	return Point{X: float64(ts.XValues[index].Unix()), Y: ts.YValues[index]}
}

// DataSeries represents a line on a chart.
type DataSeries struct {
	Name  string
	Style Style

	XValues []float64
	YValues []float64
}

// GetName returns the name of the time series.
func (ds DataSeries) GetName() string {
	return ds.Name
}

// GetStyle returns the line style.
func (ds DataSeries) GetStyle() Style {
	return ds.Style
}

// Len returns the number of elements in the series.
func (ds DataSeries) Len() int {
	return len(ds.XValues)
}

// GetValue gets a value at a given index.
func (ds DataSeries) GetValue(index int) Point {
	return Point{X: ds.XValues[index], Y: ds.YValues[index]}
}
