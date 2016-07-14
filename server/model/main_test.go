package model

import (
	"os"
	"testing"

	"github.com/wcharczuk/chart-service/server/core"
)

func TestMain(m *testing.M) {
	core.DBInit()
	os.Exit(m.Run())
}
