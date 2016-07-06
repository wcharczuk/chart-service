package yahoo

import (
	"bufio"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/blendlabs/go-util"
	"github.com/wcharczuk/chart-service/server/core"
)

// HistoricalPrice is a result from the historical price feed.
type HistoricalPrice struct {
	Date time.Time `json:"date"`

	Open          float64 `json:"open"`
	High          float64 `json:"high"`
	Low           float64 `json:"low"`
	Close         float64 `json:"close"`
	Volume        float64 `json:"volume"`
	AdjustedClose float64 `json:"adjusted_close"`
}

// Populate consumes a csv line.
func (hp *HistoricalPrice) Populate(line string) error {
	parts := strings.Split(line, ",")
	if len(parts) < 7 {
		return errors.New("Invalid line results, cannot continue")
	}

	if parsedDate, err := time.Parse("2006-01-02", parts[0]); err == nil {
		hp.Date = parsedDate
	} else {
		return err
	}
	hp.Open = util.ParseFloat64(parts[1])
	hp.High = util.ParseFloat64(parts[2])
	hp.Low = util.ParseFloat64(parts[3])
	hp.Close = util.ParseFloat64(parts[4])
	hp.Volume = util.ParseFloat64(parts[5])
	hp.AdjustedClose = util.ParseFloat64(parts[6])
	return nil
}

// GetHistoricalPrices gets historical prices for a ticker in the given window.
func GetHistoricalPrices(ticker string, start, end time.Time) ([]HistoricalPrice, error) {
	// http://real-chart.finance.yahoo.com/table.csv?s=GE&d=6&e=6&f=2016&g=d&a=0&b=2&c=1962&ignore=.csv
	var results []HistoricalPrice
	var err error

	res, err := core.NewRequest().AsGet().
		WithURL("http://real-chart.finance.yahoo.com/table.csv").
		WithQueryString("s", ticker).
		WithQueryString("a", fmt.Sprintf("%02d", int(start.Month())-1)).
		WithQueryString("b", util.IntToString(start.Day())).
		WithQueryString("c", util.IntToString(start.Year())).
		WithQueryString("d", fmt.Sprintf("%02d", int(end.Month())-1)).
		WithQueryString("e", util.IntToString(end.Day())).
		WithQueryString("f", util.IntToString(end.Year())).
		WithQueryString("g", "d").
		WithQueryString("ignore", ".csv").
		FetchRawResponse()

	if err != nil {
		return results, err
	}
	defer res.Body.Close()

	isFirstLine := true
	scanner := bufio.NewScanner(res.Body)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		if isFirstLine {
			isFirstLine = false
			continue
		}
		price := &HistoricalPrice{}
		err = price.Populate(scanner.Text())
		if err != nil {
			return results, err
		}
		results = append(results, *price)
	}
	return results, nil
}
