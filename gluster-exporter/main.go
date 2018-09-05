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
	"strconv"
	"strings"
	"time"

	"github.com/gluster/gluster-prometheus/gluster-exporter/conf"
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
	volinfo       = flag.String("volinfo", "", "Volume info json file")
	showVersion   = flag.Bool("version", false, "Show the version information")
	config        = flag.String("config", "", "Global config file path")
	collectorsCfg = flag.String("collectors-config", "", "Collectors config file path")

	peerID          string
	defaultInterval time.Duration = 5
	peerIDPattern                 = regexp.MustCompile("[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}")
)

type glusterMetric struct {
	name string
	fn   func()
}

var glusterMetrics []glusterMetric

func registerMetric(name string, fn func()) {
	glusterMetrics = append(glusterMetrics, glusterMetric{name: name, fn: fn})
}

func getPeerID() (string, error) {
	if peerID == "" {
		glusterMgmt, _ := conf.SystemConfig["gluster-mgmt"]
		workdir, _ := conf.SystemConfig["glusterd-dir"]
		keywordID := "UUID"
		filepath := workdir + "/glusterd.info"
		if glusterMgmt == "glusterd2" {
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

	if err := conf.LoadCollectorsConfig(*collectorsCfg); err != "" {
		fmt.Fprintf(os.Stderr, "Loading collectors config failed: %s", err)
		os.Exit(1)
	}
	if err := conf.LoadConfig(*config); err != "" {
		fmt.Fprintf(os.Stderr, "Loading global config failed: %s", err)
		os.Exit(1)
	}

	if *showVersion {
		dumpVersionInfo()
		return
	}

	// start := time.Now()

	for _, m := range glusterMetrics {
		if collectorConf, ok := conf.Collectors[m.name]; ok {
			if collectorConf.Disabled == false {
				go func(m glusterMetric) {
					for {
						m.fn()
						interval := defaultInterval
						if collectorConf.SyncInterval > 0 {
							interval = time.Duration(collectorConf.SyncInterval)
						}
						time.Sleep(time.Duration(time.Second * time.Duration(interval)))
					}
				}(m)
			}
		}
	}

	if len(glusterMetrics) == 0 {
		fmt.Fprintf(os.Stderr, "No Metrics registered, Exiting..\n")
		os.Exit(1)
	}

	metricsPath, _ := conf.SystemConfig["metrics-path"]
	port, _ := conf.SystemConfig["port"]
	nport, _ := strconv.Atoi(port)
	http.Handle(metricsPath, promhttp.Handler())
	err := http.ListenAndServe(fmt.Sprintf(":%d", nport), nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to run exporter\nError: %s", err)
		os.Exit(1)
	}
}
