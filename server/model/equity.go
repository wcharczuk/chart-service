package model

import (
	"database/sql"

	"github.com/blendlabs/go-util"
	"github.com/blendlabs/spiffy"
	m "github.com/blendlabs/spiffy/migration"
)

// Equity is a security that trades on an exchange.
type Equity struct {
	ID       int    `json:"id" db:"id,pk,serial"`
	Active   bool   `json:"active" db:"active"`
	Name     string `json:"name" db:"name"`
	Ticker   string `json:"ticker" db:"ticker"`
	Exchange string `json:"exchange" db:"exchange"`
}

// TableName returns the table name
func (e Equity) TableName() string {
	return "equity"
}

// IsZero returns if the object has been set or not.
func (e Equity) IsZero() bool {
	return e.ID == 0
}

// Migration returns the migration steps for the model.
func (e Equity) Migration() m.Migration {
	return m.New(
		"create or update `equity`",
		m.Step(
			m.CreateTable,
			m.Body(
				"CREATE TABLE equity (id serial not null, active bool not null, name varchar(255), ticker varchar(32) not null, exchange varchar(32) not null);",
				"ALTER TABLE equity ADD CONSTRAINT pk_equity_id PRIMARY KEY (id);",
				"ALTER TABLE equity ADD CONSTRAINT uk_equity_ticker_exchange UNIQUE (ticker,exchange)",
			),
			"equity",
		),
		m.Step(
			m.CreateColumn,
			m.Body("ALTER TABLE equity ADD exchange varchar(32) not null"),
			"equity",
			"exchange",
		),
		m.Step(
			m.CreateConstraint,
			m.Body("ALTER TABLE equity ADD CONSTRAINT uk_equity_ticker_exchange UNIQUE (ticker,exchange)"),
			"uk_equity_ticker_exchange",
		),
	)
}

// GetEquitiesActive gets active equities.
func GetEquitiesActive(txs ...*sql.Tx) ([]Equity, error) {
	var tx *sql.Tx
	if len(txs) > 0 {
		tx = txs[0]
	}

	query := `select * from equity where active = true`
	var tickers []Equity
	err := spiffy.DB().QueryInTx(query, tx).OutMany(&tickers)
	return tickers, err
}

// SearchEquities searches equities.
func SearchEquities(searchString string, txs ...*sql.Tx) ([]Equity, error) {
	var tx *sql.Tx
	if len(txs) > 0 {
		tx = txs[0]
	}
	query := `select * from equity where name ilike '%'||$1||'%' or ticker ilike '%'||$1||'%';`
	var tickers []Equity
	err := spiffy.DB().QueryInTx(query, tx, searchString).OutMany(&tickers)
	return tickers, err
}

// GetEquityByTicker gets an equity by a ticker.
func GetEquityByTicker(ticker string, txs ...*sql.Tx) (*Equity, error) {
	var tx *sql.Tx
	if len(txs) > 0 {
		tx = txs[0]
	}

	var equity Equity
	query := `select * from equity where ticker ilike $1`
	err := spiffy.DB().QueryInTx(query, tx, ticker).Out(&equity)
	return &equity, err
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

func createTestEquity(tx *sql.Tx) (*Equity, error) {
	equity := Equity{Active: true, Name: "Test Equity", Ticker: util.UUIDv4().ToShortString()}
	err := spiffy.DB().CreateInTx(&equity, tx)
	return &equity, err
}
