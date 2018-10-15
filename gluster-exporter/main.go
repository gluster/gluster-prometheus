package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"runtime"
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
	defaultConfFile         = ""
)
var (
	showVersion                   = flag.Bool("version", false, "Show the version information")
	config                        = flag.String("config", defaultConfFile, "Config file path")
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

	exporterConf, err := conf.LoadConfig(*config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Loading global config failed: %s\n", err.Error())
		os.Exit(1)
	}

	if *showVersion {
		dumpVersionInfo()
		return
	}

	// Set the Gluster Configurations used in glusterutils
	glusterConfig.GlusterMgmt = "glusterd"
	if exporterConf.GlobalConf.GlusterMgmt != "" {
		glusterConfig.GlusterMgmt = exporterConf.GlobalConf.GlusterMgmt
		if exporterConf.GlobalConf.GlusterMgmt == "glusterd2" {
			if exporterConf.GlobalConf.GD2RESTEndpoint != "" {
				glusterConfig.Glusterd2Endpoint = exporterConf.GlobalConf.GD2RESTEndpoint
			}
		}
	}
	glusterConfig.GlusterdWorkdir = getDefaultGlusterdDir(glusterConfig.GlusterMgmt)
	if exporterConf.GlobalConf.GlusterdDir != "" {
		glusterConfig.GlusterdWorkdir = exporterConf.GlobalConf.GlusterdDir
	}

	// start := time.Now()

	for _, m := range glusterMetrics {
		if collectorConf, ok := exporterConf.CollectorsConf[m.name]; ok {
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

	metricsPath := exporterConf.GlobalConf.MetricsPath
	port := exporterConf.GlobalConf.Port
	http.Handle(metricsPath, promhttp.Handler())
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to run exporter\nError: %s", err)
		os.Exit(1)
	}
}
