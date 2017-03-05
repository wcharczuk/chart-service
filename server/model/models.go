package model

import (
	"github.com/blendlabs/spiffy"
	m "github.com/blendlabs/spiffy/migration"
	"github.com/wcharczuk/chart-service/server/core"
)

var models = []spiffy.DatabaseMapped{
	Equity{},
	EquityPrice{},
}

// Migrate applies migrations.
func Migrate() error {
	var migrations []m.Migration
	for _, m := range models {
		if typed, isTyped := m.(core.Migrateable); isTyped {
			migrations = append(migrations, typed.Migration())
		}
	}

	if len(migrations) > 0 {
		runner := m.New("chart-service", migrations...)
		runner.SetLogger(m.NewLogger())
		return runner.Apply(spiffy.DB())
	}
	return nil
}
