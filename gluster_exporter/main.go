package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	port = flag.Int("port", 8080, "Exporter Port")
	metricsPath = flag.String("metrics-path", "/metrics", "Metrics API Path")
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

func getWorkingDir( glusterMgmt *string ) string {
        if( *glusterMgmt == "glusterd2" ) {
                return "/var/lib/glusterd2"
        } else {
                return "/var/lib/glusterd"
        }
}

func getPeerID( gluster_workdir *string ) string {
        if( *gluster_workdir == "/var/lib/glusterd2") {
                peer_id:=parse_toml( *gluster_workdir + "/uuid.toml" )
                return peer_id
        } else {
                peer_id:=parse_info( *gluster_workdir + "/glusterd.info" )
                return peer_id
        }
}

func parse_info(dir string) string{
        gdInfo, err := os.Open(dir)
        if err != nil {
                        fmt.Print(err)
        }
        defer gdInfo.Close()

        scanner := bufio.NewScanner(gdInfo)

        line := 0
        for scanner.Scan(){
                if strings.Contains(scanner.Text(), "UUID"){
                        lines := strings.Split(scanner.Text(), "\n")
                        parts := strings.Split(string(lines[line]), "=")
                        return parts[1]
                }
                line++
        }
        return ""
}

func parse_toml(dir string) string {
        fileptr, err := os.Open(dir)
        if err != nil {
                        fmt.Print(err)
        }
        defer fileptr.Close()

        scanner := bufio.NewScanner(fileptr)
        for scanner.Scan(){
                lines := strings.Split(scanner.Text(), "\n")
                for _ , line := range lines{
                        if strings.Contains(line, "peer-id"){
                                parts := strings.Split(string(line), "=")
                                peerID := parts[1]
                                peer_ID := peerID[2:len(peerID)-1]
                                return peer_ID
                        }
                }
        }
        return ""
}

func getVolInfoFile() string {
	return *volinfo
}

func main() {
	// The following flags are declared in the main because the flags have to be parsed one after the other due to the variable being binded to the flags used as an argument in declaring another flag
	var glusterMgmt = flag.String( "gluster-mgmt", "glusterd1", "Choice of GlusterD version i.e glusterd1 or glusterd2, Default is glusterd1")
	flag.Parse()
	var glusterd_workdir = flag.String( "glusterd-dir", getWorkingDir(glusterMgmt), "Directory where the local peer info file is stored, Default for glusterd1 is /var/lib/glusterd/ and for glusterd2 is /var/lib/glusterd2/")
	flag.Parse()
	var peerid = flag.String( "peerid", getPeerID( glusterd_workdir ), "Gluster Node's peer ID")
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

