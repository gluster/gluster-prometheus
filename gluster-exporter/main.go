package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/gluster/gluster-prometheus/gluster-exporter/conf"
	"github.com/gluster/gluster-prometheus/pkg/glusterutils"

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
	showVersion                   = flag.Bool("version", false, "Show the version information")
	config                        = flag.String("config", "", "Global config file path")
	collectorsCfg                 = flag.String("collectors-config", "", "Collectors config file path")
	defaultInterval time.Duration = 5
	glusterConfig   glusterutils.Config
)

type glusterMetric struct {
	name string
	fn   func()
}

var glusterMetrics []glusterMetric

func registerMetric(name string, fn func()) {
	glusterMetrics = append(glusterMetrics, glusterMetric{name: name, fn: fn})
}

func dumpVersionInfo() {
	fmt.Printf("version   : %s\n", ExporterVersion)
	fmt.Printf("git SHA   : %s\n", GitSHA)
	fmt.Printf("go version: %s\n", runtime.Version())
	fmt.Printf("go OS/arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}

func getDefaultGlusterdDir(mgmt string) string {
	if mgmt == "glusterd2" {
		return defaultGlusterd2Workdir
	}
	return defaultGlusterd1Workdir
}

func main() {
	flag.Parse()

	if err := conf.LoadCollectorsConfig(*collectorsCfg); err != "" {
		fmt.Fprintf(os.Stderr, "Loading collectors config failed: %s\n", err)
		os.Exit(1)
	}
	if err := conf.LoadConfig(*config); err != "" {
		fmt.Fprintf(os.Stderr, "Loading global config failed: %s\n", err)
		os.Exit(1)
	}

	if *showVersion {
		dumpVersionInfo()
		return
	}

	// Set the Gluster Configurations used in glusterutils
	glusterConfig.GlusterMgmt = "glusterd"
	mgmt, exists := conf.SystemConfig["gluster-mgmt"]
	if exists {
		glusterConfig.GlusterMgmt = mgmt
	}
	glusterConfig.GlusterdWorkdir = getDefaultGlusterdDir(glusterConfig.GlusterMgmt)
	gddir, exists := conf.SystemConfig["glusterd-dir"]
	if exists {
		glusterConfig.GlusterdWorkdir = gddir
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
