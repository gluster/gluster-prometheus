package main

import (
	"errors"
	"strings"

	"github.com/gluster/gluster-prometheus/gluster-exporter/conf"
	"github.com/gluster/gluster-prometheus/pkg/glusterutils"
	"github.com/gluster/gluster-prometheus/pkg/glusterutils/glusterconsts"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

var (
	volumeHealLabels = []MetricLabel{
		clusterIDLabel,
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

	volumeHealGaugeVecs = make(map[string]*ExportedGaugeVec)

	glusterVolumeHealCount = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_heal_count",
		Help:      "self heal count for volume",
		LongHelp:  "",
		Labels:    volumeHealLabels,
	}, &volumeHealGaugeVecs)

	glusterVolumeSplitBrainHealCount = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_split_brain_heal_count",
		Help:      "self heal count for volume in split brain",
		LongHelp:  "",
		Labels:    volumeHealLabels,
	}, &volumeHealGaugeVecs)

	volumeProfileInfoLabels = []MetricLabel{
		clusterIDLabel,
		{
			Name: "volume",
			Help: "Volume name",
		},
		{
			Name: "brick",
			Help: "Brick Name",
		},
	}

	volumeProfileGaugeVecs = make(map[string]*ExportedGaugeVec)

	glusterVolumeProfileTotalReads = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_profile_total_reads",
		Help:      "Total no of reads",
		LongHelp:  "",
		Labels:    volumeProfileInfoLabels,
	}, &volumeProfileGaugeVecs)

	glusterVolumeProfileTotalWrites = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_profile_total_writes",
		Help:      "Total no of writes",
		LongHelp:  "",
		Labels:    volumeProfileInfoLabels,
	}, &volumeProfileGaugeVecs)

	glusterVolumeProfileDuration = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_profile_duration_secs",
		Help:      "Duration",
		LongHelp:  "",
		Labels:    volumeProfileInfoLabels,
	}, &volumeProfileGaugeVecs)

	glusterVolumeProfileTotalReadsInt = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_profile_total_reads_interval",
		Help:      "Total no of reads for interval stats",
		LongHelp:  "",
		Labels:    volumeProfileInfoLabels,
	}, &volumeProfileGaugeVecs)

	glusterVolumeProfileTotalWritesInt = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_profile_total_writes_interval",
		Help:      "Total no of writes for interval stats",
		LongHelp:  "",
		Labels:    volumeProfileInfoLabels,
	}, &volumeProfileGaugeVecs)

	glusterVolumeProfileDurationInt = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_profile_duration_secs_interval",
		Help:      "Duration for interval stats",
		LongHelp:  "",
		Labels:    volumeProfileInfoLabels,
	}, &volumeProfileGaugeVecs)

	volumeProfileFopInfoLabels = []MetricLabel{
		clusterIDLabel,
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

	glusterVolumeProfileFopHits = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_profile_fop_hits",
		Help:      "Cumulative FOP hits",
		LongHelp:  "",
		Labels:    volumeProfileFopInfoLabels,
	}, &volumeProfileGaugeVecs)

	glusterVolumeProfileFopAvgLatency = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_profile_fop_avg_latency",
		Help:      "Cumulative FOP avergae latency",
		LongHelp:  "",
		Labels:    volumeProfileFopInfoLabels,
	}, &volumeProfileGaugeVecs)

	glusterVolumeProfileFopMinLatency = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_profile_fop_min_latency",
		Help:      "Cumulative FOP min latency",
		LongHelp:  "",
		Labels:    volumeProfileFopInfoLabels,
	}, &volumeProfileGaugeVecs)

	glusterVolumeProfileFopMaxLatency = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_profile_fop_max_latency",
		Help:      "Cumulative FOP max latency",
		LongHelp:  "",
		Labels:    volumeProfileFopInfoLabels,
	}, &volumeProfileGaugeVecs)

	glusterVolumeProfileFopHitsInt = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_profile_fop_hits_interval",
		Help:      "Interval based FOP hits",
		LongHelp:  "",
		Labels:    volumeProfileFopInfoLabels,
	}, &volumeProfileGaugeVecs)

	glusterVolumeProfileFopAvgLatencyInt = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_profile_fop_avg_latency_interval",
		Help:      "Interval based FOP average latency",
		LongHelp:  "",
		Labels:    volumeProfileFopInfoLabels,
	}, &volumeProfileGaugeVecs)

	glusterVolumeProfileFopMinLatencyInt = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_profile_fop_min_latency_interval",
		Help:      "Interval based FOP min latency",
		LongHelp:  "",
		Labels:    volumeProfileFopInfoLabels,
	}, &volumeProfileGaugeVecs)

	glusterVolumeProfileFopMaxLatencyInt = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_profile_fop_max_latency_interval",
		Help:      "Interval based FOP max latency",
		LongHelp:  "",
		Labels:    volumeProfileFopInfoLabels,
	}, &volumeProfileGaugeVecs)

	glusterVolumeProfileFopTotalHitsAggregatedOps = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_profile_fop_total_hits_on_aggregated_fops",
		Help: "Cumulative total hits on aggregated FOPs" +
			" like READ_WRIET_OPS, LOCK_OPS, INODE_OPS etc",
		LongHelp: "",
		Labels:   volumeProfileFopInfoLabels,
	}, &volumeProfileGaugeVecs)

	glusterVolumeProfileFopTotalHitsAggregatedOpsInt = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "volume_profile_fop_total_hits_on_aggregated_fops_interval",
		Help: "Interval based total hits on aggregated FOPs" +
			" like READ_WRIET_OPS, LOCK_OPS, INODE_OPS etc",
		LongHelp: "",
		Labels:   volumeProfileFopInfoLabels,
	}, &volumeProfileGaugeVecs)
)

// opType represents aggregated operations like
// READ_WRITE_OPS, INODE_OPS, ENTRY_OPS, LOCK_OPS etc...
type opType struct {
	opName       string
	opsSupported map[string]struct{}
}

// String method returns the name of the common 'opType'
// makes it compatible with 'Stringer' interface
func (ot opType) String() string {
	return ot.opName
}

// opSupported method checks whether the given operation is supported in this opType
func (ot opType) opSupported(opLabel string) bool {
	if _, ok := ot.opsSupported[opLabel]; ok {
		return true
	}
	return false
}

// opHits calculates total number of 'ot' type operations in a list of 'FopStat's
// and returns the total number of hits
func (ot opType) opHits(fopStats []glusterutils.FopStat) float64 {
	var totalOpHits float64 // default ZERO value is assigned
	for _, eachFopS := range fopStats {
		if ot.opSupported(eachFopS.Name) {
			totalOpHits += float64(eachFopS.Hits)
		}
	}
	return totalOpHits
}

// newOperationType creates a 'opType' object
func newOperationType(commonOpName string, supportedOpLabels []string) (opType, error) {
	commonOpName = strings.TrimSpace(commonOpName)
	var opT = opType{opName: commonOpName}
	if len(supportedOpLabels) == 0 {
		return opT, errors.New("supported operation labels should not be empty")
	} else if commonOpName == "" {
		return opT, errors.New("empty common operation name is not allowed")
	}
	opT.opsSupported = make(map[string]struct{})
	var emtS struct{}
	for _, opLabel := range supportedOpLabels {
		opT.opsSupported[opLabel] = emtS
	}
	return opT, nil
}

// newReadWriteOpType creates a 'READ_WRITE_OPS' type,
// which aggregates all the read/write operations
func newReadWriteOpType() opType {
	var opsSupported = []string{"CREATE", "DISCARD", "FALLOCATE", "FLUSH", "FSYNC",
		"FSYNCDIR", "RCHECKSUM", "READ", "READDIR", "READDIRP", "READY",
		"WRITE", "ZEROFILL",
	}
	var rwOT, _ = newOperationType("READ_WRITE_OPS", opsSupported)
	return rwOT
}

// newLockOpType creates a 'LOCK_OPS' type,
// which aggregates all the lock operations
func newLockOpType() opType {
	var opsSupported = []string{"ENTRYLK", "FENTRYLK", "FINODELK", "INODELK", "LK"}
	var lockOT, _ = newOperationType("LOCK_OPS", opsSupported)
	return lockOT
}

// newINodeOpType creates a 'INODE_OPS' type,
// which aggregates all the iNode associated operations
func newINodeOpType() opType {
	var opsSupported = []string{"ACCESS", "FGETXATTR", "FREMOVEXATTR", "FSETATTR",
		"FSETXATTR", "FSTAT", "FTRUNCATE", "FXATTROP", "GETXATTR", "LOOKUP", "OPEN",
		"OPENDIR", "READLINK", "REMOVEXATTR", "SEEK", "SETATTR", "SETXATTR", "STAT",
		"STATFS", "TRUNCATE", "XATTROP"}
	var iNodeOT, _ = newOperationType("INODE_OPS", opsSupported)
	return iNodeOT
}

// newEntryOpType creates a 'ENTRY_OPS' type,
// which aggregates all the file entry related operations
func newEntryOpType() opType {
	var opsSupported = []string{"LINK", "MKDIR", "MKNOD", "RENAME",
		"RMDIR", "SYMLINK", "UNLINK"}
	var entryOT, _ = newOperationType("ENTRY_OPS", opsSupported)
	return entryOT
}

func getVolumeHealLabels(volname string, host string, brick string) prometheus.Labels {
	return prometheus.Labels{
		"cluster_id": clusterID,
		"volume":     volname,
		"brick_path": brick,
		"host":       host,
	}

}

func healCounts(gluster glusterutils.GInterface) error {
	isLeader, err := gluster.IsLeader()

	// Reset all vecs to not export stale information
	for _, gaugeVec := range volumeHealGaugeVecs {
		gaugeVec.RemoveStaleMetrics()
	}

	if err != nil {
		// Unable to find out if the current node is leader
		// Cannot register volume metrics at this node
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

	// locHealInfoFunc is a function literal, which takes
	// arg1: f1 a function which takes a string and returns ([]HealEntry, error)
	// (can be 'HealInfo' or 'SplitBrainHealInfo')
	// arg2: gVect a pointer to GaugeVec
	// (can be either 'glusterVolumeHealCount' or 'glusterVolumeSplitBrainHealCount')
	// arg3: volName a string representing the volume name
	// arg4: errStr the error string in case of error
	locHealInfoFunc := func(f1 func(string) ([]glusterutils.HealEntry, error), gVect string, volName string, errStr string) {
		// Get the heal count
		heals, err := f1(volName)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"volume": volName,
			}).Debug(errStr)
			return
		}
		for _, healinfo := range heals {
			labels := getVolumeHealLabels(volName, healinfo.Hostname, healinfo.Brick)
			volumeHealGaugeVecs[gVect].Set(labels, float64(healinfo.NumHealEntries))
		}
	}

	for _, volume := range volumes {
		name := volume.Name
		if strings.Contains(volume.Type, "Replicate") {
			locHealInfoFunc(gluster.HealInfo, glusterVolumeHealCount, name, "Error getting heal info")
			locHealInfoFunc(gluster.SplitBrainHealInfo, glusterVolumeSplitBrainHealCount, name, "Error getting split brain heal info")
		}
		if strings.Contains(volume.Type, "Disperse") {
			locHealInfoFunc(gluster.HealInfo, glusterVolumeHealCount, name, "Error getting heal info")
		}
	}
	return nil
}

func getVolumeProfileInfoLabels(volname string, brick string) prometheus.Labels {
	return prometheus.Labels{
		"cluster_id": clusterID,
		"volume":     volname,
		"brick":      brick,
	}
}

func getVolumeProfileFopInfoLabels(volname string, brick string, host string, fop string) prometheus.Labels {
	return prometheus.Labels{
		"cluster_id": clusterID,
		"volume":     volname,
		"brick":      brick,
		"host":       host,
		"fop":        fop,
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

func profileInfo(gluster glusterutils.GInterface) error {
	// Reset all vecs to not export stale information
	for _, gaugeVec := range volumeProfileGaugeVecs {
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
	volOption := glusterconsts.CountFOPHitsGD1
	var glusterConfig *conf.GConfig
	if glusterConfig, err = conf.GConfigFromInterface(gluster); err != nil {
		return err
	}
	if glusterConfig.GlusterMgmt == glusterconsts.MgmtGlusterd2 {
		volOption = glusterconsts.CountFOPHitsGD2
	}
	var (
		// supported aggregated operations are,
		// READ_WRITE_OPS, LOCK_OPS, ENTRY_OPS, INODE_OPS
		aggregatedOps = []opType{newReadWriteOpType(), newLockOpType(),
			newEntryOpType(), newINodeOpType()}
	)
	for _, volume := range volumes {
		err := gluster.EnableVolumeProfiling(volume)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"volume": volume.Name,
			}).Debug("Error enabling profiling for volume")
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
			}).Debug("Error getting profile info")
			continue
		}
		for _, entry := range profileinfo {
			labels := getVolumeProfileInfoLabels(name, entry.BrickName)
			volumeProfileGaugeVecs[glusterVolumeProfileTotalReads].Set(labels, float64(entry.TotalReads))
			volumeProfileGaugeVecs[glusterVolumeProfileTotalWrites].Set(labels, float64(entry.TotalWrites))
			volumeProfileGaugeVecs[glusterVolumeProfileDuration].Set(labels, float64(entry.Duration))
			volumeProfileGaugeVecs[glusterVolumeProfileTotalReadsInt].Set(labels, float64(entry.TotalReadsInt))
			volumeProfileGaugeVecs[glusterVolumeProfileTotalWritesInt].Set(labels, float64(entry.TotalWritesInt))
			volumeProfileGaugeVecs[glusterVolumeProfileDurationInt].Set(labels, float64(entry.DurationInt))
			brickhost := getBrickHost(volume, entry.BrickName)
			for _, eachOp := range aggregatedOps {
				fopLbls := getVolumeProfileFopInfoLabels(name, entry.BrickName,
					brickhost, eachOp.String())
				volumeProfileGaugeVecs[glusterVolumeProfileFopTotalHitsAggregatedOps].Set(fopLbls, eachOp.opHits(entry.FopStats))
				volumeProfileGaugeVecs[glusterVolumeProfileFopTotalHitsAggregatedOpsInt].Set(fopLbls, eachOp.opHits(entry.FopStatsInt))
			}
			for _, fopInfo := range entry.FopStats {
				fopLbls := getVolumeProfileFopInfoLabels(name, entry.BrickName, brickhost, fopInfo.Name)
				volumeProfileGaugeVecs[glusterVolumeProfileFopHits].Set(fopLbls, float64(fopInfo.Hits))
				volumeProfileGaugeVecs[glusterVolumeProfileFopAvgLatency].Set(fopLbls, fopInfo.AvgLatency)
				volumeProfileGaugeVecs[glusterVolumeProfileFopMinLatency].Set(fopLbls, fopInfo.MinLatency)
				volumeProfileGaugeVecs[glusterVolumeProfileFopMaxLatency].Set(fopLbls, fopInfo.MaxLatency)
			}
			for _, fopInfo := range entry.FopStatsInt {
				fopLbls := getVolumeProfileFopInfoLabels(name, entry.BrickName, brickhost, fopInfo.Name)
				volumeProfileGaugeVecs[glusterVolumeProfileFopHitsInt].Set(fopLbls, float64(fopInfo.Hits))
				volumeProfileGaugeVecs[glusterVolumeProfileFopAvgLatencyInt].Set(fopLbls, fopInfo.AvgLatency)
				volumeProfileGaugeVecs[glusterVolumeProfileFopMinLatencyInt].Set(fopLbls, fopInfo.MinLatency)
				volumeProfileGaugeVecs[glusterVolumeProfileFopMaxLatencyInt].Set(fopLbls, fopInfo.MaxLatency)
			}
		}
	}

	return nil
}

func init() {
	registerMetric("gluster_volume_heal", healCounts)
	registerMetric("gluster_volume_profile", profileInfo)
}
