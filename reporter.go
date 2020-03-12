package reporter

import (
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
