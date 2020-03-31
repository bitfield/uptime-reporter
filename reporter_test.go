package reporter_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	reporter "github.com/bitfield/uptime-reporter"
	"github.com/google/go-cmp/cmp"
	"github.com/uptime-com/rest-api-clients/golang/uptime"
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
	inputStats := uptime.CheckStats{
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
	inputSite := reporter.Site{
		Name:         "Wakandan Airlines",
		URL:          "https://wakandanairlines.com",
		Sector:       "Travel",
		Outages:      4,
		DowntimeSecs: 117,
	}
	wantFile, err := os.Open("testdata/test.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer wantFile.Close()
	want, err := ioutil.ReadAll(wantFile)
	if err != nil {
		t.Fatal(err)
	}
	got := bytes.Buffer{}
	err = reporter.WriteCSV(&got, inputSite)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(want, got.Bytes()) {
		t.Error(cmp.Diff(string(want), got.String()))
	}
}

func TestReadCSV(t *testing.T) {
	t.Parallel()
	want := reporter.Site{
		Name:         "Wakandan Airlines",
		URL:          "https://wakandanairlines.com",
		Sector:       "Travel",
		Outages:      4,
		DowntimeSecs: 117,
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
	if len(sites) < 1 {
		t.Fatal("want 1 site, got 0")
	}
	got := sites[0]
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}
