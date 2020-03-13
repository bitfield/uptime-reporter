package reporter

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/uptime-com/rest-api-clients/golang/uptime"
)

type Reporter struct {
	Client *uptime.Client
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

func AggregateStats(input []uptime.CheckStatsTotals) (total uptime.CheckStatsTotals) {
	for _, s := range input {
		total.Outages += s.Outages
		total.DowntimeSecs += s.DowntimeSecs
	}
	return total
}

type Site struct {
	Name, URL, Sector string
	ID                int64
}

func LoadSites(path string) ([]Site, error) {
	f, err := os.Open(path)
	if err != nil {
		return []Site{}, err
	}
	defer f.Close()
	raw, err := ioutil.ReadAll(f)
	if err != nil {
		return []Site{}, err
	}
	var sites []Site
	err = json.Unmarshal(raw, &sites)
	if err != nil {
		return []Site{}, err
	}
	return sites, nil
}
