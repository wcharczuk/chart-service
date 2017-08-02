package google

import (
	"net/http"
	"time"

	"bufio"
	"bytes"
	"strconv"
	"strings"

	"github.com/wcharczuk/chart-service/server/equity"

	"encoding/json"

	exception "github.com/blendlabs/go-exception"
	logger "github.com/blendlabs/go-logger"
	request "github.com/blendlabs/go-request"
)

func formatDate(t time.Time) string {
	return t.Format("Jan 02 2006")
}

func parseCSVDate(val string) (time.Time, error) {
	return time.Parse("2-Jan-06", val)
}

// GetHistoricalPrices returns historical prices.
func GetHistoricalPrices(ticker string, start, end time.Time) (prices []equity.HistoricalPrice, err error) {
	var response []byte
	var meta *request.ResponseMeta
	response, meta, err = request.New().WithURL("http://www.google.com/finance/historical").
		WithQueryString("q", ticker).
		WithQueryString("startdate", formatDate(start)).
		WithQueryString("enddate", formatDate(end)).
		WithQueryString("output", "csv").
		WithLogger(logger.Default()).
		WithMockProvider(request.MockedResponseInjector).
		BytesWithMeta()

	if err != nil {
		return
	}

	if meta.StatusCode > http.StatusOK {
		err = exception.Newf("non-2xx from google finance")
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

		if pieces[1] != "-" {
			price.Open, err = strconv.ParseFloat(pieces[1], 64)
			if err != nil {
				println("error with:", pieces[1])
				return
			}
		}

		if pieces[2] != "-" {
			price.High, err = strconv.ParseFloat(pieces[2], 64)
			if err != nil {
				println("error with:", pieces[2])
				return
			}
		}

		if pieces[3] != "-" {
			price.Low, err = strconv.ParseFloat(pieces[3], 64)
			if err != nil {
				println("error with:", pieces[3])
				return
			}
		}

		if pieces[4] != "-" {
			price.Close, err = strconv.ParseFloat(pieces[4], 64)
			if err != nil {
				println("error with:", pieces[4])
				return
			}
		}

		if pieces[5] != "-" {
			price.Volume, err = strconv.ParseInt(pieces[5], 10, 64)
			if err != nil {
				println("error with:", pieces[5])
				return
			}
		}
		prices = append(prices, price)
	}
	return
}

// GetCurrentPrices gets current prices.
func GetCurrentPrices(tickers []string) ([]equity.Quote, error) {
	response, meta, err := request.New().WithURL("https://www.google.com/finance/info").
		WithQueryString("q", strings.Join(tickers, ",")).
		WithLogger(logger.Default()).
		WithMockProvider(request.MockedResponseInjector).
		BytesWithMeta()

	if err != nil {
		return nil, err
	}

	if meta.StatusCode > http.StatusOK {
		return nil, exception.Newf("non-2xx returned from google finance")
	}

	var prices []price
	err = json.Unmarshal(response[3:], &prices)
	if err != nil {
		return nil, err
	}

	output := make([]equity.Quote, len(prices))
	for i := 0; i < len(output); i++ {
		output[i] = prices[i].Quote()
	}

	return output, nil
}

type price struct {
	ID                 string `json:"id"`
	Ticker             string `json:"t"`
	Exchange           string `json:"e"`
	Last               string `json:"l"`
	LastCurrent        string `json:"l_cur"`
	LastTradeTime      string `json:"ltt"`
	LastTrade          string `json:"lt_dts"`
	Change             string `json:"c"`
	ChangeFixed        string `json:"c_fix"`
	ChangePercent      string `json:"cp"`
	ChangePercentFixed string `json:"cp_fix"`
}

func (p price) Quote() equity.Quote {
	return equity.Quote{
		Timestamp: dt(p.LastTrade),
		Ticker:    p.Ticker,
		Exchange:  p.Exchange,
		Last:      f64(p.Last),
		Change:    f64(p.ChangeFixed),
		ChangePCT: f64(p.ChangePercentFixed),
	}
}

func f64(v string) float64 {
	out, _ := strconv.ParseFloat(v, 64)
	return out
}

func dt(v string) time.Time {
	t, _ := time.Parse("2006-01-02T15:04:05Z", v)
	return t
}
