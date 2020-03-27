package main

import (
	"log"
	"os"

	reporter "github.com/bitfield/uptime-reporter"
)

func main() {
	token := os.Getenv("UPTIME_API_TOKEN")
	if token == "" {
		log.Fatal("no UPTIME_API_TOKEN set")
	}
	r, err := reporter.New(token)
	if err != nil {
		log.Fatal(err)
	}
	sites, err := reporter.ReadCSV(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

}
