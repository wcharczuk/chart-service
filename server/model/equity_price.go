package model

import (
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	"github.com/blendlabs/spiffy"
	m "github.com/blendlabs/spiffy/migration"
	"github.com/wcharczuk/go-chart"
)

// EquityPrice is a price for an equity at a given time.
type EquityPrice struct {
	EquityID     int       `json:"equity_id" db:"equity_id"`
	TimestampUTC time.Time `json:"timestamp_utc" db:"timestamp_utc"`
	Price        float64   `json:"price" db:"price"`
	Volume       int64     `json:"volume" db:"volume"`
}

// TableName returns the mapped tablename.
func (ep EquityPrice) TableName() string {
	return "equity_price"
}

// Migration returns the migration steps for the model.
func (ep EquityPrice) Migration() m.Migration {
	return m.New(
		"create or update `equity_price`",
		m.Step(
			m.CreateTable,
			m.Body(
				"CREATE TABLE equity_price (equity_id int not null, timestamp_utc timestamp not null, price numeric(18,2), volume bigint);",
				"ALTER TABLE equity_price ADD CONSTRAINT fk_equity_price_equity_id FOREIGN KEY (equity_id) REFERENCES equity(id);",
			),
			"equity_price",
		),
	)
}

// GetEquityPrices gets equity prices in a date range.
func GetEquityPrices(ticker string, txs ...*sql.Tx) ([]EquityPrice, error) {
	var tx *sql.Tx
	if len(txs) > 0 {
		tx = txs[0]
	}

	query := `
	select * from 
		equity_price ep 
		join equity e on e.id = ep.equity_id 
	where
		e.ticker ilike $1
	`
	var prices []EquityPrice
	return prices, spiffy.DefaultDb().QueryInTransaction(query, tx, ticker).OutMany(&prices)
}

// GetEquityPricesByDate gets equity prices in a date range.
func GetEquityPricesByDate(ticker string, start, end time.Time, txs ...*sql.Tx) ([]EquityPrice, error) {
	var tx *sql.Tx
	if len(txs) > 0 {
		tx = txs[0]
	}

	query := `
	select * from 
		equity_price ep 
		join equity e on e.id = ep.equity_id 
	where
		e.ticker ilike $1
		and ep.timestamp_utc > $2 and ep.timestamp_utc < $3
	`
	var prices []EquityPrice
	return prices, spiffy.DefaultDb().QueryInTransaction(query, tx, ticker, start, end).OutMany(&prices)
}

// EquityPrices is an array of EquityPrice
type EquityPrices []EquityPrice

// Prices returns the x,y ranges as []time.Time and []float64
func (ep EquityPrices) Prices() ([]time.Time, []float64) {
	xvalues := make([]time.Time, len(ep))
	yvalues := make([]float64, len(ep))

	for x := 0; x < len(ep); x++ {
		xvalues[x] = ep[x].TimestampUTC
		yvalues[x] = ep[x].Price
	}
	return xvalues, yvalues
}

// PercentChange returns the x,y ranges as []time.Time and []float64
func (ep EquityPrices) PercentChange() ([]time.Time, []float64) {
	xvalues := make([]time.Time, len(ep))
	yvalues := make([]float64, len(ep))

	if len(ep) == 0 {
		return xvalues, yvalues
	}
	firstValue := ep[0].Price
	for x := 0; x < len(ep); x++ {
		xvalues[x] = ep[x].TimestampUTC
		if x > 0 {
			yvalues[x] = chart.PercentDifference(firstValue, ep[x].Price) * 100.0
		}
	}
	return xvalues, yvalues
}

// Len returns the length.
func (ep EquityPrices) Len() int {
	return len(ep)
}

// Swap swaps values.
func (ep EquityPrices) Swap(i, j int) {
	ep[i], ep[j] = ep[j], ep[i]
}

// Less returns if i is before j.
func (ep EquityPrices) Less(i, j int) bool {
	return ep[i].TimestampUTC.Before(ep[j].TimestampUTC)
}

// LastValueAnnotation returns a last value annotation for the prices.
func (ep EquityPrices) LastValueAnnotation(ticker string, vf chart.ValueFormatter) chart.Annotation {
	if len(ep) == 0 {
		return chart.Annotation{}
	}
	lastValue := ep[len(ep)-1].Price
	return chart.Annotation{
		X:     chart.TimeToFloat64(ep[len(ep)-1].TimestampUTC),
		Y:     lastValue,
		Label: fmt.Sprintf("%s %s", ticker, vf(lastValue)),
	}
}

// LastValueAnnotationPercentChange returns a last value annotation for the prices.
func (ep EquityPrices) LastValueAnnotationPercentChange(ticker string, vf chart.ValueFormatter) chart.Annotation {
	if len(ep) == 0 {
		return chart.Annotation{}
	}
	last := ep[len(ep)-1]
	firstValue := ep[0].Price
	lastValue := last.Price
	value := chart.PercentDifference(firstValue, lastValue) * 100.0
	return chart.Annotation{
		X:     chart.TimeToFloat64(last.TimestampUTC),
		Y:     value,
		Label: fmt.Sprintf("%s %s", ticker, vf(value)),
	}
}

func createTestEquityPrice(equityID int, timestamp time.Time, tx *sql.Tx) (*EquityPrice, error) {
	rp := rand.New(rand.NewSource(time.Now().Unix()))
	ep := EquityPrice{
		EquityID:     equityID,
		TimestampUTC: timestamp,
		Price:        rp.Float64() * 1024,
		Volume:       rp.Int63n(10000),
	}
	err := spiffy.DefaultDb().CreateInTransaction(&ep, tx)
	return &ep, err
}
