package model

import (
	"time"

	m "github.com/blendlabs/spiffy/migration"
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
	return "price_equity"
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
