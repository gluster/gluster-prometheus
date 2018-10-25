package main

import (
	"github.com/gluster/gluster-prometheus/pkg/glusterutils"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	volumeLabels = []string{
		"volume",
	}

	glusterVolumeTotalCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "gluster",
			Name:      "volume_total_count",
			Help:      "Total no of volumes",
		},
		[]string{},
	)
	glusterVolumeCreatedCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "gluster",
			Name:      "volume_created_count",
			Help:      "Freshly created no of volumes",
		},
		[]string{},
	)
	glusterVolumeStartedCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "gluster",
			Name:      "volume_started_count",
			Help:      "Total no of started volumes",
		},
		[]string{},
	)
	glusterVolumeBrickCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "gluster",
			Name:      "volume_brick_count",
			Help:      "Total no of bricks in volume",
		},
		volumeLabels,
	)
)

func getVolumeLabels(volname string) prometheus.Labels {
	return prometheus.Labels{
		"volume": volname,
	}
}

func volumeCounts() {
	volumes, err := glusterutils.VolumeInfo(&glusterConfig)
	if err != nil {
		// TODO: Log error
		return
	}

	var volCount, volStartCount, volCreatedCount int

	volCount = len(volumes)
	for _, volume := range volumes {
		switch volume.State {
		case glusterutils.VolumeStateStarted:
			volStartCount++
		case glusterutils.VolumeStateCreated:
			volCreatedCount++
		default:
			// Volume is stopped, nothing to do as the stopped count
			// could be derived using total - started - created
		}
		volBrickCount := 0
		for _, subvol := range volume.SubVolumes {
			volBrickCount += len(subvol.Bricks)
		}
		glusterVolumeBrickCount.With(getVolumeLabels(volume.Name)).Set(float64(volBrickCount))
	}
	glusterVolumeTotalCount.With(prometheus.Labels{}).Set(float64(volCount))
	glusterVolumeStartedCount.With(prometheus.Labels{}).Set(float64(volStartCount))
	glusterVolumeCreatedCount.With(prometheus.Labels{}).Set(float64(volCreatedCount))
}

func init() {
	prometheus.MustRegister(glusterVolumeTotalCount)
	prometheus.MustRegister(glusterVolumeStartedCount)
	prometheus.MustRegister(glusterVolumeCreatedCount)
	prometheus.MustRegister(glusterVolumeBrickCount)

	// Name, Callback Func, Interval Seconds
	isLeader, err := glusterutils.IsLeader(&glusterConfig)
	if !isLeader || err != nil {
		// Unable to find out if the current node is leader
		// Cannot register volume metrics at this node
		// TODO: log the error
		return
	}
	registerMetric("gluster_volume_counts", volumeCounts)
}
