package main

import (
	"log"

	"github.com/wcharczuk/chart-service/server"
)

func main() {
	log.Fatal(server.Init().Start())
}
