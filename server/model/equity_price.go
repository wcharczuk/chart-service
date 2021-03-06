package model

import (
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	"github.com/blendlabs/spiffy"
	m "github.com/blendlabs/spiffy/migration"
	"github.com/wcharczuk/go-chart"
	util "github.com/wcharczuk/go-chart/util"
)

// EquityPrice is a price for an equity at a given time.
type EquityPrice struct {
	EquityID     int       `json:"equity_id" db:"equity_id"`
	TimestampUTC time.Time `json:"timestamp_utc" db:"timestamp_utc"`
	Price        float64   `json:"price" db:"price"`
	Volume       int64     `json:"volume" db:"volume"`

	IsHistorical bool    `json:"is_historical" db:"-"`
	Open         float64 `json:"open" db:"-"`
	Close        float64 `json:"close" db:"-"`
	High         float64 `json:"high" db:"-"`
	Low          float64 `json:"low" db:"-"`
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
	return prices, spiffy.Default().QueryInTx(query, tx, ticker).OutMany(&prices)
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
	return prices, spiffy.Default().QueryInTx(query, tx, ticker, start, end).OutMany(&prices)
}

// EquityPrices is an array of EquityPrice
type EquityPrices []EquityPrice

// In changes the timezone of the prices.
func (ep EquityPrices) In(loc *time.Location) []EquityPrice {
	newEP := make([]EquityPrice, len(ep))
	for x := 0; x < len(ep); x++ {
		p := ep[x]

		newEP[x] = EquityPrice{
			EquityID:     p.EquityID,
			TimestampUTC: p.TimestampUTC.In(loc),
			Price:        p.Price,
			Volume:       p.Volume,
		}
	}
	return newEP
}

// First returns the first element.
func (ep EquityPrices) First() *EquityPrice {
	if len(ep) > 0 {
		return &ep[0]
	}
	return nil
}

// Last returns the last element.
func (ep EquityPrices) Last() *EquityPrice {
	if len(ep) > 0 {
		return &ep[len(ep)-1]
	}
	return nil
}

// Prices returns the x,y ranges as []time.Time and []float64
// NOTE: This changes the ep timestamp timezones to eastern.
func (ep EquityPrices) Prices() ([]time.Time, []float64) {

	xvalues := make([]time.Time, len(ep))
	yvalues := make([]float64, len(ep))

	for x := 0; x < len(ep); x++ {
		xvalues[x] = ep[x].TimestampUTC
		yvalues[x] = ep[x].Price
	}
	return xvalues, yvalues
}

// PercentChange returns the x,y ranges as []time.Time and []float64.
// NOTE: This changes the ep timestamp timezones to eastern.
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
			yvalues[x] = util.Math.PercentDifference(firstValue, ep[x].Price)
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
func (ep EquityPrices) LastValueAnnotation(ticker string, vf chart.ValueFormatter) chart.Value2 {
	if len(ep) == 0 {
		return chart.Value2{}
	}
	lastValue := ep[len(ep)-1].Price
	return chart.Value2{
		XValue: util.Time.ToFloat64(ep[len(ep)-1].TimestampUTC),
		YValue: lastValue,
		Label:  fmt.Sprintf("%s %s", ticker, vf(lastValue)),
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
	err := spiffy.Default().CreateInTx(&ep, tx)
	return &ep, err
}
