package main

import (
	"github.com/gluster/gluster-prometheus/pkg/glusterutils"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"strings"
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

	volumeProfileInfoLabels = []MetricLabel{
		{
			Name: "volume",
			Help: "Volume name",
		},
		{
			Name: "brick",
			Help: "Brick Name",
		},
	}

	glusterVolumeProfileTotalReads = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_profile_total_reads",
		Help:      "Total no of reads",
		LongHelp:  "",
		Labels:    volumeProfileInfoLabels,
	})

	glusterVolumeProfileTotalWrites = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_profile_total_writes",
		Help:      "Total no of writes",
		LongHelp:  "",
		Labels:    volumeProfileInfoLabels,
	})

	glusterVolumeProfileDuration = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_profile_duration_secs",
		Help:      "Duration",
		LongHelp:  "",
		Labels:    volumeProfileInfoLabels,
	})

	volumeProfileFopInfoLabels = []MetricLabel{
		{
			Name: "volume",
			Help: "Volume name",
		},
		{
			Name: "brick",
			Help: "Brick Name",
		},
		{
			Name: "host",
			Help: "Hostname or IP",
		},
		{
			Name: "fop",
			Help: "File Operation name",
		},
	}

	glusterVolumeProfileFopHits = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_profile_fop_hits",
		Help:      "Duration",
		LongHelp:  "",
		Labels:    volumeProfileFopInfoLabels,
	})

	glusterVolumeProfileFopAvgLatency = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_profile_fop_avg_latency",
		Help:      "Duration",
		LongHelp:  "",
		Labels:    volumeProfileFopInfoLabels,
	})

	glusterVolumeProfileFopMinLatency = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_profile_fop_min_latency",
		Help:      "Duration",
		LongHelp:  "",
		Labels:    volumeProfileFopInfoLabels,
	})

	glusterVolumeProfileFopMaxLatency = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_profile_fop_max_latency",
		Help:      "Duration",
		LongHelp:  "",
		Labels:    volumeProfileFopInfoLabels,
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

func getVolumeProfileInfoLabels(volname string, brick string) prometheus.Labels {
	return prometheus.Labels{
		"volume": volname,
		"brick":  brick,
	}
}

func getVolumeProfileFopInfoLabels(volname string, brick string, host string, fop string) prometheus.Labels {
	return prometheus.Labels{
		"volume": volname,
		"brick":  brick,
		"host":   host,
		"fop":    fop,
	}
}

func getBrickHost(vol glusterutils.Volume, brickname string) string {
	hostid := strings.Split(brickname, ":")[0]
	for _, subvol := range vol.SubVolumes {
		for _, brick := range subvol.Bricks {
			if brick.PeerID == hostid {
				return brick.Host
			}
		}
	}
	return ""
}

func profileInfo() error {
	isLeader, err := gluster.IsLeader()
	if err != nil {
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
		profileinfo, err := gluster.VolumeProfileInfo(name)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"volume": name,
			}).Error("Error getting profile info")
			continue
		}
		for _, entry := range profileinfo {
			labels := getVolumeProfileInfoLabels(name, entry.BrickName)
			glusterVolumeProfileTotalReads.With(labels).Set(float64(entry.TotalReads))
			glusterVolumeProfileTotalWrites.With(labels).Set(float64(entry.TotalWrites))
			glusterVolumeProfileDuration.With(labels).Set(float64(entry.Duration))
			brickhost := getBrickHost(volume, entry.BrickName)
			for _, fopInfo := range entry.FopStats {
				fopLbls := getVolumeProfileFopInfoLabels(name, entry.BrickName, brickhost, fopInfo.Name)
				glusterVolumeProfileFopHits.With(fopLbls).Set(float64(fopInfo.Hits))
				glusterVolumeProfileFopAvgLatency.With(fopLbls).Set(fopInfo.AvgLatency)
				glusterVolumeProfileFopMinLatency.With(fopLbls).Set(fopInfo.MinLatency)
				glusterVolumeProfileFopMaxLatency.With(fopLbls).Set(fopInfo.MaxLatency)
			}
		}
	}

	return nil
}

func init() {
	registerMetric("gluster_volume_heal", healCounts)
	registerMetric("gluster_volume_profile", profileInfo)
}
