package main

import (
	"github.com/gluster/gluster-prometheus/pkg/glusterutils"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

var (
	volStatusBrickCountLabels = []MetricLabel{
		{
			Name: "instance",
			Help: "Hostname of the gluster-prometheus instance providing this metric",
		},
		{
			Name: "volume_name",
			Help: "Name of the volume",
		},
	}
	volStatusPerBrickLabels = []MetricLabel{
		{
			Name: "instance",
			Help: "Hostname of the gluster-prometheus instance providing this metric",
		},
		{
			Name: "volume_name",
			Help: "Name of the volume",
		},
		{
			Name: "hostname",
			Help: "Hostname of the brick",
		},
		{
			Name: "peerid",
			Help: "Uuid of the peer hosting this brick",
		},
	}

	volStatusGaugeVecs = make(map[string]*ExportedGaugeVec)

	glusterVolStatusBrickCount = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_status_brick_count",
		Help:      "Number of bricks for volume",
		Labels:    volStatusBrickCountLabels,
	}, &volStatusGaugeVecs)

	glusterVolumeBrickStatus = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_brick_status",
		Help:      "Per node brick status for volume",
		Labels:    volStatusPerBrickLabels,
	}, &volStatusGaugeVecs)

	glusterVolumeBrickPort = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_brick_port",
		Help:      "Brick port",
		Labels:    volStatusPerBrickLabels,
	}, &volStatusGaugeVecs)

	glusterVolumeBrickPid = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_brick_pid",
		Help:      "Brick pid",
		Labels:    volStatusPerBrickLabels,
	}, &volStatusGaugeVecs)

	glusterVolumeBrickTotalInodes = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_brick_total_inodes",
		Help:      "Brick total inodes",
		Labels:    volStatusPerBrickLabels,
	}, &volStatusGaugeVecs)

	glusterVolumeBrickFreeInodes = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_brick_free_inodes",
		Help:      "Brick free inodes",
		Labels:    volStatusPerBrickLabels,
	}, &volStatusGaugeVecs)

	glusterVolumeBrickTotalBytes = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_brick_total_bytes",
		Help:      "Brick total bytes",
		Labels:    volStatusPerBrickLabels,
	}, &volStatusGaugeVecs)

	glusterVolumeBrickFreeBytes = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_brick_free_bytes",
		Help:      "Brick free bytes",
		Labels:    volStatusPerBrickLabels,
	}, &volStatusGaugeVecs)
)

func volumeInfo(gluster glusterutils.GInterface) (err error) {
	// Reset all vecs to not export stale information
	for _, gaugeVec := range volStatusGaugeVecs {
		gaugeVec.RemoveStaleMetrics()
	}

	var peerID string

	if gluster != nil {
		if peerID, err = gluster.LocalPeerID(); err != nil {
			return
		}
	}

	volumes, err := gluster.VolumeStatus()
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"peer": peerID}).Debug("[Gluster Volume Status] Error:", err)
		return err
	}

	// Get monitored gluster instance FQDN
	peers, err := gluster.Peers()
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"peer": peerID}).Debug("[Gluster Volume Status] Error:", err)
		return err
	}
	fqdn := "n/a"
	for _, peer := range peers {
		if peer.ID == peerID {
			// TODO: figure out which value of PeerAddresses may
			// be hostname -- or resolve ip ourselves
			fqdn = peer.PeerAddresses[0]
			break
		}
	}

	for _, vol := range volumes {
		brickCountLabels := prometheus.Labels{
			"instance":    fqdn,
			"volume_name": vol.Name,
		}
		volStatusGaugeVecs[glusterVolStatusBrickCount].Set(brickCountLabels, float64(len(vol.Nodes)))

		for _, node := range vol.Nodes {
			perBrickLabels := prometheus.Labels{
				"instance":    fqdn,
				"volume_name": vol.Name,
				"hostname":    node.Hostname,
				"peerid":      node.PeerID,
			}
			volStatusGaugeVecs[glusterVolumeBrickStatus].Set(perBrickLabels, float64(node.Status))
			volStatusGaugeVecs[glusterVolumeBrickPort].Set(perBrickLabels, float64(node.Port))
			volStatusGaugeVecs[glusterVolumeBrickPid].Set(perBrickLabels, float64(node.PID))

			volStatusGaugeVecs[glusterVolumeBrickTotalInodes].Set(perBrickLabels, float64(node.Gd1InodesTotal))
			volStatusGaugeVecs[glusterVolumeBrickFreeInodes].Set(perBrickLabels, float64(node.Gd1InodesFree))

			volStatusGaugeVecs[glusterVolumeBrickTotalBytes].Set(perBrickLabels, float64(node.Capacity))
			volStatusGaugeVecs[glusterVolumeBrickFreeBytes].Set(perBrickLabels, float64(node.Free))
		}
	}

	return
}

func init() {
	registerMetric("gluster_volume_status", volumeInfo)
}
