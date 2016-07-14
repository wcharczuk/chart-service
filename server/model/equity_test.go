package model

import (
	"testing"

	"github.com/blendlabs/go-assert"
	"github.com/blendlabs/spiffy"
)

func TestGetEquityByTicker(t *testing.T) {
	assert := assert.New(t)
	tx, err := spiffy.DefaultDb().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	eq, err := createTestEquity(tx)
	assert.Nil(err)

	eq2, err := GetEquityByTicker(eq.Ticker, tx)
	assert.Nil(err)
	assert.Equal(eq.ID, eq2.ID)
}
