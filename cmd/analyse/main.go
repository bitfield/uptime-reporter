package main

import (
	"fmt"
	"log"
	"os"

	reporter "github.com/bitfield/uptime-reporter"
	"github.com/montanaflynn/stats"
)

func main() {
	sites, err := reporter.ReadCSV(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	outages := map[string]stats.Float64Data{}
	downtimes := map[string]stats.Float64Data{}
	for _, s := range sites {
		outages[s.Sector] = append(outages[s.Sector], float64(s.Outages))
		outages["All"] = append(outages["All"], float64(s.Outages))
		downtimes[s.Sector] = append(downtimes[s.Sector], float64(s.DowntimeSecs))
		downtimes["All"] = append(downtimes["All"], float64(s.DowntimeSecs))
	}
	for sector := range outages {
		fmt.Printf("Sector: %s (%d sites)\n", sector, len(outages[sector]))
		outageSummary, err := reporter.StatsSummary(outages[sector])
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Outages: %s\n", outageSummary)
		downtimeSummary, err := reporter.StatsSummary(downtimes[sector])
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Downtimes: %s\n", downtimeSummary)
		fmt.Println()
	}
}
