package model

import (
	"log"
	"os"
	"testing"

	"github.com/wcharczuk/chart-service/server/core"
)

func TestMain(m *testing.M) {
	err := core.SetupDatabaseContext()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}
