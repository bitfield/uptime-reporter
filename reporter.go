package reporter

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/montanaflynn/stats"
	"github.com/uptime-com/rest-api-clients/golang/uptime"
)

// Reporter stores the Uptime.com client configuration.
type Reporter struct {
	client *uptime.Client
}

// Site represents metadata about an Uptime.com check, and can also store data
// on its outages and downtime within a specified period.
type Site struct {
	ID                int
	Name, URL, Sector string
	Outages           int
	DowntimeSecs      int64
}

// New takes an Uptime.com API token and returns a Reporter object which can
// then be used to query the Uptime API.
func New(APIToken string) (Reporter, error) {
	client, err := uptime.NewClient(&uptime.Config{
		Token:            APIToken,
		RateMilliseconds: 8000,
	})
	if err != nil {
		return Reporter{}, err
	}
	return Reporter{client}, nil
}

// GetSiteIDs returns a slice of check IDs, one for each check in the account
// associated with the Reporter's API token.
func (r Reporter) GetSiteIDs() ([]int, error) {
	checks, err := r.client.Checks.ListAll(context.Background(), &uptime.CheckListOptions{PageSize: 250})
	if err != nil {
		return []int{}, fmt.Errorf("listing all checks: %w", err)
	}
	IDs := make([]int, len(checks))
	for i, c := range checks {
		IDs[i] = c.PK
	}
	return IDs, nil
}

// GetDowntimes takes the ID of a check, and two time values indicating the
// start and end of the period to query. It returns a Site object containing
// metadata about the site, plus the number of outages in the period, and the
// total amount of downtime in the period.
func (r Reporter) GetDowntimes(ID int, startDate, endDate string) (Site, error) {
	opt := &uptime.CheckStatsOptions{
		StartDate: startDate,
		EndDate:   endDate,
	}

	stats, _, err := r.client.Checks.Stats(context.Background(), ID, opt)
	if err != nil {
		return Site{}, err
	}
	check, _, err := r.client.Checks.Get(context.Background(), ID)
	if err != nil {
		return Site{}, err
	}
	return SiteFromCheck(*check, *stats), nil
}

// GetDowntimesWithRetry calls GetDowntimes for the given ID. If there is an API
// rate limit error, it sleeps for a while and tries again, and keeps
// trying forever.
func (r Reporter) GetDowntimesWithRetry(ID int, startDate, endDate string) (Site, error) {
	sleep := 5 * time.Second
	for {
		site, err := r.GetDowntimes(ID, startDate, endDate)
		if err == nil {
			return site, nil
		}
		if !strings.Contains(err.Error(), "API_RATE_LIMIT") {
			return Site{}, err
		}
		log.Printf("rate-limited; sleeping %s before retry\n", sleep.String())
		time.Sleep(sleep)
		sleep *= 2
		if sleep > 10*time.Minute {
			sleep = 10 * time.Minute
		}
	}
}

// Summary represents the statistical summary data for a group of Sites.
type Summary struct {
	Sum, Mean, Dev, Min, Max, Median, Q1, Q3 float64
}

// String returns a formatted version of the Summary data suitable for printing.
func (s Summary) String() string {
	return fmt.Sprintf("Total %.1f Min %.1f Max %.1f Median %.1f Mean %.1f Standard deviation %.1f", s.Sum, s.Min, s.Max, s.Median, s.Mean, s.Dev)
}

// StatsSummary takes a dataset of floating-point values and calculates various
// statistical values for them, returning a Summary object containing the
// computed data.
func StatsSummary(input stats.Float64Data) (Summary, error) {
	sum, err := stats.Sum(input)
	if err != nil {
		return Summary{}, err
	}
	mean, err := stats.Mean(input)
	if err != nil {
		return Summary{}, err
	}
	dev, err := stats.StandardDeviation(input)
	if err != nil {
		return Summary{}, err
	}
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
		Sum:    sum,
		Mean:   mean,
		Dev:    dev,
		Max:    max,
		Min:    min,
		Q1:     quartiles.Q1,
		Median: quartiles.Q2,
		Q3:     quartiles.Q3,
	}, nil
}

// SiteFromCheck translates from an uptime.Check and uptime.CheckStats object to
// a Site object containing the data from both objects.
func SiteFromCheck(c uptime.Check, s uptime.CheckStatsResponse) Site {
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

// WriteCSV takes a Site object and prints a CSV-formatted version of it to the
// supplied writer.
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

// ReadCSV reads CSV data representing a group of Sites, one per line, from the
// given input.
func ReadCSV(input io.Reader) (SiteSet, error) {
	var sites SiteSet
	r := csv.NewReader(input)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return SiteSet{}, err
		}
		s := Site{
			Name:   record[0],
			URL:    record[1],
			Sector: record[2],
		}
		outages, err := strconv.Atoi(record[3])
		if err != nil {
			return SiteSet{}, fmt.Errorf("data line %q: %w", record, err)
		}
		s.Outages = outages
		downtime, err := strconv.ParseInt(record[4], 10, 64)
		if err != nil {
			return SiteSet{}, fmt.Errorf("data line %q: %w", record, err)
		}
		s.DowntimeSecs = downtime
		sites = append(sites, s)
	}
	return sites, nil
}

// SiteSet represents a slice of Sites.
type SiteSet []Site

// BySector operates on a SiteSet and returns a map of sectors to sites (that
// is, the map key is the sector name, and the corresponding value is the
// SiteSet of all the sites in that sector).
func (ss SiteSet) BySector() map[string]SiteSet {
	sectors := map[string]SiteSet{}
	for _, s := range ss {
		sectors[s.Sector] = append(sectors[s.Sector], s)
	}
	return sectors
}

// SortByDowntime sorts the SiteSet by downtime, highest first, then by
// outages, most outages first, and then by name alphabetically.
func (ss SiteSet) SortByDowntime() {
	sort.Slice(ss, func(i, j int) bool {
		a, b := ss[i], ss[j]
		if a.DowntimeSecs != b.DowntimeSecs {
			return a.DowntimeSecs > b.DowntimeSecs
		}
		if a.Outages != b.Outages {
			return a.Outages > b.Outages
		}
		return a.Name < b.Name
	})
}

// FilterDowntimeOver returns the set of sites with less than or equal to the
// specified amount of downtime, in seconds.
func (ss SiteSet) FilterDowntimeOver(limit int64) SiteSet {
	filtered := make(SiteSet, 0, len(ss))
	for _, s := range ss {
		if s.DowntimeSecs <= limit {
			filtered = append(filtered, s)
		}
	}
	return filtered
}
