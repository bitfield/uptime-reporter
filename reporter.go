package reporter

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/montanaflynn/stats"
	"github.com/uptime-com/rest-api-clients/golang/uptime"
)

type Reporter struct {
	client *uptime.Client
}

func New(APIToken string) (Reporter, error) {
	client, err := uptime.NewClient(&uptime.Config{
		Token:            APIToken,
		RateMilliseconds: 2000,
	})
	if err != nil {
		return Reporter{}, err
	}
	return Reporter{client}, nil
}

func (r Reporter) GetDowntimesWithRetry(ID int, start, end time.Time) (Site, error) {
	for {
		site, err := r.GetDowntimes(ID, start, end)
		if err == nil {
			return site, nil
		}
		if !strings.Contains(err.Error(), "API_RATE_LIMIT") {
			return Site{}, err
		}
		log.Println("rate-limited; sleeping before retry")
		time.Sleep(5 * time.Second)
	}
}

func (r Reporter) GetDowntimes(ID int, start, end time.Time) (Site, error) {
	stats, _, err := r.client.Checks.Stats(context.Background(), ID, start, end)
	if err != nil {
		return Site{}, err
	}
	check, _, err := r.client.Checks.Get(context.Background(), ID)
	if err != nil {
		return Site{}, err
	}
	return SiteFromCheck(*check, *stats), nil
}

type Site struct {
	ID                int
	Name, URL, Sector string
	Outages           int
	DowntimeSecs      int64
}

func (r Reporter) GetSiteIDs() ([]int, error) {
	checks, err := r.client.Checks.ListAll(context.Background(), &uptime.CheckListOptions{PageSize: 1000})
	if err != nil {
		return []int{}, fmt.Errorf("listing all checks: %w", err)
	}
	return IDsFromChecks(checks), nil
}

type Summary struct {
	Min, Max, Median, Q1, Q3 float64
}

func (s Summary) String() string {
	return fmt.Sprintf("Min: %.1f Max: %.1f Median: %.1f Q1: %.1f Q3: %.1f", s.Min, s.Max, s.Median, s.Q1, s.Q3)
}

func StatsSummary(input []float64) (Summary, error) {
	min, err := stats.Min(input)
	if err != nil {
		return Summary{}, err
	}
	max, err := stats.Max(input)
	if err != nil {
		return Summary{}, err
	}
	quartiles, err := stats.Quartile(input)
	if err != nil {
		return Summary{}, err
	}
	return Summary{
		Max:    max,
		Min:    min,
		Q1:     quartiles.Q1,
		Median: quartiles.Q2,
		Q3:     quartiles.Q3,
	}, nil

}

func IDsFromChecks(checks []*uptime.Check) []int {
	IDs := make([]int, len(checks))
	for i, c := range checks {
		IDs[i] = c.PK
	}
	return IDs
}

func SiteFromCheck(c uptime.Check, s uptime.CheckStats) Site {
	site := Site{
		ID:           c.PK,
		Name:         c.Name,
		URL:          c.Address,
		Outages:      s.Totals.Outages,
		DowntimeSecs: s.Totals.DowntimeSecs,
	}
	if len(c.Tags) > 0 {
		site.Sector = c.Tags[0]
	}
	return site
}

func WriteCSV(output io.Writer, site Site) error {
	w := csv.NewWriter(output)
	record := []string{
		site.Name,
		site.URL,
		site.Sector,
		strconv.Itoa(site.Outages),
		strconv.FormatInt(site.DowntimeSecs, 10),
	}
	err := w.Write(record)
	if err != nil {
		return err
	}
	w.Flush()
	return w.Error()
}

func ReadCSV(input io.Reader) ([]Site, error) {
	var sites []Site
	r := csv.NewReader(input)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return []Site{}, err
		}
		s := Site{
			Name:   record[0],
			URL:    record[1],
			Sector: record[2],
		}
		outages, err := strconv.Atoi(record[3])
		if err != nil {
			return []Site{}, err
		}
		s.Outages = outages
		downtime, err := strconv.ParseInt(record[4], 10, 64)
		if err != nil {
			return []Site{}, err
		}
		s.DowntimeSecs = downtime
		sites = append(sites, s)
	}
	return sites, nil
}
