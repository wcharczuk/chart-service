package viewmodel

import (
	"sort"
	"time"

	"github.com/wcharczuk/chart-service/server/core"
	"github.com/wcharczuk/chart-service/server/model"
	"github.com/wcharczuk/chart-service/server/yahoo"
)

const (
	secondsPerDay = 60 * 60 * 24
)

// GetEquityPricesByDate gets pricing data from both yahoo and the database.
func GetEquityPricesByDate(ticker string, start, end time.Time, useLocalData, useRemoteData bool) ([]model.EquityPrice, error) {
	var union []model.EquityPrice

	if useLocalData {
		db, err := model.GetEquityPricesByDate(ticker, start, end)
		if err != nil {
			return union, err
		}
		db = model.EquityPrices(db).In(core.GetEasternTimezone())
		union = append(union, db...)
	}

	if useRemoteData {
		hist, err := yahoo.GetHistoricalPrices(ticker, start, end)
		if err != nil {
			return union, err
		}
		histPrices := yahoo.HistoricalPrices(hist).Prices()
		union = append(union, histPrices...)
	}
	sort.Sort(model.EquityPrices(union))
	return union, nil
}
