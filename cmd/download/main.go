package main

import (
	"log"
	"os"
	"time"

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
	IDs, err := r.GetSiteIDs()
	if err != nil {
		log.Fatal(err)
	}
	start, err := time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")
	if err != nil {
		log.Fatal(err)
	}
	end, err := time.Parse(time.RFC3339, "2020-03-31T23:59:59Z")
	if err != nil {
		log.Fatal(err)
	}
	for _, ID := range IDs {
		site, info, err := r.GetDowntimesWithRetry(ID, start, end)
		if err != nil {
			log.Fatal(err)
		}
		reporter.WriteCSV(os.Stdout, site, info)
	}
}
