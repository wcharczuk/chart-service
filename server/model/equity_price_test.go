package model

import (
	"testing"
	"time"

	"github.com/blendlabs/go-assert"
	"github.com/blendlabs/spiffy"
)

func TestGetEquityPricesByDate(t *testing.T) {
	assert := assert.New(t)
	tx, err := spiffy.DB().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	eq, err := createTestEquity(tx)
	assert.Nil(err)

	now := time.Now()

	_, err = createTestEquityPrice(eq.ID, now.AddDate(0, 0, -1), tx)
	assert.Nil(err)
	_, err = createTestEquityPrice(eq.ID, now.AddDate(0, 0, -2), tx)
	assert.Nil(err)
	_, err = createTestEquityPrice(eq.ID, now.AddDate(0, 0, -3), tx)
	assert.Nil(err)
	_, err = createTestEquityPrice(eq.ID, now.AddDate(0, 0, -4), tx)
	assert.Nil(err)

	prices, err := GetEquityPricesByDate(eq.Ticker, now.AddDate(0, 0, -3).Add(-1*time.Hour), now, tx)
	assert.Nil(err)
	assert.Len(prices, 3)
}
