package main

import (
	"github.com/gluster/gluster-prometheus/pkg/glusterutils"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

var (
	peerCountMetricLabels = []MetricLabel{
		{
			Name: "instance",
			Help: "Hostname of the gluster-prometheus instance providing this metric",
		},
	}
	peerSCMetricLabels = []MetricLabel{
		{
			Name: "instance",
			Help: "Hostname of the gluster-prometheus instance providing this metric",
		},
		{
			Name: "hostname",
			Help: "Hostname of the peer for which data is collected",
		},
		{
			Name: "uuid",
			Help: "Uuid of the peer for which data is collected",
		},
	}

	peerGaugeVecs []*prometheus.GaugeVec

	glusterPeerCount = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "peer_count",
		Help:      "Number of peers in cluster",
		Labels:    peerCountMetricLabels,
	}, &peerGaugeVecs)

	glusterPeerStatus = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "peer_status",
		Help:      "Peer status info",
		Labels:    peerSCMetricLabels,
	}, &peerGaugeVecs)

	glusterPeerConnected = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "peer_connected",
		Help:      "Peer connection status",
		Labels:    peerSCMetricLabels,
	}, &peerGaugeVecs)
)

func peerInfo(gluster glusterutils.GInterface) (err error) {
	// Reset all vecs to not export stale information
	for _, gaugeVec := range peerGaugeVecs {
		gaugeVec.Reset()
	}

	var peerID string

	if gluster != nil {
		if peerID, err = gluster.LocalPeerID(); err != nil {
			return
		}
	}

	peers, err := gluster.Peers()
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"peer": peerID}).Debug("[Gluster Peers] Error:", err)
		return err
	}

	fqdn := "n/a"
	for _, peer := range peers {
		if peer.ID == peerID {
			// TODO: figure out which value of PeerAddresses may
			// be hostname -- or resolve ip ourselves
			fqdn = peer.PeerAddresses[0]
		}
	}

	peerCountLabels := prometheus.Labels{
		"instance": fqdn,
	}

	glusterPeerCount.With(peerCountLabels).Set(float64(len(peers)))

	var connected int
	for _, peer := range peers {
		peerSCLabels := prometheus.Labels{
			"instance": fqdn,
			"hostname": peer.PeerAddresses[0],
			"uuid":     peer.ID,
		}
		if peer.Online {
			connected = 1
		} else {
			connected = 0
		}
		// Only update glusterPeerStatus when we retrieved a
		// non-negative peer state, i.e. we're running with the GD1
		// backend.
		if peer.Gd1State > -1 {
			glusterPeerStatus.With(peerSCLabels).Set(float64(peer.Gd1State))
		}
		glusterPeerConnected.With(peerSCLabels).Set(float64(connected))
	}

	return
}

func init() {
	registerMetric("gluster_peer_info", peerInfo)
}
