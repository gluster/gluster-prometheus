package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/gluster/gluster-prometheus/pkg/glusterutils"
	"github.com/gluster/gluster-prometheus/pkg/glusterutils/glusterconsts"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

var (
	brickLabels = []MetricLabel{
		clusterIDLabel,
		{
			Name: "host",
			Help: "Host name or IP",
		},
		{
			Name: "id",
			Help: "Brick ID",
		},
		{
			Name: "brick_path",
			Help: "Brick Path",
		},
		{
			Name: "volume",
			Help: "Volume Name",
		},
		{
			Name: "subvolume",
			Help: "Sub Volume name",
		},
	}

	subvolLabels = []MetricLabel{
		clusterIDLabel,
		{
			Name: "volume",
			Help: "Volume Name",
		},
		{
			Name: "subvolume",
			Help: "Sub volume name",
		},
	}

	lvmLbls = []MetricLabel{
		clusterIDLabel,
		{
			Name: "host",
			Help: "Host name or IP",
		},
		{
			Name: "id",
			Help: "Brick ID",
		},
		{
			Name: "brick_path",
			Help: "Brick Path",
		},
		{
			Name: "volume",
			Help: "Volume Name",
		},
		{
			Name: "subvolume",
			Help: "Sub Volume name",
		},
		{
			Name: "vg_name",
			Help: "VG Name",
		},
		{
			Name: "lv_path",
			Help: "LV Path",
		},
		{
			Name: "lv_uuid",
			Help: "UUID of LV",
		},
	}

	brickStatusLbls = []MetricLabel{
		clusterIDLabel,
		{
			Name: "volume",
			Help: "Volume Name",
		},
		{
			Name: "hostname",
			Help: "Host name or IP",
		},
		{
			Name: "brick_path",
			Help: "Brick Path",
		},
		{
			Name: "peer_id",
			Help: "Peer ID",
		},
		{
			Name: "pid",
			Help: "Process ID of brick",
		},
	}

	thinLvmLbls = []MetricLabel{
		clusterIDLabel,
		{
			Name: "host",
			Help: "Host name or IP",
		},
		{
			Name: "thinpool_name",
			Help: "Name of the thinpool LV",
		},
		{
			Name: "vg_name",
			Help: "Name of the Volume Group",
		},
		{
			Name: "volume",
			Help: "Volume Name",
		},
		{
			Name: "subvolume",
			Help: "Name of the Subvolume",
		},
		{
			Name: "brick_path",
			Help: "Brick Path",
		},
	}

	brickGaugeVecs = make(map[string]*ExportedGaugeVec)

	glusterBrickCapacityUsed = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "brick_capacity_used_bytes",
		Help:      "Used capacity of gluster bricks in bytes",
		LongHelp:  "",
		Labels:    brickLabels,
	}, &brickGaugeVecs)

	glusterBrickCapacityFree = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "brick_capacity_free_bytes",
		Help:      "Free capacity of gluster bricks in bytes",
		LongHelp:  "",
		Labels:    brickLabels,
	}, &brickGaugeVecs)

	glusterBrickCapacityTotal = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "brick_capacity_bytes_total",
		Help:      "Total capacity of gluster bricks in bytes",
		LongHelp:  "",
		Labels:    brickLabels,
	}, &brickGaugeVecs)

	glusterBrickInodesTotal = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "brick_inodes_total",
		Help:      "Total no of inodes of gluster brick disk",
		LongHelp:  "",
		Labels:    brickLabels,
	}, &brickGaugeVecs)

	glusterBrickInodesFree = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "brick_inodes_free",
		Help:      "Free no of inodes of gluster brick disk",
		LongHelp:  "",
		Labels:    brickLabels,
	}, &brickGaugeVecs)

	glusterBrickInodesUsed = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "brick_inodes_used",
		Help:      "Used no of inodes of gluster brick disk",
		LongHelp:  "",
		Labels:    brickLabels,
	}, &brickGaugeVecs)

	glusterSubvolCapacityUsed = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "subvol_capacity_used_bytes",
		Help:      "Effective used capacity of gluster subvolume in bytes",
		LongHelp:  "",
		Labels:    subvolLabels,
	}, &brickGaugeVecs)

	glusterSubvolCapacityTotal = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "subvol_capacity_total_bytes",
		Help:      "Effective total capacity of gluster subvolume in bytes",
		LongHelp:  "",
		Labels:    subvolLabels,
	}, &brickGaugeVecs)

	glusterBrickLVSize = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "brick_lv_size_bytes",
		Help:      "Bricks LV size Bytes",
		LongHelp:  "",
		Labels:    lvmLbls,
	}, &brickGaugeVecs)

	glusterBrickLVPercent = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "brick_lv_percent",
		Help:      "Bricks LV usage percent",
		LongHelp:  "",
		Labels:    lvmLbls,
	}, &brickGaugeVecs)

	glusterBrickLVMetadataSize = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "brick_lv_metadata_size_bytes",
		Help:      "Bricks LV metadata size Bytes",
		LongHelp:  "",
		Labels:    lvmLbls,
	}, &brickGaugeVecs)

	glusterBrickLVMetadataPercent = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "brick_lv_metadata_percent",
		Help:      "Bricks LV metadata usage percent",
		LongHelp:  "",
		Labels:    lvmLbls,
	}, &brickGaugeVecs)

	glusterVGExtentTotal = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "vg_extent_total_count",
		Help:      "VG extent total count ",
		LongHelp:  "",
		Labels:    lvmLbls,
	}, &brickGaugeVecs)

	glusterVGExtentAlloc = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "vg_extent_alloc_count",
		Help:      "VG extent allocated count ",
		LongHelp:  "",
		Labels:    lvmLbls,
	}, &brickGaugeVecs)

	glusterThinPoolDataTotal = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "thinpool_data_total_bytes",
		Help:      "Thin pool size Bytes",
		LongHelp:  "",
		Labels:    thinLvmLbls,
	}, &brickGaugeVecs)

	glusterThinPoolDataUsed = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "thinpool_data_used_bytes",
		Help:      "Thin pool data used Bytes",
		LongHelp:  "",
		Labels:    thinLvmLbls,
	}, &brickGaugeVecs)

	glusterThinPoolMetadataTotal = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "thinpool_metadata_total_bytes",
		Help:      "Thin pool metadata size Bytes",
		LongHelp:  "",
		Labels:    thinLvmLbls,
	}, &brickGaugeVecs)

	glusterThinPoolMetadataUsed = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "thinpool_metadata_used_bytes",
		Help:      "Thin pool metadata used Bytes",
		LongHelp:  "",
		Labels:    thinLvmLbls,
	}, &brickGaugeVecs)

	brickStatusGaugeVecs = make(map[string]*ExportedGaugeVec)

	glusterBrickUp = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "brick_up",
		Help:      "Brick up (1-up, 0-down)",
		LongHelp:  "",
		Labels:    brickStatusLbls,
	}, &brickStatusGaugeVecs)
)

func getGlusterBrickLabels(brick glusterutils.Brick, subvol string) prometheus.Labels {
	return prometheus.Labels{
		"cluster_id": clusterID,
		"host":       brick.Host,
		"id":         brick.ID,
		"brick_path": brick.Path,
		"volume":     brick.VolumeName,
		"subvolume":  subvol,
	}
}

func getGlusterSubvolLabels(volname string, subvol string) prometheus.Labels {
	return prometheus.Labels{
		"cluster_id": clusterID,
		"volume":     volname,
		"subvolume":  subvol,
	}
}

// DiskStatus represents Disk usage
type DiskStatus struct {
	All        float64 `json:"all"`
	Used       float64 `json:"used"`
	Free       float64 `json:"free"`
	InodesAll  float64 `json:"inodesall"`
	InodesFree float64 `json:"inodesfree"`
	InodesUsed float64 `json:"inodesused"`
}

func diskUsage(path string) (disk DiskStatus, err error) {
	fs := syscall.Statfs_t{}
	err = syscall.Statfs(path, &fs)
	if err != nil {
		return
	}
	disk.All = float64(fs.Blocks * uint64(fs.Bsize))
	disk.Free = float64(fs.Bfree * uint64(fs.Bsize))
	disk.Used = disk.All - disk.Free
	disk.InodesAll = float64(fs.Files)
	disk.InodesFree = float64(fs.Ffree)
	disk.InodesUsed = disk.InodesAll - disk.InodesFree
	return
}

// LVMStat represents LVM details
type LVMStat struct {
	Device          string
	UUID            string
	Name            string
	DataPercent     float64
	PoolLV          string
	Attr            string
	Size            float64
	Path            string
	MetadataSize    float64
	MetadataPercent float64
	VGName          string
	VGExtentTotal   float64
	VGExtentAlloc   float64
}

// ThinPoolStat represents thin pool LV details
type ThinPoolStat struct {
	ThinPoolName          string
	ThinPoolVGName        string
	ThinPoolDataTotal     float64
	ThinPoolDataUsed      float64
	ThinPoolMetadataTotal float64
	ThinPoolMetadataUsed  float64
}

// VGReport represents VG details
type VGReport struct {
	Report []VGs `json:"report"`
}

// VGs represents list VG Details
type VGs struct {
	Vgs []VGDetails `json:"vg"`
}

// VGDetails represents a single VG detail
type VGDetails struct {
	LVUUID          string `json:"lv_uuid"`
	LVName          string `json:"lv_name"`
	DataPercent     string `json:"data_percent"`
	PoolLV          string `json:"pool_lv"`
	LVAttr          string `json:"lv_attr"`
	LVSize          string `json:"lv_size"`
	LVPath          string `json:"lv_path"`
	LVMetadataSize  string `json:"lv_metadata_size"`
	MetadataPercent string `json:"metadata_percent"`
	VGName          string `json:"vg_name"`
	VGExtentTotal   string `json:"vg_extent_count"`
	VGExtentFree    string `json:"vg_free_count"`
}

func getLVS() ([]LVMStat, []ThinPoolStat, error) {
	cmd := "lvm vgs --unquoted --reportformat=json --noheading --nosuffix --units m -o lv_uuid,lv_name,data_percent,pool_lv,lv_attr,lv_size,lv_path,lv_metadata_size,metadata_percent,vg_name,vg_extent_count,vg_free_count"

	out, err := exec.Command("sh", "-c", cmd).Output()
	lvmDet := []LVMStat{}
	thinPool := []ThinPoolStat{}
	var vgExtentFreeTemp float64
	if err != nil {
		log.WithError(err).Debug("Error getting lvm usage details")
		return lvmDet, thinPool, err
	}
	var vgReport VGReport
	if err1 := json.Unmarshal(out, &vgReport); err1 != nil {
		log.WithError(err1).Debug("Error parsing lvm usage details")
		return lvmDet, thinPool, err1
	}

	for _, vg := range vgReport.Report[0].Vgs {
		var obj LVMStat
		obj.UUID = vg.LVUUID
		obj.Name = vg.LVName
		if vg.DataPercent == "" {
			obj.DataPercent = 0.0
		} else {
			if obj.DataPercent, err = strconv.ParseFloat(vg.DataPercent, 64); err != nil {
				log.WithError(err).Debug("Error parsing DataPercent value of lvm usage")
				return lvmDet, thinPool, err
			}
		}
		obj.PoolLV = vg.PoolLV
		obj.Attr = vg.LVAttr
		if vg.LVSize == "" {
			obj.Size = 0.0
		} else {
			if obj.Size, err = strconv.ParseFloat(vg.LVSize, 64); err != nil {
				log.WithError(err).Debug("Error parsing LVSize value of lvm usage")
				return lvmDet, thinPool, err
			}
		}
		obj.Path = vg.LVPath
		if vg.LVMetadataSize == "" {
			obj.MetadataSize = 0.0
		} else {
			if obj.MetadataSize, err = strconv.ParseFloat(vg.LVMetadataSize, 64); err != nil {
				log.WithError(err).Debug("Error parsing LVMetadataSize value of lvm usage")
				return lvmDet, thinPool, err
			}
		}
		if vg.MetadataPercent == "" {
			obj.MetadataPercent = 0.0
		} else {
			obj.MetadataPercent, err = strconv.ParseFloat(vg.MetadataPercent, 64)
			if err != nil {
				log.WithError(err).Debug("Error parsing MetadataPercent value of lvm usage")
				return lvmDet, thinPool, err
			}
		}
		if vg.VGExtentTotal == "" {
			obj.VGExtentTotal = 0.0
		} else {
			obj.VGExtentTotal, err = strconv.ParseFloat(vg.VGExtentTotal, 64)
			if err != nil {
				log.WithError(err).Debug("Error parsing VGExtenTotal value of lvm usage")
				return lvmDet, thinPool, err
			}
		}
		if vg.VGExtentFree == "" {
			vgExtentFreeTemp = 0.0
		} else {
			vgExtentFreeTemp, err = strconv.ParseFloat(vg.VGExtentFree, 64)
			if err != nil {
				log.WithError(err).Debug("Error parsing VGExtentAlloc value of lvm usage")
				return lvmDet, thinPool, err
			}
		}
		obj.VGExtentAlloc = obj.VGExtentTotal - vgExtentFreeTemp
		obj.VGName = vg.VGName
		if obj.Attr[0] == 't' {
			obj.Device = fmt.Sprintf("%s/%s", obj.VGName, obj.Name)
			var TPUsage ThinPoolStat
			TPUsage.ThinPoolName = obj.Name
			TPUsage.ThinPoolVGName = obj.VGName
			TPUsage.ThinPoolDataTotal = obj.Size
			TPUsage.ThinPoolDataUsed = (obj.Size * obj.DataPercent) / 100
			TPUsage.ThinPoolMetadataTotal = obj.MetadataSize
			TPUsage.ThinPoolMetadataUsed = (obj.MetadataSize * obj.MetadataPercent) / 100
			thinPool = append(thinPool, TPUsage)
		} else {
			obj.Device, err = filepath.EvalSymlinks(obj.Path)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"path": obj.Path,
				}).Debug("Error evaluating realpath")
				return lvmDet, thinPool, err
			}
		}
		lvmDet = append(lvmDet, obj)
	}
	return lvmDet, thinPool, nil
}

// ProcMounts represents list of items from /proc/mounts
type ProcMounts struct {
	Name         string
	Device       string
	FSType       string
	MountOptions string
}

func parseProcMounts() ([]ProcMounts, error) {
	procMounts := []ProcMounts{}
	b, err := ioutil.ReadFile("/proc/mounts")
	if err != nil {
		return procMounts, err
	}
	for _, line := range strings.Split(string(b), "\n") {
		if strings.HasPrefix(line, "/") {
			tokens := strings.Fields(line)
			procMounts = append(procMounts,
				ProcMounts{Name: tokens[1], Device: tokens[0], FSType: tokens[2], MountOptions: tokens[3]})
		}
	}
	return procMounts, nil
}

func getGlusterLVMLabels(brick glusterutils.Brick, subvol string, stat LVMStat) prometheus.Labels {
	return prometheus.Labels{
		"cluster_id": clusterID,
		"host":       brick.Host,
		"id":         brick.ID,
		"brick_path": brick.Path,
		"volume":     brick.VolumeName,
		"subvolume":  subvol,
		"vg_name":    stat.VGName,
		"lv_path":    stat.Path,
		"lv_uuid":    stat.UUID,
	}
}

func getGlusterThinPoolLabels(brick glusterutils.Brick, vol string, subvol string, thinStat ThinPoolStat) prometheus.Labels {
	return prometheus.Labels{
		"cluster_id":    clusterID,
		"host":          brick.Host,
		"thinpool_name": thinStat.ThinPoolName,
		"vg_name":       thinStat.ThinPoolVGName,
		"volume":        vol,
		"subvolume":     subvol,
		"brick_path":    brick.Path,
	}
}

func lvmUsage(path string) (stats []LVMStat, thinPoolStats []ThinPoolStat, err error) {
	mountPoints, err := parseProcMounts()
	if err != nil {
		return stats, thinPoolStats, err
	}
	var thinPoolNames []string
	lvs, tpStats, err := getLVS()
	if err != nil {
		return stats, thinPoolStats, err
	}
	for _, lv := range lvs {
		for _, mount := range mountPoints {
			dev, err := filepath.EvalSymlinks(mount.Device)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"path": mount.Device,
				}).Debug("Error evaluating realpath")
				continue
			}
			// Check if the logical volume is mounted as a gluster brick
			if lv.Device == dev && strings.HasPrefix(path, mount.Name) && strings.Contains(path, "/"+lv.VGName+"/"+lv.Name) {
				// Check if the LV is a thinly provisioned volume and if yes then get the thin pool LV name
				if lv.Attr[0] == 'V' {
					tpName := lv.PoolLV
					thinPoolNames = append(thinPoolNames, tpName)
				}
				stats = append(stats, lv)
			}
		}
	}
	// Iterate and select only those thin pool LVs whose thinly provisioned volumes are mounted as gluster bricks
	for _, tpName := range thinPoolNames {
		for _, tpStat := range tpStats {
			if tpName == tpStat.ThinPoolName {
				thinPoolStats = append(thinPoolStats, tpStat)
			}
		}
	}

	return stats, thinPoolStats, nil
}

func brickUtilization(gluster glusterutils.GInterface) error {
	// Reset all vecs to not export stale information
	for _, gaugeVec := range brickGaugeVecs {
		gaugeVec.RemoveStaleMetrics()
	}

	volumes, err := gluster.VolumeInfo()

	if err != nil {
		// Return without exporting metric in this cycle
		return err
	}

	localPeerID, err := gluster.LocalPeerID()
	if err != nil {
		// Return without exporting metric in this cycle
		return err
	}

	for _, volume := range volumes {
		if volume.State != glusterconsts.VolumeStateStarted {
			// Export brick metrics only if the Volume
			// is is in Started state
			continue
		}
		subvols := volume.SubVolumes
		for _, subvol := range subvols {
			bricks := subvol.Bricks
			var maxBrickUsed float64
			var leastBrickTotal float64
			for _, brick := range bricks {
				if brick.PeerID == localPeerID {
					usage, err := diskUsage(brick.Path)
					if err != nil {
						log.WithError(err).WithFields(log.Fields{
							"volume":     volume.Name,
							"brick_path": brick.Path,
						}).Debug("Error getting disk usage")
						continue
					}
					var lbls = getGlusterBrickLabels(brick, subvol.Name)
					// Update the metrics
					brickGaugeVecs[glusterBrickCapacityUsed].Set(lbls, usage.Used)
					brickGaugeVecs[glusterBrickCapacityFree].Set(lbls, usage.Free)
					brickGaugeVecs[glusterBrickCapacityTotal].Set(lbls, usage.All)
					brickGaugeVecs[glusterBrickInodesTotal].Set(lbls, usage.InodesAll)
					brickGaugeVecs[glusterBrickInodesFree].Set(lbls, usage.InodesFree)
					brickGaugeVecs[glusterBrickInodesUsed].Set(lbls, usage.InodesUsed)
					// Skip exporting utilization data in case of arbiter
					// brick to avoid wrong values when both the data bricks
					// are down
					if brick.Type != glusterconsts.BrickTypeArbiter && usage.Used >= maxBrickUsed {
						maxBrickUsed = usage.Used
					}
					if brick.Type != glusterconsts.BrickTypeArbiter {
						if leastBrickTotal == 0 || usage.All <= leastBrickTotal {
							leastBrickTotal = usage.All
						}
					}
					// Get lvm usage details
					stats, thinStats, err := lvmUsage(brick.Path)
					if err != nil {
						log.WithError(err).WithFields(log.Fields{
							"volume":     volume.Name,
							"brick_path": brick.Path,
						}).Debug("Error getting lvm usage")
						continue
					}
					// Add metrics
					for _, stat := range stats {
						var lvmLbls = getGlusterLVMLabels(brick, subvol.Name, stat)
						// Convert to bytes
						brickGaugeVecs[glusterBrickLVSize].Set(lvmLbls, stat.Size*1024*1024)
						brickGaugeVecs[glusterBrickLVPercent].Set(lvmLbls, stat.DataPercent)
						// Convert to bytes
						brickGaugeVecs[glusterBrickLVMetadataSize].Set(lvmLbls, stat.MetadataSize*1024*1024)
						brickGaugeVecs[glusterBrickLVMetadataPercent].Set(lvmLbls, stat.MetadataPercent)
						brickGaugeVecs[glusterVGExtentTotal].Set(lvmLbls, stat.VGExtentTotal)
						brickGaugeVecs[glusterVGExtentAlloc].Set(lvmLbls, stat.VGExtentAlloc)
					}
					for _, thinStat := range thinStats {
						var thinLvmLbls = getGlusterThinPoolLabels(brick, volume.Name, subvol.Name, thinStat)
						brickGaugeVecs[glusterThinPoolDataTotal].Set(thinLvmLbls, thinStat.ThinPoolDataTotal*1024*1024)
						brickGaugeVecs[glusterThinPoolDataUsed].Set(thinLvmLbls, thinStat.ThinPoolDataUsed*1024*1024)
						brickGaugeVecs[glusterThinPoolMetadataTotal].Set(thinLvmLbls, thinStat.ThinPoolMetadataTotal*1024*1024)
						brickGaugeVecs[glusterThinPoolMetadataUsed].Set(thinLvmLbls, thinStat.ThinPoolMetadataUsed*1024*1024)
					}
				}
			}
			effectiveCapacity := maxBrickUsed
			effectiveTotalCapacity := leastBrickTotal
			var subvolLabels = getGlusterSubvolLabels(volume.Name, subvol.Name)
			if subvol.Type == glusterconsts.SubvolTypeDisperse {
				// In disperse volume data bricks contribute to the sub
				// volume size
				effectiveCapacity = maxBrickUsed * float64(subvol.DisperseDataCount)
				effectiveTotalCapacity = leastBrickTotal * float64(subvol.DisperseDataCount)
			}

			// Export the metric only if available. it will be zero if the subvolume
			// contains only arbiter brick on current node or no local bricks on
			// this node
			if effectiveCapacity > 0 {
				brickGaugeVecs[glusterSubvolCapacityUsed].Set(subvolLabels, effectiveCapacity)
			}
			if effectiveTotalCapacity > 0 {
				brickGaugeVecs[glusterSubvolCapacityTotal].Set(subvolLabels, effectiveTotalCapacity)
			}
		}
	}
	return nil
}

func getBrickStatusLabels(vol string, host string, brickPath string, peerID string, pid int) prometheus.Labels {
	return prometheus.Labels{
		"cluster_id": clusterID,
		"volume":     vol,
		"hostname":   host,
		"brick_path": brickPath,
		"peer_id":    peerID,
		"pid":        strconv.Itoa(pid),
	}
}

func brickStatus(gluster glusterutils.GInterface) error {
	// Reset all vecs to not export stale information
	for _, gaugeVec := range brickStatusGaugeVecs {
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

	for _, volume := range volumes {
		// If volume is down, the bricks should be marked down
		var brickStatus []glusterutils.BrickStatus
		if volume.State != glusterconsts.VolumeStateStarted {
			for _, subvol := range volume.SubVolumes {
				for _, brick := range subvol.Bricks {
					status := glusterutils.BrickStatus{
						Hostname: brick.Host,
						PeerID:   brick.PeerID,
						Status:   0,
						PID:      0,
						Path:     brick.Path,
						Volume:   volume.Name,
					}
					brickStatus = append(brickStatus, status)
				}
			}
		} else {
			brickStatus, err = gluster.VolumeBrickStatus(volume.Name)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"volume": volume.Name,
				}).Debug("Error getting bricks status")
				continue
			}
		}
		for _, entry := range brickStatus {
			labels := getBrickStatusLabels(volume.Name, entry.Hostname, entry.Path, entry.PeerID, entry.PID)
			brickStatusGaugeVecs[glusterBrickUp].Set(labels, float64(entry.Status))
		}
	}

	return nil
}

func init() {
	registerMetric("gluster_brick", brickUtilization)
	registerMetric("gluster_brick_status", brickStatus)
}
