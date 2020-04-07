package main

import (
	"log"
	"os"
	"time"

	reporter "github.com/bitfield/uptime-reporter"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatalf("Usage: %s START_DATE END_DATE", os.Args[0])
	}
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
	// start, err := time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")
	start, err := time.Parse(time.RFC3339, os.Args[1])
	if err != nil {
		log.Fatalf("date must be in RFC339 format ('2020-01-01T00:00:00Z'): %s", os.Args[1])
	}
	end, err := time.Parse(time.RFC3339, os.Args[2])
	if err != nil {
		log.Fatalf("date must be in RFC339 format ('2020-01-01T00:00:00Z'): %s", os.Args[2])
	}
	for _, ID := range IDs {
		site, err := r.GetDowntimesWithRetry(ID, start, end)
		if err != nil {
			log.Fatal(err)
		}
		reporter.WriteCSV(os.Stdout, site)
	}
}
