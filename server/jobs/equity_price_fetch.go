package jobs

import (
	"time"

	"github.com/blendlabs/go-chronometer"
	"github.com/blendlabs/go-util"
	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/chart-service/server/google"
	"github.com/wcharczuk/chart-service/server/model"
)

// EquityPriceFetch is the job that fetches stock data.
type EquityPriceFetch struct {
	eastern *time.Location
}

// Name returns the job name.
func (epf *EquityPriceFetch) Name() string {
	return "equity_price_fetch"
}

func (epf *EquityPriceFetch) ensureTimezone() error {
	if epf.eastern == nil {
		eastern, err := time.LoadLocation("America/New_York")
		if err != nil {
			return err
		}
		epf.eastern = eastern
	}
	return nil
}

// Execute is the job body.
func (epf *EquityPriceFetch) Execute(ct *chronometer.CancellationToken) error {
	err := epf.ensureTimezone()
	if err != nil {
		return err
	}
	var stocks []model.Equity
	err = spiffy.Default().GetAll(&stocks)
	if err != nil {
		return err
	}

	ticks := model.Equities(stocks).Tickers()

	infos, err := google.GetCurrentPrices(ticks)
	if err != nil {
		return err
	}

	timestamp := time.Now().UTC()
	//create prices for infos
	for _, i := range infos {
		if epf.tradeDayIsValid(i.Timestamp, timestamp.In(epf.eastern)) {
			equity, err := model.GetEquityByTicker(i.Ticker)
			if err != nil {
				return err
			}
			err = spiffy.Default().Create(model.EquityPrice{
				EquityID:     equity.ID,
				TimestampUTC: timestamp,
				Price:        i.Last,
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Schedule returns the schedule.
func (epf *EquityPriceFetch) Schedule() chronometer.Schedule {
	return epf
}

func (epf *EquityPriceFetch) tradeDayIsValid(lastTradeDate time.Time, current time.Time) bool {
	return lastTradeDate.Day() == current.Day() && lastTradeDate.Month() == current.Month() && lastTradeDate.Year() == current.Year()
}

// GetNextRunTime gets the next runtime for the job.
func (epf *EquityPriceFetch) GetNextRunTime(after *time.Time) *time.Time {
	epf.ensureTimezone()
	if after == nil {
		after = util.OptionalTime(time.Now().UTC())
	}
	afterEastern := after.In(epf.eastern)
	if chronometer.IsWeekendDay(afterEastern.Weekday()) {
		return util.OptionalTime(epf.getNextMarketOpen(afterEastern).UTC())
	}

	open := epf.getMarketOpen(afterEastern)
	close := epf.getMarketClose(afterEastern)

	if afterEastern.Before(open) {
		return util.OptionalTime(open.UTC())
	}

	if afterEastern.After(open) && afterEastern.Before(close) {
		next := afterEastern.Add(15 * time.Minute)
		minuteRemainder := next.Minute() % 15
		if minuteRemainder > 0 {
			next = next.Add(-(time.Duration(minuteRemainder) * time.Minute))
		}

		if next.Before(close) {
			return util.OptionalTime(next.UTC())
		}
		return util.OptionalTime(epf.getNextMarketOpen(next).UTC())
	}
	return util.OptionalTime(epf.getNextMarketOpen(afterEastern).UTC())
}

func (epf *EquityPriceFetch) getMarketOpen(after time.Time) time.Time {
	return time.Date(after.Year(), after.Month(), after.Day(), 9, 30, 0, 0, epf.eastern)
}

func (epf *EquityPriceFetch) getNextMarketOpen(after time.Time) time.Time {
	for cursorDay := 1; cursorDay < 5; cursorDay++ {
		newDay := after.AddDate(0, 0, cursorDay)
		dayOfWeek := newDay.Weekday()
		if chronometer.IsWeekDay(dayOfWeek) {
			return time.Date(newDay.Year(), newDay.Month(), newDay.Day(), 9, 30, 0, 0, epf.eastern)
		}
	}
	return chronometer.Epoch //no but really.
}

func (epf *EquityPriceFetch) getMarketClose(after time.Time) time.Time {
	return time.Date(after.Year(), after.Month(), after.Day(), 16, 0, 0, 0, epf.eastern)
}
