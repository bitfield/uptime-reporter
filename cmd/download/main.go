package main

import (
	"log"
	"os"

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
	for _, ID := range IDs {
		site, err := r.GetDowntimesWithRetry(ID, os.Args[1], os.Args[2])
		if err != nil {
			log.Fatal(err)
		}
		reporter.WriteCSV(os.Stdout, site)
	}
}
