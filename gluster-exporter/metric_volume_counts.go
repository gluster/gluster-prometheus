package main

import (
	"github.com/gluster/gluster-prometheus/pkg/glusterutils"
	"github.com/gluster/gluster-prometheus/pkg/glusterutils/glusterconsts"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

var (
	volumeLabels = []MetricLabel{
		clusterIDLabel,
		{
			Name: "volume",
			Help: "Volume Name",
		},
	}

	countLabels = []MetricLabel{
		clusterIDLabel,
	}

	volumeCountGaugeVecs = make(map[string]*ExportedGaugeVec)

	glusterVolumeTotalCount = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_total_count",
		Help:      "Total no of volumes",
		LongHelp:  "",
		Labels:    countLabels,
	}, &volumeCountGaugeVecs)

	glusterVolumeCreatedCount = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_created_count",
		Help:      "Freshly created no of volumes",
		LongHelp:  "",
		Labels:    countLabels,
	}, &volumeCountGaugeVecs)

	glusterVolumeStartedCount = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_started_count",
		Help:      "Total no of started volumes",
		LongHelp:  "",
		Labels:    countLabels,
	}, &volumeCountGaugeVecs)

	glusterVolumeBrickCount = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_brick_count",
		Help:      "Total no of bricks in volume",
		LongHelp:  "",
		Labels:    volumeLabels,
	}, &volumeCountGaugeVecs)

	glusterVolumeSnapshotBrickCountTotal = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_snapshot_brick_count_total",
		Help:      "Total count of snapshots bricks for volume",
		LongHelp:  "",
		Labels:    volumeLabels,
	}, &volumeCountGaugeVecs)

	glusterVolumeSnapshotBrickCountActive = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_snapshot_brick_count_active",
		Help:      "Total active count of snapshots bricks for volume",
		LongHelp:  "",
		Labels:    volumeLabels,
	}, &volumeCountGaugeVecs)

	glusterVolumeUp = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_up",
		Help:      "Volume is started or not (1-started, 0-not started)",
		LongHelp:  "",
		Labels:    volumeLabels,
	}, &volumeCountGaugeVecs)
)

func getVolumeLabels(volname string) prometheus.Labels {
	return prometheus.Labels{
		"cluster_id": clusterID,
		"volume":     volname,
	}
}

func volumeCounts(gluster glusterutils.GInterface) error {
	// Reset all vecs to not export stale information
	for _, gaugeVec := range volumeCountGaugeVecs {
		gaugeVec.RemoveStaleMetrics()
	}

	isLeader, err := gluster.IsLeader()

	if err != nil {
		log.WithError(err).Debug("Unable to find if the current node is leader")
		return err
	}
	if !isLeader {
		return nil
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
		case glusterconsts.VolumeStateStarted:
			up = 1
			volStartCount++
		case glusterconsts.VolumeStateCreated:
			volCreatedCount++
		default:
			// Volume is stopped, nothing to do as the stopped count
			// could be derived using total - started - created
		}
		volumeCountGaugeVecs[glusterVolumeUp].Set(getVolumeLabels(volume.Name), float64(up))
		volBrickCount := 0
		for _, subvol := range volume.SubVolumes {
			volBrickCount += len(subvol.Bricks)
		}
		volumeCountGaugeVecs[glusterVolumeBrickCount].Set(getVolumeLabels(volume.Name), float64(volBrickCount))
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
		volumeCountGaugeVecs[glusterVolumeSnapshotBrickCountTotal].Set(getVolumeLabels(volume.Name), float64(volSnapBrickCountTotal))
		volumeCountGaugeVecs[glusterVolumeSnapshotBrickCountActive].Set(getVolumeLabels(volume.Name), float64(volSnapBrickCountActive))
	}
	volumeCountGaugeVecs[glusterVolumeTotalCount].Set(prometheus.Labels{
		"cluster_id": clusterID,
	}, float64(volCount))
	volumeCountGaugeVecs[glusterVolumeStartedCount].Set(prometheus.Labels{
		"cluster_id": clusterID,
	}, float64(volStartCount))
	volumeCountGaugeVecs[glusterVolumeCreatedCount].Set(prometheus.Labels{
		"cluster_id": clusterID,
	}, float64(volCreatedCount))
	return nil
}

func init() {
	registerMetric("gluster_volume_counts", volumeCounts)
}
