package chart

import (
	"fmt"
	"time"
)

// MarketHoursRange is a special type of range that compresses a time range into just the
// market (i.e. NYSE operating hours and days) range.
type MarketHoursRange struct {
	Min time.Time
	Max time.Time

	MarketOpen  time.Time
	MarketClose time.Time

	HolidayProvider HolidayProvider

	ValueFormatter ValueFormatter

	Domain int
}

// IsZero returns if the range is setup or not.
func (mhr MarketHoursRange) IsZero() bool {
	return mhr.Min.IsZero() && mhr.Max.IsZero()
}

// GetMin returns the min value.
func (mhr MarketHoursRange) GetMin() float64 {
	return TimeToFloat64(mhr.Min)
}

// GetMax returns the max value.
func (mhr MarketHoursRange) GetMax() float64 {
	return TimeToFloat64(mhr.GetEffectiveMax())
}

// GetEffectiveMax gets either the close on the max, or the max itself.
func (mhr MarketHoursRange) GetEffectiveMax() time.Time {
	maxClose := Date.On(mhr.MarketClose, mhr.Max)
	if maxClose.After(mhr.Max) {
		return maxClose
	}
	return mhr.Max
}

// SetMin sets the min value.
func (mhr *MarketHoursRange) SetMin(min float64) {
	mhr.Min = Float64ToTime(min)
}

// SetMax sets the max value.
func (mhr *MarketHoursRange) SetMax(max float64) {
	mhr.Max = Float64ToTime(max)
}

// GetDelta gets the delta.
func (mhr MarketHoursRange) GetDelta() float64 {
	min := mhr.GetMin()
	max := mhr.GetMax()
	return max - min
}

// GetDomain gets the domain.
func (mhr MarketHoursRange) GetDomain() int {
	return mhr.Domain
}

// SetDomain sets the domain.
func (mhr *MarketHoursRange) SetDomain(domain int) {
	mhr.Domain = domain
}

// GetHolidayProvider coalesces a userprovided holiday provider and the date.DefaultHolidayProvider.
func (mhr MarketHoursRange) GetHolidayProvider() HolidayProvider {
	if mhr.HolidayProvider == nil {
		return defaultHolidayProvider
	}
	return mhr.HolidayProvider
}

// GetTicks returns the ticks for the range.
// This is to override the default continous ticks that would be generated for the range.
func (mhr *MarketHoursRange) GetTicks(vf ValueFormatter) []Tick {
	dateDiff := Date.Diff(mhr.Min, mhr.Max)
	if dateDiff > 3 {
		return mhr.getTicksByDay(vf)
	}
	return mhr.getTicksByHour(vf)
}

func (mhr MarketHoursRange) getTicksByHour(vf ValueFormatter) []Tick {
	var ticks []Tick

	cursor := Date.On(mhr.MarketClose, mhr.Min)
	maxClose := Date.On(mhr.MarketClose, mhr.Max)

	for cursor.Before(maxClose) {
		if cursor.After(Date.On(mhr.MarketOpen, cursor)) && cursor.Before(Date.On(mhr.MarketClose, cursor)) {
			ticks = append(ticks, Tick{
				Value: TimeToFloat64(cursor),
				Label: vf(cursor),
			})
		} else {
			cursor = Date.NextMarketOpen(cursor, mhr.MarketOpen, mhr.GetHolidayProvider())
		}

		cursor = cursor.Add(1 * time.Hour)
	}

	return ticks
}

func (mhr MarketHoursRange) getTicksByDay(vf ValueFormatter) []Tick {
	var ticks []Tick

	cursor := Date.On(mhr.MarketClose, mhr.Min)
	maxClose := Date.On(mhr.MarketClose, mhr.Max)

	for Date.Before(cursor, maxClose) {
		if Date.IsWeekDay(cursor.Weekday()) && !mhr.GetHolidayProvider()(cursor) {
			ticks = append(ticks, Tick{
				Value: TimeToFloat64(cursor),
				Label: vf(cursor),
			})
		}

		cursor = cursor.AddDate(0, 0, 1)
	}

	endMarketClose := Date.On(mhr.MarketClose, cursor)
	if Date.IsWeekDay(endMarketClose.Weekday()) && !mhr.GetHolidayProvider()(endMarketClose) {
		ticks = append(ticks, Tick{
			Value: TimeToFloat64(endMarketClose),
			Label: vf(endMarketClose),
		})
	}

	return ticks
}

func (mhr MarketHoursRange) String() string {
	return fmt.Sprintf("MarketHoursRange [%s, %s] => %d", mhr.Min.Format(DefaultDateMinuteFormat), mhr.Max.Format(DefaultDateMinuteFormat), mhr.Domain)
}

// Translate maps a given value into the ContinuousRange space.
func (mhr MarketHoursRange) Translate(value float64) int {
	valueTime := Float64ToTime(value)
	valueTimeEastern := valueTime.In(Date.Eastern())
	totalSeconds := Date.CalculateMarketSecondsBetween(mhr.Min, mhr.GetEffectiveMax(), mhr.MarketOpen, mhr.MarketClose, mhr.HolidayProvider)
	valueDelta := Date.CalculateMarketSecondsBetween(mhr.Min, valueTimeEastern, mhr.MarketOpen, mhr.MarketClose, mhr.HolidayProvider)
	translated := int((float64(valueDelta) / float64(totalSeconds)) * float64(mhr.Domain))
	return translated
}
