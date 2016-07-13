package yahoo

import (
	"testing"

	"github.com/blendlabs/go-assert"
)

func TestGenerateStockInfoFormat(t *testing.T) {
	assert := assert.New(t)

	assert.Equal("snn4l1c1c6p2k2vd1d2t1rr6", generateStockInfoFormat())
}
