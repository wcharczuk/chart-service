package yahoo

import (
	"testing"

	"github.com/blendlabs/go-assert"
)

func TestGenerateStockInfoFormat(t *testing.T) {
	assert := assert.New(t)

	assert.Equal("snn4l1c1p2vd1d2t1r", generateStockInfoFormat())
}
