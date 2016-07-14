package jobs

import (
	"fmt"
	"time"

	"github.com/blendlabs/go-chronometer"
	"github.com/blendlabs/go-util"
	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/chart-service/server/model"
	"github.com/wcharczuk/chart-service/server/yahoo"
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
	err = spiffy.DefaultDb().GetAll(&stocks)
	if err != nil {
		return err
	}

	ticks := model.Equities(stocks).Tickers()

	infos, err := yahoo.GetStockPrice(ticks)
	if err != nil {
		return err
	}

	timestamp := time.Now().UTC()
	//create prices for infos
	for _, i := range infos {
		if epf.tradeDayIsValid(i.LastTradeDate, timestamp.In(epf.eastern)) {
			equity, err := model.GetEquityByTicker(i.Ticker)
			if err != nil {
				return err
			}
			err = spiffy.DefaultDb().Create(model.EquityPrice{
				EquityID:     equity.ID,
				TimestampUTC: timestamp,
				Price:        i.LastPrice,
				Volume:       i.Volume,
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

func (epf *EquityPriceFetch) tradeDayIsValid(lastTradeDate string, current time.Time) bool {
	parsed, err := time.Parse("01/02/2016", lastTradeDate)
	if err != nil {
		return false
	}
	return parsed.Day() == current.Day() && parsed.Month() == current.Month() && parsed.Year() == current.Year()
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

	if afterEastern.After(open) && afterEastern.Before(close) {
		next := afterEastern.Add(15 * time.Minute)
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
		println(newDay.Format(time.RFC3339), " is a weekend??", newDay.Weekday())

	}
	return chronometer.Epoch //no but really.
}

func (epf *EquityPriceFetch) getMarketClose(after time.Time) time.Time {
	return time.Date(after.Year(), after.Month(), after.Day(), 16, 0, 0, 0, epf.eastern)
}

// OnStart runs before the job body.
func (epf *EquityPriceFetch) OnStart() {
	epf.logf("Job `%s` starting.", epf.Name())
}

// OnComplete runs after the job body.
func (epf *EquityPriceFetch) OnComplete(err error) {
	if err == nil {
		epf.logf("Job `%s` complete.", epf.Name())
	} else {
		epf.logf("Job `%s` failed.", epf.Name())
		epf.error(err)
	}
}

func (epf *EquityPriceFetch) logf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	fmt.Printf("%s %s\n", util.Color(time.Now().UTC().Format(time.RFC3339), util.ColorGray), message)
}

func (epf *EquityPriceFetch) error(err error) {
	message := fmt.Sprintf("%s:\n%v", util.Color("Exception", util.ColorRed), err)
	fmt.Printf("%s %s\n", util.Color(time.Now().UTC().Format(time.RFC3339), util.ColorGray), message)
}
