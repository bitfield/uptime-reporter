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
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
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
	boxplot(sector, "Outages", plotter.Values(outages))
	downtimeSummary, err := reporter.StatsSummary(downtimes)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Downtimes: %s\n", downtimeSummary)
	fmt.Println()
	boxplot(sector, "Downtimes", plotter.Values(downtimes))
}

func printWorst(sector string, sites []reporter.Site) {
	fmt.Println("Sites with most downtime:")
	max := 10
	if len(sites) < max {
		max = len(sites)
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Rank\tName\tURL\tOutages\tDowntime\n")
	rank := 1
	for _, s := range sites {
		if s.Outages == 0 {
			continue
		}
		if rank > 10 {
			break
		}
		downtime := time.Duration(s.DowntimeSecs * 1e9)
		fmt.Fprintf(w, "%d\t%s\t%s\t%d\t%s\n", rank, s.Name, s.URL, s.Outages, downtime.String())
		rank++
	}
	fmt.Fprintln(w)
	w.Flush()
}

func boxplot(sector, title string, data plotter.Values) {
	// Create the plot and set its title and axis label.
	p, err := plot.New()
	if err != nil {
		panic(err)
	}
	p.Title.Text = sector
	p.Y.Label.Text = title

	// Make boxes for our data and add them to the plot.
	w := vg.Points(20)
	b0, err := plotter.NewBoxPlot(w, 0, data)
	if err != nil {
		panic(err)
	}
	p.Add(b0)
	if err := p.Save(3*vg.Inch, 4*vg.Inch, fmt.Sprintf("%s_%s.png", sector, title)); err != nil {
		panic(err)
	}
}
