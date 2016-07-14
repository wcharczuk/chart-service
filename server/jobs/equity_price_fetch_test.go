package jobs

import (
	"testing"
	"time"

	"github.com/blendlabs/go-assert"
)

func TestEquityPriceFetchGetNextRunTime(t *testing.T) {
	assert := assert.New(t)

	epf := &EquityPriceFetch{}
	err := epf.ensureTimezone()
	assert.Nil(err)

	beforeWeekday := time.Date(2016, 07, 13, 18, 30, 0, 0, time.UTC)
	next := epf.GetNextRunTime(&beforeWeekday)
	assert.Equal(beforeWeekday.Year(), next.Year())
	assert.Equal(beforeWeekday.Month(), next.Month())
	assert.Equal(beforeWeekday.Day(), next.Day())

	afterWeekday := time.Date(2016, 07, 13, 23, 30, 0, 0, time.UTC)
	next = epf.GetNextRunTime(&afterWeekday)
	assert.Equal(afterWeekday.Year(), next.Year())
	assert.Equal(afterWeekday.Month(), next.Month())
	assert.Equal(afterWeekday.Day()+1, next.Day())

	weekend := time.Date(2016, 07, 16, 18, 30, 0, 0, time.UTC)
	next = epf.GetNextRunTime(&weekend)
	assert.Equal(weekend.Year(), next.Year())
	assert.Equal(weekend.Month(), next.Month())
	assert.Equal(weekend.Day()+2, next.Day())

	lateNightRegression, err := time.Parse(time.RFC3339, "2016-07-14T08:49:19Z")
	assert.Nil(err)
	next = epf.GetNextRunTime(&lateNightRegression)

	assert.Equal(lateNightRegression.Year(), next.Year())
	assert.Equal(lateNightRegression.Month(), next.Month())
	assert.Equal(lateNightRegression.Day(), next.Day())
	assert.Equal(13, next.Hour())
	assert.Equal(30, next.Minute())
}
