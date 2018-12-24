package main

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gluster/gluster-prometheus/pkg/glusterutils"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

const (
	// GDaemonLabel provides static label to info/error provided from this metrics
	GDaemonLabel = "Gluster_Daemon_Status"
)

func gdStatus(conf *glusterutils.Config) error {
	if conf.GlusterMgmt == glusterutils.MgmtGlusterd {
		if _, err := glusterutils.ExecuteCmd("ps --no-header -ww -o pid,comm -C glusterd"); err != nil {
			return err
		}
	} else if conf.GlusterMgmt == glusterutils.MgmtGlusterd2 {
		if conf.Glusterd2Endpoint == "" {
			return errors.New("[" + GDaemonLabel + "] Empty GD2 Endpoint")
		}
		pingURL := strings.Join([]string{conf.Glusterd2Endpoint, "ping"}, "/")
		resp, err := http.Get(pingURL)
		if err != nil {
			return errors.New("[" + GDaemonLabel + "]" + "Endpoint Get Error: " + err.Error())
		}
		_, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.New("[" + GDaemonLabel + "]" + "Endpoint Read Error: " + err.Error())
		}
	}
	return nil
}

var (
	// metric labels
	gdStatusLbl = []MetricLabel{
		{
			Name: "name",
			Help: "Metric name, for which the status is collected",
		},
		{
			Name: "gdType",
			Help: "Type of gluster daemon / service running (GD1 or GD2)",
		},
		{
			Name: "peerID",
			Help: "Peer ID of the host for which this metric status is updated",
		},
	}
	gExprtrLbl = []MetricLabel{
		{
			Name: "name",
			Help: "Metric Name to be collected",
		},
		{
			Name: "peerID",
			Help: "Peer ID of the host, on which the data is collected",
		},
	}
	glusterDaemonStatus = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "daemon_status",
		Help:      "Status of gluster management daemon (1 = running and 0 = not-running)",
		LongHelp:  "",
		Labels:    gdStatusLbl,
	})
	glusterExporterStatus = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "exporter_status",
		Help:      "Status of gluster exporter (will be set 1 always)",
		LongHelp:  "",
		Labels:    gExprtrLbl,
	})
)

// gDaemonStatus registering function,
// provides the status of services running on the machine
func gDaemonStatus() error {
	var conf *glusterutils.Config
	if gluster == nil {
		return errors.New("[" + GDaemonLabel + "] Unable to get a 'GInterface' object")
	}
	peerID, err := gluster.LocalPeerID()
	if err != nil {
		return errors.New("[" + GDaemonLabel + "] Getting Peer ID failed: " + err.Error())
	}
	conf, err = glusterutils.GDConfigFromInterface(gluster)
	if err != nil {
		return errors.New("[" + GDaemonLabel + "] Error: " + err.Error())
	}
	log.Println("GD Management: ", conf.GlusterMgmt)
	genrlLbls := prometheus.Labels{
		"name":   "Glusterd_Status",
		"gdType": conf.GlusterMgmt,
		"peerID": peerID,
	}
	err = gdStatus(conf)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"peer": peerID,
			"name": "Glusterd_Status",
		}).Errorln("["+GDaemonLabel+"] Error:", err)
		glusterDaemonStatus.With(genrlLbls).Set(float64(0))
	} else {
		glusterDaemonStatus.With(genrlLbls).Set(float64(1))
	}
	genrlLbls = prometheus.Labels{
		"name":   "Gluster_Exporter_Status",
		"peerID": peerID,
	}
	glusterExporterStatus.With(genrlLbls).Set(float64(1))
	return nil
}

func init() {
	registerMetric("gluster_daemon_status", gDaemonStatus)
}
