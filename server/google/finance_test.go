package google

import (
	"testing"
	"time"

	assert "github.com/blendlabs/go-assert"
	request "github.com/blendlabs/go-request"
)

func TestGetHistoricalPrices(t *testing.T) {
	assert := assert.New(t)
	//                                   http: //www.google.com/finance/historical?enddate=May+15+2017&output=csv&q=spy&startdate=May+15+2016
	request.MockResponseFromFile("GET", "http://www.google.com/finance/historical?enddate=May+15+2017&output=csv&q=spy&startdate=May+15+2016", 200, "./testdata/historical.csv")
	defer request.ClearMockedResponses()

	prices, err := GetHistoricalPrices("spy", time.Date(2016, 05, 15, 0, 0, 0, 0, time.UTC), time.Date(2017, 05, 15, 0, 0, 0, 0, time.UTC))
	assert.Nil(err)
	assert.NotEmpty(prices)
}
