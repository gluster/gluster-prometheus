package main

import (
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

var (
	volumeHealLabels = []MetricLabel{
		{
			Name: "volume",
			Help: "Volume Name",
		},
		{
			Name: "brick_path",
			Help: "Brick Path",
		},
		{
			Name: "host",
			Help: "Hostname or IP",
		},
	}

	glusterVolumeHealCount = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_heal_count",
		Help:      "self heal count for volume",
		LongHelp:  "",
		Labels:    volumeHealLabels,
	})
)

func getVolumeHealLabels(volname string, host string, brick string) prometheus.Labels {
	return prometheus.Labels{
		"volume":     volname,
		"brick_path": brick,
		"host":       host,
	}

}

func healCounts() error {
	isLeader, err := gluster.IsLeader()

	if err != nil {
		// Unable to find out if the current node is leader
		// Cannot register volume metrics at this node
		log.WithError(err).Error("Unable to find if the current node is leader")
		return err
	}
	if !isLeader {
		return nil
	}
	volumes, err := gluster.VolumeInfo()
	if err != nil {
		return err
	}

	for _, volume := range volumes {
		name := volume.Name
		heals, err := gluster.HealInfo(name)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"volume": name,
			}).Error("Error getting heal info")

			continue
		}
		for _, healinfo := range heals {
			labels := getVolumeHealLabels(name, healinfo.Hostname, healinfo.Brick)
			glusterVolumeHealCount.With(labels).Set(float64(healinfo.NumHealEntries))
		}
	}
	return nil
}

func init() {
	registerMetric("gluster_volume_heal", healCounts)
}
