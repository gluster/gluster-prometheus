package main

import (
	"errors"

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

	glusterVolumeProfileTotalReadsInt = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_profile_total_reads_interval",
		Help:      "Total no of reads for interval stats",
		LongHelp:  "",
		Labels:    volumeProfileInfoLabels,
	})

	glusterVolumeProfileTotalWritesInt = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_profile_total_writes_interval",
		Help:      "Total no of writes for interval stats",
		LongHelp:  "",
		Labels:    volumeProfileInfoLabels,
	})

	glusterVolumeProfileDurationInt = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_profile_duration_secs_interval",
		Help:      "Duration for interval stats",
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
		Help:      "Cumulative FOP hits",
		LongHelp:  "",
		Labels:    volumeProfileFopInfoLabels,
	})

	glusterVolumeProfileFopAvgLatency = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_profile_fop_avg_latency",
		Help:      "Cumulative FOP avergae latency",
		LongHelp:  "",
		Labels:    volumeProfileFopInfoLabels,
	})

	glusterVolumeProfileFopMinLatency = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_profile_fop_min_latency",
		Help:      "Cumulative FOP min latency",
		LongHelp:  "",
		Labels:    volumeProfileFopInfoLabels,
	})

	glusterVolumeProfileFopMaxLatency = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_profile_fop_max_latency",
		Help:      "Cumulative FOP max latency",
		LongHelp:  "",
		Labels:    volumeProfileFopInfoLabels,
	})

	glusterVolumeProfileFopHitsInt = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_profile_fop_hits_interval",
		Help:      "Interval based FOP hits",
		LongHelp:  "",
		Labels:    volumeProfileFopInfoLabels,
	})

	glusterVolumeProfileFopAvgLatencyInt = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_profile_fop_avg_latency_interval",
		Help:      "Interval based FOP average latency",
		LongHelp:  "",
		Labels:    volumeProfileFopInfoLabels,
	})

	glusterVolumeProfileFopMinLatencyInt = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_profile_fop_min_latency_interval",
		Help:      "Interval based FOP min latency",
		LongHelp:  "",
		Labels:    volumeProfileFopInfoLabels,
	})

	glusterVolumeProfileFopMaxLatencyInt = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_profile_fop_max_latency_interval",
		Help:      "Interval based FOP max latency",
		LongHelp:  "",
		Labels:    volumeProfileFopInfoLabels,
	})

	glusterVolumeProfileFopTotalHitsAggregatedOps = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_profile_fop_total_hits_on_aggregated_fops",
		Help: "Cumulative total hits on aggregated FOPs" +
			" like READ_WRIET_OPS, LOCK_OPS, INODE_OPS etc",
		LongHelp: "",
		Labels:   volumeProfileFopInfoLabels,
	})

	glusterVolumeProfileFopTotalHitsAggregatedOpsInt = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_profile_fop_total_hits_on_aggregated_fops_interval",
		Help: "Interval based total hits on aggregated FOPs" +
			" like READ_WRIET_OPS, LOCK_OPS, INODE_OPS etc",
		LongHelp: "",
		Labels:   volumeProfileFopInfoLabels,
	})
)

// OpType represents aggregated operations like
// READ_WRITE_OPS, INODE_OPS, ENTRY_OPS, LOCK_OPS etc...
type OpType struct {
	opName       string
	opsSupported map[string]struct{}
}

// String method returns the name of the common 'OpType'
// makes it compatible with 'Stringer' interface
func (ot OpType) String() string {
	return ot.opName
}

// ListSupportedOps method returns a list of supported operations
func (ot OpType) ListSupportedOps() []string {
	var supportedOps = make([]string, 0, len(ot.opsSupported))
	for opLabel := range ot.opsSupported {
		supportedOps = append(supportedOps, opLabel)
	}
	return supportedOps
}

// OpSupported method checks whether the given operation is supported in this OpType
func (ot OpType) OpSupported(opLabel string) bool {
	if _, ok := ot.opsSupported[opLabel]; ok {
		return true
	}
	return false
}

// AppendOp method allow others to extend the existing OpTypes
func (ot OpType) AppendOp(opLabel string) {
	if ot.opsSupported == nil {
		ot.opsSupported = make(map[string]struct{})
	}
	ot.opsSupported[opLabel] = struct{}{}
}

// OpHits calculates total number of 'ot' type operations in a list of 'FopStat's
// and returns the total number of hits
func (ot OpType) OpHits(fopStats []glusterutils.FopStat) float64 {
	var totalOpHits float64 // default ZERO value is assigned
	for _, eachFopS := range fopStats {
		if ot.OpSupported(eachFopS.Name) {
			totalOpHits += float64(eachFopS.Hits)
		}
	}
	return totalOpHits
}

// NewOperationType creates a 'OpType' object
func NewOperationType(commonOpName string, supportedOpLabels []string) (OpType, error) {
	commonOpName = strings.TrimSpace(commonOpName)
	var opT = OpType{opName: commonOpName}
	if len(supportedOpLabels) == 0 {
		return opT, errors.New("Supported operation labels should not be empty")
	} else if commonOpName == "" {
		return opT, errors.New("Empty common operation name is not allowed")
	}
	opT.opsSupported = make(map[string]struct{})
	var emtS struct{}
	for _, opLabel := range supportedOpLabels {
		opT.opsSupported[opLabel] = emtS
	}
	return opT, nil
}

// NewReadWriteOpType creates a 'READ_WRITE_OPS' type,
// which aggregates all the read/write operations
func NewReadWriteOpType() OpType {
	var opsSupported = []string{"CREATE", "DISCARD", "FALLOCATE", "FLUSH", "FSYNC",
		"FSYNCDIR", "RCHECKSUM", "READ", "READDIR", "READDIRP", "READY",
		"WRITE", "ZEROFILL",
	}
	var rwOT, _ = NewOperationType("READ_WRITE_OPS", opsSupported)
	return rwOT
}

// NewLockOpType creates a 'LOCK_OPS' type,
// which aggregates all the lock operations
func NewLockOpType() OpType {
	var opsSupported = []string{"ENTRYLK", "FENTRYLK", "FINODELK", "INODELK", "LK"}
	var lockOT, _ = NewOperationType("LOCK_OPS", opsSupported)
	return lockOT
}

// NewINodeOpType creates a 'INODE_OPS' type,
// which aggregates all the iNode associated operations
func NewINodeOpType() OpType {
	var opsSupported = []string{"ACCESS", "FGETXATTR", "FREMOVEXATTR", "FSETATTR",
		"FSETXATTR", "FSTAT", "FTRUNCATE", "FXATTROP", "GETXATTR", "LOOKUP", "OPEN",
		"OPENDIR", "READLINK", "REMOVEXATTR", "SEEK", "SETATTR", "SETXATTR", "STAT",
		"STATFS", "TRUNCATE", "XATTROP"}
	var iNodeOT, _ = NewOperationType("INODE_OPS", opsSupported)
	return iNodeOT
}

// NewEntryOpType creates a 'ENTRY_OPS' type,
// which aggregates all the file entry related operations
func NewEntryOpType() OpType {
	var opsSupported = []string{"LINK", "MKDIR", "MKNOD", "RENAME",
		"RMDIR", "SYMLINK", "UNLINK"}
	var entryOT, _ = NewOperationType("ENTRY_OPS", opsSupported)
	return entryOT
}

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
	volOption := glusterutils.CountFOPHitsGD1
	if glusterConfig.GlusterMgmt == glusterutils.MgmtGlusterd2 {
		volOption = glusterutils.CountFOPHitsGD2
	}
	var (
		// supported aggregated operations are,
		// READ_WRITE_OPS, LOCK_OPS, ENTRY_OPS, INODE_OPS
		aggregatedOps = []OpType{NewReadWriteOpType(), NewLockOpType(),
			NewEntryOpType(), NewINodeOpType()}
	)
	for _, volume := range volumes {
		err := gluster.EnableVolumeProfiling(volume)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"volume": volume.Name,
			}).Error("Error enabling profiling for volume")
			continue
		}
		if value, exists := volume.Options[volOption]; !exists || value == "off" {
			// Volume profiling is explicitly switched off for volume, dont collect profile metrics
			continue
		}
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
			glusterVolumeProfileTotalReadsInt.With(labels).Set(float64(entry.TotalReadsInt))
			glusterVolumeProfileTotalWritesInt.With(labels).Set(float64(entry.TotalWritesInt))
			glusterVolumeProfileDurationInt.With(labels).Set(float64(entry.DurationInt))
			brickhost := getBrickHost(volume, entry.BrickName)
			for _, eachOp := range aggregatedOps {
				fopLbls := getVolumeProfileFopInfoLabels(name, entry.BrickName,
					brickhost, eachOp.String())
				glusterVolumeProfileFopTotalHitsAggregatedOps.With(fopLbls).Set(eachOp.OpHits(entry.FopStats))
				glusterVolumeProfileFopTotalHitsAggregatedOpsInt.With(fopLbls).Set(eachOp.OpHits(entry.FopStatsInt))
			}
			for _, fopInfo := range entry.FopStats {
				fopLbls := getVolumeProfileFopInfoLabels(name, entry.BrickName, brickhost, fopInfo.Name)
				glusterVolumeProfileFopHits.With(fopLbls).Set(float64(fopInfo.Hits))
				glusterVolumeProfileFopAvgLatency.With(fopLbls).Set(fopInfo.AvgLatency)
				glusterVolumeProfileFopMinLatency.With(fopLbls).Set(fopInfo.MinLatency)
				glusterVolumeProfileFopMaxLatency.With(fopLbls).Set(fopInfo.MaxLatency)
			}
			for _, fopInfo := range entry.FopStatsInt {
				fopLbls := getVolumeProfileFopInfoLabels(name, entry.BrickName, brickhost, fopInfo.Name)
				glusterVolumeProfileFopHitsInt.With(fopLbls).Set(float64(fopInfo.Hits))
				glusterVolumeProfileFopAvgLatencyInt.With(fopLbls).Set(fopInfo.AvgLatency)
				glusterVolumeProfileFopMinLatencyInt.With(fopLbls).Set(fopInfo.MinLatency)
				glusterVolumeProfileFopMaxLatencyInt.With(fopLbls).Set(fopInfo.MaxLatency)
			}
		}
	}

	return nil
}

func init() {
	registerMetric("gluster_volume_heal", healCounts)
	registerMetric("gluster_volume_profile", profileInfo)
}
