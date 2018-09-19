package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"
	"os"
	"strings"
	"bufio"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	port = flag.Int("port", 8080, "Exporter Port")
	metricsPath = flag.String("metrics-path", "/metrics", "Metrics API Path")
	peerid = flag.String("peerid", getPeerID(), "Gluster Node's peer ID")
	volinfo = flag.String("volinfo", "", "Volume info json file")

	defaultInterval time.Duration = 5
)


type glusterMetric struct {
	name string
	fn func()
	intervalSeconds time.Duration
}

var glusterMetrics []glusterMetric

func registerMetric(name string, fn func(), intervalSeconds int64) {
	glusterMetrics = append(glusterMetrics, glusterMetric{name: name, fn: fn, intervalSeconds: time.Duration(intervalSeconds)})
}

func getPeerID() string {
	gdInfo, err := os.Open("/var/lib/glusterd/glusterd.info")
	if err != nil {
			fmt.Print(err)
	}
	defer gdInfo.Close()

	scanner := bufio.NewScanner(gdInfo)
	uuid := nil
	line := 0
	for scanner.Scan(){
		if strings.Contains(scanner.Text(), "UUID"){
			lines := strings.Split(scanner.Text(), "\n")
			parts := strings.Split(string(lines[line]), "=")
			uuid := parts[1]
			return uuid
		}
		line++
	}
	return uuid
}

func getVolInfoFile() string {
	return *volinfo
}

func main() {
	flag.Parse()

	// start := time.Now()

	for _, m := range glusterMetrics {
		go func(m glusterMetric) {
			for {
				m.fn()
				interval := defaultInterval
				if m.intervalSeconds > 0 {
					interval = m.intervalSeconds
				}
				time.Sleep(time.Duration(time.Second * time.Duration(interval)))
			}
		}(m)
	}

	if len(glusterMetrics) == 0 {
		fmt.Fprintf(os.Stderr, "No Metrics registered, Exiting..\n")
		os.Exit(1)
	}
	
	http.Handle(*metricsPath, promhttp.Handler())
	err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to run exporter\nError: %s", err)
		os.Exit(1)
	}
}

