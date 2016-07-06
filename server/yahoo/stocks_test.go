package yahoo

import (
	"testing"

	"github.com/blendlabs/go-assert"
)

func TestGenerateStockInfoFormat(t *testing.T) {
	assert := assert.New(t)

	assert.Equal("snn4l1c1c6p2k2vrr6", generateStockInfoFormat())
}

func TestStockInfoParse(t *testing.T) {
	assert := assert.New(t)

	mockLine := `"goog","Alphabet Inc.",N/A,694.49,-4.72,N/A,"-0.68%",N/A,1462616,28.26,20.69
`
	stock := &StockInfo{}
	err := stock.Parse(mockLine)
	assert.Nil(err)
	assert.False(stock.IsZero(), stock.String())
}

func TestStockInfoParseInvalid(t *testing.T) {
	assert := assert.New(t)

	mockLine := `"balls",N/A,N/A,N/A,N/A,N/A,N/A,N/A,N/A,N/A,N/A
`
	stock := &StockInfo{}
	err := stock.Parse(mockLine)
	assert.Nil(err)
	assert.True(stock.IsZero(), stock.String())
}
