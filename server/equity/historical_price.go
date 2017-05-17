package equity

import (
	"time"

	"github.com/wcharczuk/chart-service/server/model"
)

// HistoricalPrice is a historical equity price.
type HistoricalPrice struct {
	Date   time.Time
	Open   float64
	Close  float64
	High   float64
	Low    float64
	Volume int64
}

// HistoricalPrices is an array of historical prices.
type HistoricalPrices []HistoricalPrice

// Opens returns the open price from the prices
func (hp HistoricalPrices) Opens() (prices []float64) {
	prices = make([]float64, len(hp))
	for i := 0; i < len(hp); i++ {
		prices[i] = hp[i].Open
	}

	return
}

// Closes returns the close price from the prices
func (hp HistoricalPrices) Closes() (prices []float64) {
	prices = make([]float64, len(hp))
	for i := 0; i < len(hp); i++ {
		prices[i] = hp[i].Close
	}

	return
}

// Prices returns the close price from the prices
func (hp HistoricalPrices) Prices() (prices []model.EquityPrice) {
	prices = make([]model.EquityPrice, len(hp))
	for i := 0; i < len(hp); i++ {
		prices[i] = model.EquityPrice{
			IsHistorical: true,
			TimestampUTC: hp[i].Date,
			Price:        hp[i].Close,
			Volume:       hp[i].Volume,

			Open:  hp[i].Open,
			Close: hp[i].Close,
			High:  hp[i].High,
			Low:   hp[i].Low,
		}
	}

	return
}
