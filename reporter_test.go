package reporter_test

import (
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

func TestAggregateStats(t *testing.T) {
	t.Parallel()
	input := []uptime.CheckStatsTotals{
		{
			Outages:      1,
			DowntimeSecs: 217,
		},
		{
			Outages:      16,
			DowntimeSecs: 219,
		},
	}
	want := uptime.CheckStatsTotals{
		Outages:      17,
		DowntimeSecs: 436,
	}
	got := reporter.AggregateStats(input)
	if !cmp.Equal(got, want) {
		t.Error(cmp.Diff(got, want))
	}
}
