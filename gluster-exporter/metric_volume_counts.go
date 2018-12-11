package main

import (
	"github.com/gluster/gluster-prometheus/pkg/glusterutils"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

var (
	volumeLabels = []MetricLabel{
		{
			Name: "volume",
			Help: "Volume Name",
		},
	}

	glusterVolumeTotalCount = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_total_count",
		Help:      "Total no of volumes",
		LongHelp:  "",
		Labels:    []MetricLabel{},
	})

	glusterVolumeCreatedCount = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_created_count",
		Help:      "Freshly created no of volumes",
		LongHelp:  "",
		Labels:    []MetricLabel{},
	})

	glusterVolumeStartedCount = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_started_count",
		Help:      "Total no of started volumes",
		LongHelp:  "",
		Labels:    []MetricLabel{},
	})

	glusterVolumeBrickCount = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_brick_count",
		Help:      "Total no of bricks in volume",
		LongHelp:  "",
		Labels:    volumeLabels,
	})

	glusterVolumeSnapshotBrickCountTotal = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_snapshot_brick_count_total",
		Help:      "Total count of snapshots bricks for volume",
		LongHelp:  "",
		Labels:    volumeLabels,
	})

	glusterVolumeSnapshotBrickCountActive = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_snapshot_brick_count_active",
		Help:      "Total active count of snapshots bricks for volume",
		LongHelp:  "",
		Labels:    volumeLabels,
	})

	glusterVolumeUp = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_up",
		Help:      "Volume is started or not (1-started, 0-not started)",
		LongHelp:  "",
		Labels:    volumeLabels,
	})
)

func getVolumeLabels(volname string) prometheus.Labels {
	return prometheus.Labels{
		"volume": volname,
	}
}

func volumeCounts() error {
	isLeader, err := gluster.IsLeader()
	if !isLeader || err != nil {
		// Unable to find out if the current node is leader
		// Cannot register volume metrics at this node
		log.WithError(err).Error("Unable to find if the current node is leader")
		return err
	}
	volumes, err := gluster.VolumeInfo()
	if err != nil {
		return err
	}
	snapshots, err := gluster.Snapshots()
	if err != nil {
		return err
	}

	var volCount, volStartCount, volCreatedCount int

	volCount = len(volumes)
	for _, volume := range volumes {
		up := 0
		switch volume.State {
		case glusterutils.VolumeStateStarted:
			up = 1
			volStartCount++
		case glusterutils.VolumeStateCreated:
			volCreatedCount++
		default:
			// Volume is stopped, nothing to do as the stopped count
			// could be derived using total - started - created
		}
		glusterVolumeUp.With(getVolumeLabels(volume.Name)).Set(float64(up))
		volBrickCount := 0
		for _, subvol := range volume.SubVolumes {
			volBrickCount += len(subvol.Bricks)
		}
		glusterVolumeBrickCount.With(getVolumeLabels(volume.Name)).Set(float64(volBrickCount))
		volSnapBrickCountTotal := 0
		volSnapBrickCountActive := 0
		for _, snap := range snapshots {
			if volume.Name == snap.VolumeName {
				volSnapBrickCountTotal += volBrickCount
				if snap.Started {
					volSnapBrickCountActive += volBrickCount
				}
			}
		}
		glusterVolumeSnapshotBrickCountTotal.With(getVolumeLabels(volume.Name)).Set(float64(volSnapBrickCountTotal))
		glusterVolumeSnapshotBrickCountActive.With(getVolumeLabels(volume.Name)).Set(float64(volSnapBrickCountActive))
	}
	glusterVolumeTotalCount.With(prometheus.Labels{}).Set(float64(volCount))
	glusterVolumeStartedCount.With(prometheus.Labels{}).Set(float64(volStartCount))
	glusterVolumeCreatedCount.With(prometheus.Labels{}).Set(float64(volCreatedCount))
	return nil
}

func init() {
	registerMetric("gluster_volume_counts", volumeCounts)
}
