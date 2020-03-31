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

Now you're ready to start downloading your check data. Run this command:

```sh
go run cmd/download/main.go |tee data.csv
```

If all is well, you'll start to see CSV data printed out, one line for each of your checks:

```
My Main Website,https://example.com/,Sites,0,0
...
```

Occasionally you may see a message that the downloader is being rate-limited by the Uptime.com API. That's okay, just wait a few seconds and it will keep trying:

```
rate-limited; sleeping before retry
My Secondary Website,https://subdomain.example.com,Sites,0,0
```

## Analysing data

Once the downloader has finished running, your data will be stored in the file `data.csv`. To produce statistics on it, run:

```
go run cmd/analyse/main.go <data.csv
```

You should see some output like this:

```
Sector: Sites (128 sites)
Outages: Total 38.0 Min 0.0 Max 3.0 Median 0.0 Mean 0.3 Standard deviation 0.7
Downtimes: Total 655891.0 Min 0.0 Max 473418.0 Median 0.0 Mean 5124.1 Standard deviation 41910.5
...
```

## Using the Go package

If you want to do your own analysis, or download additional data, you can write your own Go programs which use the `uptime-reporter` library. This provides some useful functions such as a rate-limit-aware API client, CSV input and output, data types, and so on.

```go
import reporter "github.com/bitfield/uptime-reporter"
```
