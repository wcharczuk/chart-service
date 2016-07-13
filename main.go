package main

import (
	"log"

	"github.com/blendlabs/go-chronometer"
	"github.com/wcharczuk/chart-service/server"
	"github.com/wcharczuk/chart-service/server/jobs"
	"github.com/wcharczuk/chart-service/server/model"
)

func main() {
	err := server.DBInit()
	if err != nil {
		log.Fatal(err)
	}

	err = model.Migrate()
	if err != nil {
		log.Fatal(err)
	}

	err = chronometer.Default().LoadJob(new(jobs.EquityPriceFetch))
	if err != nil {
		log.Fatal(err)
	}
	chronometer.Default().Start()

	log.Fatal(server.Init().Start())
}
