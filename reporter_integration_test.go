// +build integration

package reporter_test

import (
	"log"
	"os"
	"testing"
	"time"

	reporter "github.com/bitfield/uptime-reporter"
)

func TestIntegration(t *testing.T) {
	t.Parallel()
	token := os.Getenv("UPTIME_API_TOKEN")
	if token == "" {
		t.Fatal("no UPTIME_API_TOKEN set")
	}
	r, err := reporter.New(token)
	if err != nil {
		t.Fatal(err)
	}
	IDs, err := r.GetSiteIDs()
	if err != nil {
		log.Fatal(err)
	}
	end := time.Now()
	start := end.Add(-24 * time.Hour)
	site, err := r.GetDowntimesWithRetry(IDs[0], start, end)
	if err != nil {
		t.Fatal(err)
	}
	if site.ID != IDs[0] {
		t.Fatalf("want ID %d, got %d", IDs[0], site.ID)
	}
}
