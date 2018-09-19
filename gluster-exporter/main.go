package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strings"
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
	glusterMgmt     = flag.String("gluster-mgmt", "glusterd1", "Choice of GlusterD version i.e glusterd1 or glusterd2, Default is glusterd1")
	glusterdWorkdir = flag.String("glusterd-dir", "", "Directory where the local peer info file is stored, Default for glusterd1 is /var/lib/glusterd/ and for glusterd2 is /var/lib/glusterd2/")
	port            = flag.Int("port", 8080, "Exporter Port")
	metricsPath     = flag.String("metrics-path", "/metrics", "Metrics API Path")
	volinfo         = flag.String("volinfo", "", "Volume info json file")
	showVersion     = flag.Bool("version", false, "Show the version information")

	peerID          string
	defaultInterval time.Duration = 5
	peerIDPattern                 = regexp.MustCompile("[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}")
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

func getGlusterdWorkdir() string {
	if *glusterdWorkdir != "" {
		return *glusterdWorkdir
	}
	if *glusterMgmt == "glusterd2" {
		return defaultGlusterd2Workdir
	}
	return defaultGlusterd1Workdir
}

func getPeerID() (string, error) {
	if peerID == "" {
		workdir := getGlusterdWorkdir()
		keywordID := "UUID"
		filepath := workdir + "/glusterd.info"
		if *glusterMgmt == "glusterd2" {
			keywordID = "peer-id"
			filepath = workdir + "/uuid.toml"
		}
		fileStream, err := os.Open(filepath)
		if err != nil {
			return "", err
		}
		defer fileStream.Close()

		scanner := bufio.NewScanner(fileStream)
		for scanner.Scan() {
			lines := strings.Split(scanner.Text(), "\n")
			for _, line := range lines {
				if strings.Contains(line, keywordID) {
					parts := strings.Split(string(line), "=")
					unformattedPeerID := parts[1]
					peerID = peerIDPattern.FindString(unformattedPeerID)
					if peerID == "" {
						return "", errors.New("unable to find peer address")
					}
					return peerID, nil
				}
			}
		}
		return "", errors.New("unable to find peer address")
	}
	return peerID, nil
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
