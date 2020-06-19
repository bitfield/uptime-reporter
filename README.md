# uptime-reporter

`uptime-reporter` is a Go package and a set of command-line tools for reporting on website check data from the [Uptime.com](https://uptime.com) monitoring service.

## What does it do?

If you'd like to know how many outages each of your site checks had, and how much downtime each of them had in a given period, for example, `uptime-reporter` can produce that information for you.

It can also produce statistics, such as the average number of outages (both mean and median), average downtime, and so on. If your checks have tags (to classify them by site or business area, for example), `uptime-reporter` can produce statistics for each tag, as well as overall.

## How do I use it?

First, you'll need an Uptime.com account (if you don't have one yet, [sign up for a free trial](https://uptime.com)). Then, you'll need an [API token](https://uptime.com/api/tokens).

Clone this repository if you haven't already. Then run the following command to set your API token, substituting your token for `XXX`:

```sh
export UPTIME_API_TOKEN="XXX"
```

## Downloading data

Now you're ready to start downloading your check data. The downloader will fetch data for the period you specify, so you will need to specify the exact start and end times of this period in the following format: `YYYY-MM-DDTHH:MM:SSZ`.

Run this command, substituting the start and end times of the period you're interested in:

```sh
go run cmd/download/main.go 2019-01-01T00:00:00Z 2020-01-01T00:00:00Z |tee data.csv
```

If all is well, you'll start to see CSV data printed out, one line for each of your checks:

```
My Main Website,https://example.com/,Sites,0,0
...
```

Occasionally you may see a message that the downloader is being rate-limited by the Uptime.com API. That's okay, just wait a few seconds and it will keep trying:

```
rate-limited; sleeping 5s before retry
My Secondary Website,https://subdomain.example.com,Sites,0,0
```

## Resuming interrupted downloads

If you have a large number of checks in your account, and something interrupts your statistics download, don't worry—you don't have to start again from scratch. Just specify the ID of the last check you successfully downloaded, as an extra parameter on the command line:

```
go run cmd/download/main.go 2019-01-01T00:00:00Z 2020-01-01T00:00:00Z 21999 |tee -a data.csv
```

The downloader will skip all the checks up to and including ID 21999, and begin downloading data from the next check onwards.

## Analysing data

Once the downloader has finished running, your data will be stored in the file `data.csv`. To produce statistics on it, run:

```
go run cmd/analyse/main.go <data.csv
```

You should see some output like this:

```
Sector: All (160 sites)
Outages: Total 60.0 Min 0.0 Max 19.0 Median 0.0 Mean 0.4 Standard deviation 1.6
Downtimes: Total 1831856.0 Min 0.0 Max 1160496.0 Median 0.0 Mean 11449.1 Standard deviation 98558.7

Sites with most downtime:
Rank  Name     Outages  Downtime
1     Site A   19       322h21m36s
2     Site B   2        131h30m18s
3     Site C   2        14h4m1s
...
```

## Filtering excessive downtime

The downtime figure for some sites may not be meaningful. For example, if you're analysing 90 days of data and a given site only existed or was only monitored for the last day of that period, it would appear to have 89 days of downtime, which is incorrect.

To eliminate these outliers, you can specify a 'maximum downtime' argument to the analyser:

```
go run cmd/analyse/main.go 100h <data.csv
```

The `100h` here represents a maximum downtime of 100 hours, and sites with more downtime than this will not be included in the statistics or league tables. Try different numbers to find the best one to eliminate the noise from the dataset you're working with.

You can specify the maximum downtime duration as, for example, `5h10m`, or `90s`, or any combination of units supported by [Go's `time` library](https://golang.org/pkg/time/#ParseDuration).

## Using the Go package

If you want to do your own analysis, or download additional data, you can write your own Go programs which use the `uptime-reporter` library. This provides some useful functions such as a rate-limit-aware API client, CSV input and output, data types, and so on.

```go
import reporter "github.com/bitfield/uptime-reporter"
```
