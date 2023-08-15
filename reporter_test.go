package reporter_test

import (
	"bytes"
	"io"
	"os"
	"testing"

	reporter "github.com/bitfield/uptime-reporter"
	"github.com/google/go-cmp/cmp"
	"github.com/uptime-com/rest-api-clients/golang/uptime"
)

var (
	WakandanAirlines = reporter.Site{
		Name:         "Wakandan Airlines",
		URL:          "https://wakandanairlines.com",
		Sector:       "Travel",
		Outages:      4,
		DowntimeSecs: 117,
	}
	BankofMetropolis = reporter.Site{
		Name:         "Bank of Metropolis",
		URL:          "https://bankofmetropolis.com",
		Sector:       "Financial Services",
		Outages:      1,
		DowntimeSecs: 3,
	}
	DailyPlanet = reporter.Site{
		Name:         "Daily Planet",
		URL:          "https://dailyplanet.com",
		Sector:       "News & Media",
		Outages:      6,
		DowntimeSecs: 117,
	}
)

func TestReporter(t *testing.T) {
	t.Parallel()
	_, err := reporter.New("dummy API token")
	if err != nil {
		t.Fatal(err)
	}
}

func TestStatsSummary(t *testing.T) {
	t.Parallel()
	tcs := []struct {
		input []float64
		want  reporter.Summary
	}{
		{
			input: []float64{1, 32, 16},
			want: reporter.Summary{
				Sum:    49,
				Min:    1,
				Max:    32,
				Q1:     1,
				Median: 16,
				Q3:     32,
				Mean:   16.333333333333332,
				Dev:    12.657891697365017,
			},
		},
		{
			input: []float64{168, 68, 255, 104, 244, 17, 237, 200, 189, 145},
			want: reporter.Summary{
				Sum:    1627,
				Min:    17,
				Max:    255,
				Q1:     104,
				Median: 178.5,
				Q3:     237,
				Mean:   162.7,
				Dev:    75.31009228516454,
			},
		},
	}
	for _, tc := range tcs {
		got, err := reporter.StatsSummary(tc.input)
		if err != nil {
			t.Fatal(err)
		}
		if !cmp.Equal(tc.want, got) {
			t.Errorf("data: %v %s", tc.input, cmp.Diff(tc.want, got))
		}
	}
}

func TestSiteFromCheck(t *testing.T) {
	t.Parallel()
	inputCheck := uptime.Check{
		PK:      6923,
		Name:    "Wakandan Airlines",
		Address: "https://wakandanairlines.com",
		Tags:    []string{"Travel"},
	}
	inputStats := uptime.CheckStatsResponse{
		Totals: uptime.CheckStatsTotals{
			Outages:      4,
			DowntimeSecs: 117,
		},
	}
	want := reporter.Site{
		ID:           6923,
		Name:         "Wakandan Airlines",
		URL:          "https://wakandanairlines.com",
		Sector:       "Travel",
		Outages:      4,
		DowntimeSecs: 117,
	}
	got := reporter.SiteFromCheck(inputCheck, inputStats)
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestWriteCSV(t *testing.T) {
	t.Parallel()
	wantFile, err := os.Open("testdata/test.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer wantFile.Close()
	want, err := io.ReadAll(wantFile)
	if err != nil {
		t.Fatal(err)
	}
	got := bytes.Buffer{}
	_ = reporter.WriteCSV(&got, WakandanAirlines)
	_ = reporter.WriteCSV(&got, BankofMetropolis)
	_ = reporter.WriteCSV(&got, DailyPlanet)
	if !bytes.Equal(want, got.Bytes()) {
		t.Error(cmp.Diff(string(want), got.String()))
	}
}

func TestReadCSV(t *testing.T) {
	t.Parallel()
	want := reporter.SiteSet{
		WakandanAirlines,
		BankofMetropolis,
		DailyPlanet,
	}
	inFile, err := os.Open("testdata/test.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer inFile.Close()
	got, err := reporter.ReadCSV(inFile)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestBySector(t *testing.T) {
	t.Parallel()
	want := map[string]reporter.SiteSet{
		"Travel":             {WakandanAirlines},
		"Financial Services": {BankofMetropolis},
		"News & Media":       {DailyPlanet},
	}
	inFile, err := os.Open("testdata/test.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer inFile.Close()
	sites, err := reporter.ReadCSV(inFile)
	if err != nil {
		t.Fatal(err)
	}
	got := sites.BySector()
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestSortByDowntime(t *testing.T) {
	t.Parallel()
	input := reporter.SiteSet{
		WakandanAirlines,
		BankofMetropolis,
		DailyPlanet,
	}
	want := reporter.SiteSet{
		DailyPlanet,
		WakandanAirlines,
		BankofMetropolis,
	}
	input.SortByDowntime()
	if !cmp.Equal(want, input) {
		t.Error(cmp.Diff(want, input))
	}
}

func TestFilterDowntimeOver(t *testing.T) {
	t.Parallel()
	input := reporter.SiteSet{
		DailyPlanet,
		WakandanAirlines,
		BankofMetropolis,
	}
	want := reporter.SiteSet{
		BankofMetropolis,
	}
	got := input.FilterDowntimeOver(10)
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}
