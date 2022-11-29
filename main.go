package main

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

type Metric struct {
	Metric     map[string]interface{} `json:"metric"`
	Values     []float64              `json:"values"`
	Timestamps []int64                `json:"timestamps"`
}

type Config struct {
	url         string
	user        string
	password    string
	exportStart time.Time
	exportEnd   time.Time
	removeStart time.Time
	removeEnd   time.Time
	match       string
	file        string
}

func main() {
	config := parseFlags()
	downloadMetrics(config)
	removeBadPoints(config, 1668067200000, 1668585600000)
	// drop metrics
	dropMetrics(config)
	// reupload
	uploadMetrics(config)
}

// parseFlags reads in user options.
func parseFlags() Config {
	// filename to save (default to metrics.jsonl in current dir)
	url := flag.String("url", "http://localhost:8428", "VM url")
	user := flag.String("user", "", "VM user to authenticate")
	password := flag.String("password", "", "VM user password to authenticate")
	exportStart := flag.String("export-start", "", "Start time for the exported metrics")
	exportEnd := flag.String("export-end", time.Now().Format(time.RFC3339), "End time for the exported metrics")
	removeStart := flag.String("remove-start", "", "Start time of the points to remove from exported metrics")
	removeEnd := flag.String("remove-end", time.Now().Format(time.RFC3339), "End time of the points to remove from exported metrics")
	match := flag.String("match", "", "Metric expression to export from VM")
	file := flag.String("file", "metrics.jsonl", "File path to export metrics to")
	flag.Parse()
	parseTime := func(s *string) time.Time {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			log.Fatal(err)
		}
		return t
	}
	return Config{
		url:         *url,
		user:        *user,
		password:    *password,
		exportStart: parseTime(exportStart),
		exportEnd:   parseTime(exportEnd),
		removeStart: parseTime(removeStart),
		removeEnd:   parseTime(removeEnd),
		match:       *match,
		file:        *file,
	}
}

// downloadMetrics queries VM for metrics and saves them to a file.
func downloadMetrics(config Config) {
	// download all apple health metrics, reduce memory
	u, err := url.Parse(config.url + "/api/v1/export")
	if err != nil {
		log.Fatal(err)
	}
	params := url.Values{}
	params.Set("match[]", config.match)
	params.Set("reduce_mem_usage", "1")
	params.Set("start", config.exportStart.Format(time.RFC3339))
	params.Set("end", config.exportEnd.Format(time.RFC3339))
	u.RawQuery = params.Encode()
	req, err := http.NewRequest("POST", u.String(), nil)
	if err != nil {
		log.Fatal(err)
	}
	req.SetBasicAuth(config.user, config.password)
	client := &http.Client{}
	resp, err := client.Do(req)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(resp.Body)
	out, err := os.Create(config.file)
	if err != nil {
		log.Fatal(err)
	}
	if _, err = io.Copy(out, resp.Body); err != nil {
		log.Fatal(err)
	}
	err = out.Close()
	if err != nil {
		log.Fatal(err)
	}
}

// removeBadPoints removes points from the metrics in the file whose timestamps
// line up with the period requested to be removed.
func removeBadPoints(config Config, badStart, badEnd int64) {
	// remove bad data points
	nf, err := os.Create("update.jsonl")
	if err != nil {
		log.Fatal(err)
	}
	r, err := os.Open(config.file)
	if err != nil {
		log.Fatal(err)
	}
	d := json.NewDecoder(r)
	enc := json.NewEncoder(nf)
	var i = 1
	for {
		i += 1
		println(i)
		var metric Metric
		err = d.Decode(&metric)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		l := len(metric.Values)
		var firstBad int = l
		var lastBad = firstBad
		for i, v := range metric.Timestamps {
			if v > badStart && v < badEnd {
				lastBad = i + 1
				if firstBad == l {
					firstBad = i
				}
			}
		}
		patchedMetric := Metric{
			Metric:     metric.Metric,
			Values:     append(metric.Values[0:firstBad], metric.Values[lastBad:l]...),
			Timestamps: append(metric.Timestamps[0:firstBad], metric.Timestamps[lastBad:l]...),
		}
		if err = enc.Encode(patchedMetric); err != nil {
			log.Fatal(err)
		}
	}

	if err := os.Rename("update.jsonl", config.file); err != nil {
		log.Fatal(err)
	}
	if err := os.Remove("update.jsonl"); err != nil {
		log.Fatal(err)
	}
}

// dropMetrics drops all metrics in VM that match the query.
func dropMetrics(config Config) {
	u, err := url.Parse(config.url + "/api/v1/admin/tsdb/delete_series")
	if err != nil {
		log.Fatal(err)
	}
	params := url.Values{}
	params.Set("match[]", config.match)
	u.RawQuery = params.Encode()
	req, err := http.NewRequest("POST", u.String(), nil)
	if err != nil {
		log.Fatal(err)
	}
	req.SetBasicAuth(config.user, config.password)
	client := &http.Client{}
	resp, err := client.Do(req)
	if resp.StatusCode != 200 {
		log.Fatal(err)
	}
}

// uploadMetrics uploads the changed file of metrics to VM
func uploadMetrics(config Config) {
	r, err := os.Open(config.file)
	if err != nil {
		log.Fatal(err)
	}
	req, err := http.NewRequest("POST", config.url+"/api/v1/import", r)
	if err != nil {
		log.Fatal(err)
	}
	req.SetBasicAuth(config.user, config.password)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != 200 {
		log.Fatalf("Error: status code was not 200, %v", resp)
	}
}
