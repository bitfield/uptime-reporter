package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"text/tabwriter"
	"time"

	reporter "github.com/bitfield/uptime-reporter"
	"github.com/montanaflynn/stats"
)

func main() {
	sites, err := reporter.ReadCSV(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	var sectors = map[string][]reporter.Site{}
	for _, s := range sites {
		sectors[s.Sector] = append(sectors[s.Sector], s)
	}
	for sector, sites := range sectors {
		printSummary(sector, sites)
		sortByDowntime(sites)
		printWorst(sector, sites)
	}
	printSummary("All", sites)
	sortByDowntime(sites)
	printWorst("All", sites)
}

func sortByDowntime(sites []reporter.Site) {
	sort.Slice(sites, func(i, j int) bool {
		a, b := sites[i], sites[j]
		if a.DowntimeSecs != b.DowntimeSecs {
			return a.DowntimeSecs > b.DowntimeSecs
		}
		if a.Outages != b.Outages {
			return a.Outages > b.Outages
		}
		return a.Name < b.Name
	})
}

func printSummary(sector string, sites []reporter.Site) {
	var outages, downtimes stats.Float64Data
	for _, s := range sites {
		outages = append(outages, float64(s.Outages))
		downtimes = append(downtimes, float64(s.DowntimeSecs))
	}
	fmt.Printf("Sector: %s (%d sites)\n", sector, len(outages))
	outageSummary, err := reporter.StatsSummary(outages)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Outages: %s\n", outageSummary)
	downtimeSummary, err := reporter.StatsSummary(downtimes)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Downtimes: %s\n", downtimeSummary)
	fmt.Println()
}

func printWorst(sector string, sites []reporter.Site) {
	fmt.Println("Sites with most downtime:")
	max := 10
	if len(sites) < max {
		max = len(sites)
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Rank\tName\tOutages\tDowntime\n")
	rank := 1
	for _, s := range sites {
		if s.Outages == 0 {
			continue
		}
		if rank >= 10 {
			break
		}
		downtime := time.Duration(s.DowntimeSecs * 1e9)
		fmt.Fprintf(w, "%d\t%s\t%d\t%s\n", rank, s.Name, s.Outages, downtime.String())
		rank++
	}
	fmt.Fprintln(w)
	w.Flush()
}
