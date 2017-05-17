package google

import (
	"fmt"
	"net/http"
	"time"

	"bufio"
	"bytes"
	"strconv"
	"strings"

	"github.com/wcharczuk/chart-service/server/equity"

	logger "github.com/blendlabs/go-logger"
	request "github.com/blendlabs/go-request"
)

func formatDate(t time.Time) string {
	return t.Format("Jan 01 2006")
}

func parseCSVDate(val string) (time.Time, error) {
	return time.Parse("2-Jan-06", val)
}

// GetHistoricalPrices returns historical prices.
func GetHistoricalPrices(ticker string, start, end time.Time) (prices []equity.HistoricalPrice, err error) {
	var response []byte
	var meta *request.ResponseMeta
	response, meta, err = request.New().WithURL("http://www.google.com/finance/historical/").
		WithQueryString("q", ticker).
		WithQueryString("startdate", formatDate(start)).
		WithQueryString("enddate", formatDate(end)).
		WithLogger(logger.Default()).
		WithMockProvider(request.MockedResponseInjector).
		BytesWithMeta()

	if err != nil {
		return
	}

	if meta.StatusCode > http.StatusOK {
		err = fmt.Errorf("non-2xx from google finance")
		return
	}

	scanner := bufio.NewScanner(bytes.NewBuffer(response))
	scanner.Scan() // skip the first line

	var line string

	for scanner.Scan() {
		price := equity.HistoricalPrice{}
		line = scanner.Text()

		pieces := strings.Split(line, ",")
		if len(pieces) < 6 {
			continue
		}

		// date, open, high, low, close, volume
		price.Date, err = parseCSVDate(pieces[0])
		if err != nil {
			return
		}

		price.Open, err = strconv.ParseFloat(pieces[1], 64)
		if err != nil {
			return
		}

		price.High, err = strconv.ParseFloat(pieces[2], 64)
		if err != nil {
			return
		}

		price.Low, err = strconv.ParseFloat(pieces[3], 64)
		if err != nil {
			return
		}

		price.Close, err = strconv.ParseFloat(pieces[4], 64)
		if err != nil {
			return
		}

		price.Volume, err = strconv.ParseInt(pieces[5], 10, 64)
		if err != nil {
			return
		}
		prices = append(prices, price)
	}
	return
}

// GetCurrentPrices gets current prices.
func GetCurrentPrices(tickers []string) ([]equity.Quote, error) {
	return nil, nil
}
