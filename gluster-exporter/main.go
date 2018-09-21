package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ExporterVersion and GitSHA
// These are set as flags during build time. The current values are just placeholders
var (
	ExporterVersion         = ""
	GitSHA                  = ""
	defaultGlusterd1Workdir = ""
	defaultGlusterd2Workdir = ""
)

var (
	port                          = flag.Int("port", 8080, "Exporter Port")
	metricsPath                   = flag.String("metrics-path", "/metrics", "Metrics API Path")
	peerid                        = flag.String("peerid", "", "Gluster Node's peer ID")
	volinfo                       = flag.String("volinfo", "", "Volume info json file")
	showVersion                   = flag.Bool("version", false, "Show the version information")
	defaultInterval time.Duration = 5
)

type glusterMetric struct {
	name            string
	fn              func()
	intervalSeconds time.Duration
}

var glusterMetrics []glusterMetric

func registerMetric(name string, fn func(), intervalSeconds int64) {
	glusterMetrics = append(glusterMetrics, glusterMetric{name: name, fn: fn, intervalSeconds: time.Duration(intervalSeconds)})
}

func getPeerID() string {
	return *peerid
}

func getVolInfoFile() string {
	return *volinfo
}

func dumpVersionInfo() {
	fmt.Printf("version   : %s\n", ExporterVersion)
	fmt.Printf("git SHA   : %s\n", GitSHA)
	fmt.Printf("go version: %s\n", runtime.Version())
	fmt.Printf("go OS/arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}

func main() {
	flag.Parse()

	if *showVersion {
		dumpVersionInfo()
		return
	}

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
