package model

import (
	m "github.com/blendlabs/spiffy/migration"
)

// Equity is a security that trades on an exchange.
type Equity struct {
	ID     int    `json:"id" db:"id,pk"`
	Name   string `json:"name" db:"name"`
	Ticker string `json:"ticker" db:"ticker"`
}

// TableName returns the table name
func (e Equity) TableName() string {
	return "equity"
}

// Migration returns the migration steps for the model.
func (e Equity) Migration() m.Migration {
	return m.New(
		"create or update `equity`",
		m.Step(
			m.CreateTable,
			m.Body(
				"CREATE TABLE equity (id serial not null, name varchar(255), ticker varchar(32) not null);",
				"ALTER TABLE equity ADD CONSTRAINT pk_equity_id PRIMARY KEY (id);",
			),
			"equity",
		),
	)
}

// Equities is an array of Equity.
type Equities []Equity

// Tickers returns the tickers for the equities.
func (e Equities) Tickers() []string {
	var ticks []string
	for _, eq := range e {
		ticks = append(ticks, eq.Ticker)
	}
	return ticks
}
