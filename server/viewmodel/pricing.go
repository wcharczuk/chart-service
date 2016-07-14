package viewmodel

import (
	"sort"
	"time"

	"github.com/wcharczuk/chart-service/server/model"
	"github.com/wcharczuk/chart-service/server/yahoo"
)

// GetEquityPrices gets pricing data from both yahoo and the database.
func GetEquityPrices(ticker string, start, end time.Time) ([]model.EquityPrice, error) {
	var union []model.EquityPrice
	db, err := model.GetEquityPrices(ticker, start, end)
	if err != nil {
		return union, err
	}
	union = append(union, db...)
	hist, err := yahoo.GetHistoricalPrices(ticker, start, end)
	if err != nil {
		return union, err
	}
	histPrices := yahoo.HistoricalPrices(hist).Prices()
	union = append(union, histPrices...)
	sort.Sort(model.EquityPrices(union))
	return union, nil
}
