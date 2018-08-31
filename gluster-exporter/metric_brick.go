package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gluster/gluster-prometheus/pkg/glusterutils"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

var (
	brickLabels = []MetricLabel{
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
	}

	glusterBrickCapacityUsed = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "brick_capacity_used_bytes",
		Help:      "Used capacity of gluster bricks in bytes",
		LongHelp:  "",
		Labels:    brickLabels,
	})

	glusterBrickCapacityFree = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "brick_capacity_free_bytes",
		Help:      "Free capacity of gluster bricks in bytes",
		LongHelp:  "",
		Labels:    brickLabels,
	})

	glusterBrickCapacityTotal = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "brick_capacity_bytes_total",
		Help:      "Total capacity of gluster bricks in bytes",
		LongHelp:  "",
		Labels:    brickLabels,
	})

	glusterBrickInodesTotal = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "brick_inodes_total",
		Help:      "Total no of inodes of gluster brick disk",
		LongHelp:  "",
		Labels:    brickLabels,
	})

	glusterBrickInodesFree = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "brick_inodes_free",
		Help:      "Free no of inodes of gluster brick disk",
		LongHelp:  "",
		Labels:    brickLabels,
	})

	glusterBrickInodesUsed = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "brick_inodes_used",
		Help:      "Used no of inodes of gluster brick disk",
		LongHelp:  "",
		Labels:    brickLabels,
	})

	glusterSubvolCapacityUsed = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "subvol_capacity_used_bytes",
		Help:      "Effective used capacity of gluster subvolume in bytes",
		LongHelp:  "",
		Labels:    subvolLabels,
	})

	glusterBrickLVSize = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "brick_lv_size_bytes",
		Help:      "Bricks LV size Bytes",
		LongHelp:  "",
		Labels:    lvmLbls,
	})

	glusterBrickLVPercent = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "brick_lv_percent",
		Help:      "Bricks LV usage percent",
		LongHelp:  "",
		Labels:    lvmLbls,
	})

	glusterBrickLVMetadataSize = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "brick_lv_metadata_size_bytes",
		Help:      "Bricks LV metadata size Bytes",
		LongHelp:  "",
		Labels:    lvmLbls,
	})

	glusterBrickLVMetadataPercent = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "brick_lv_metadata_percent",
		Help:      "Bricks LV metadata usage percent",
		LongHelp:  "",
		Labels:    lvmLbls,
	})

	brickDiskLabels = []MetricLabel{
		{
			Name: "host",
			Help: "Host FDQN or IP",
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
			Name: "device",
			Help: "Device Name",
		},
		{
			Name: "disk",
			Help: "Physical disk name",
		},
	}

	glusterDiskReadIOs = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "disk_read_ios",
		Help:      "Brick disk's read IOs",
		LongHelp:  "",
		Labels:    brickDiskLabels,
	})

	glusterDiskReadMerges = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "disk_read_merges",
		Help:      "Brick disk's read merges",
		LongHelp:  "",
		Labels:    brickDiskLabels,
	})

	glusterDiskReadSectors = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "disk_read_sectors",
		Help:      "Brick disk's read sectors",
		LongHelp:  "",
		Labels:    brickDiskLabels,
	})

	glusterDiskReadTicks = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "disk_read_ticks",
		Help:      "Brick disk's read ticks",
		LongHelp:  "",
		Labels:    brickDiskLabels,
	})

	glusterDiskWriteIOs = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "disk_write_ios",
		Help:      "Brick disk's write IOs",
		LongHelp:  "",
		Labels:    brickDiskLabels,
	})

	glusterDiskWriteMerges = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "disk_write_merges",
		Help:      "Brick disk's write_merges",
		LongHelp:  "",
		Labels:    brickDiskLabels,
	})

	glusterDiskWriteSectors = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "disk_write_sectors",
		Help:      "Brick disk's write sectors",
		LongHelp:  "",
		Labels:    brickDiskLabels,
	})

	glusterDiskWriteTicks = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "disk_write_ticks",
		Help:      "Brick disk's write_ticks",
		LongHelp:  "",
		Labels:    brickDiskLabels,
	})

	glusterDiskInflight = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "disk_inflight",
		Help:      "Brick disk's inflight no of requests",
		LongHelp:  "",
		Labels:    brickDiskLabels,
	})

	glusterDiskTotalTicks = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "disk_total_ticks",
		Help:      "Brick disk's total ticks (millis)",
		LongHelp:  "",
		Labels:    brickDiskLabels,
	})

	glusterDiskTimeInqueue = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "disk_timeinqueue",
		Help:      "Brick disk's time in queue (millis)",
		LongHelp:  "",
		Labels:    brickDiskLabels,
	})
)

func getGlusterBrickLabels(brick glusterutils.Brick, subvol string) prometheus.Labels {
	return prometheus.Labels{
		"host":       brick.Host,
		"id":         brick.ID,
		"brick_path": brick.Path,
		"volume":     brick.VolumeName,
		"subvolume":  subvol,
	}
}

func getGlusterSubvolLabels(volname string, subvol string) prometheus.Labels {
	return prometheus.Labels{
		"volume":    volname,
		"subvolume": subvol,
	}
}

func getBrickDiskLabels(
	brick glusterutils.Brick,
	subvol string,
	volume string,
	device string,
	disk string) prometheus.Labels {
	return prometheus.Labels{
		"host":       brick.Host,
		"id":         brick.ID,
		"brick_path": brick.Path,
		"volume":     volume,
		"subvolume":  subvol,
		"device":     device,
		"disk":       disk,
	}
}

func getBrickDevice(brickPath string) string {
	cmd := fmt.Sprintf("df --output=source %s", brickPath)
	command := exec.Command(cmd)
	out := &bytes.Buffer{}
	command.Stdout = out
	err := command.Start()
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"command": cmd,
		}).Error("Error executing command")
		return ""
	}

	// Create a ticker that outputs elapsed time
	ticker := time.NewTicker(time.Second)
	// Wait for 5 seconds for output, else kill the process
	timer := time.NewTimer(time.Second * 5)
	go func(timer *time.Timer, ticker *time.Ticker, cmd *exec.Cmd) {
		for _ = range timer.C {
			err := cmd.Process.Signal(os.Kill)
			log.WithError(err).WithFields(log.Fields{
				"command": cmd,
			}).Error("Command timed out")
			ticker.Stop()
			return
		}
	}(timer, ticker, command)

	// Procedd only if process has finished
	command.Wait()

	dfData := strings.Split(strings.TrimSpace(string(out.Bytes())), "\n")[1]
	fInfo, err := os.Lstat(dfData)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"brickpath": brickPath,
		}).Error("Error getting df data")
		return ""
	}
	return fInfo.Name()
}

func getPhysicalDisks(blockDevice string) []string {
	cmd := "lsblk --all --bytes --raw --noheadings --output NAME,PKNAME,TYPE"
	out, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"command": cmd,
		}).Error("Error executing command")
		return []string{}
	}
	parentDetails := make(map[string][]string)
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if line != "" {
			fields := strings.Split(line, " ")
			if fields[1] != "" {
				parentDetails[fields[0]] = append(parentDetails[fields[0]], fields[1])
			}
		}
	}
	sl := strings.Split(blockDevice, "/")
	dev := sl[len(sl)-1]
	parents := parentDetails[dev]
	ret_val := []string{}
	for _, parent := range parents {
		for {
			if _, exists := parentDetails[parent]; exists {
				// now on there would be one entry for parent
				parent = parentDetails[parent][0]
			} else {
				ret_val = append(ret_val, parent)
				break
			}
		}
	}
	return ret_val
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
}

func getLVS() ([]LVMStat, error) {
	cmd := "lvm vgs --unquoted --reportformat=json --noheading --nosuffix --units m -o lv_uuid,lv_name,data_percent,pool_lv,lv_attr,lv_size,lv_path,lv_metadata_size,metadata_percent,vg_name"

	out, err := exec.Command("sh", "-c", cmd).Output()
	lvmDet := []LVMStat{}
	if err != nil {
		log.WithError(err).Error("Error getting lvm usage details")
		return lvmDet, err
	}
	var vgReport VGReport
	if err1 := json.Unmarshal(out, &vgReport); err1 != nil {
		log.WithError(err1).Error("Error parsing lvm usage details")
		return lvmDet, err1
	}

	for _, vg := range vgReport.Report[0].Vgs {
		var obj LVMStat
		obj.UUID = vg.LVUUID
		obj.Name = vg.LVName
		if vg.DataPercent == "" {
			obj.DataPercent = 0.0
		} else {
			if obj.DataPercent, err = strconv.ParseFloat(vg.DataPercent, 64); err != nil {
				log.WithError(err).Error("Error parsing DataPercent value of lvm usage")
				return lvmDet, err
			}
		}
		obj.PoolLV = vg.PoolLV
		obj.Attr = vg.LVAttr
		if vg.LVSize == "" {
			obj.Size = 0.0
		} else {
			if obj.Size, err = strconv.ParseFloat(vg.LVSize, 64); err != nil {
				log.WithError(err).Error("Error parsing LVSize value of lvm usage")
				return lvmDet, err
			}
		}
		obj.Path = vg.LVPath
		if vg.LVMetadataSize == "" {
			obj.MetadataSize = 0.0
		} else {
			if obj.MetadataSize, err = strconv.ParseFloat(vg.LVMetadataSize, 64); err != nil {
				log.WithError(err).Error("Error parsing LVMetadataSize value of lvm usage")
				return lvmDet, err
			}
		}
		if vg.MetadataPercent == "" {
			obj.MetadataPercent = 0.0
		} else {
			obj.MetadataPercent, err = strconv.ParseFloat(vg.MetadataPercent, 64)
			if err != nil {
				log.WithError(err).Error("Error parsing MetadataPercent value of lvm usage")
				return lvmDet, err
			}
		}
		obj.VGName = vg.VGName
		if obj.Attr[0] == 't' {
			obj.Device = fmt.Sprintf("%s/%s", obj.VGName, obj.Name)
		} else {
			obj.Device, err = filepath.EvalSymlinks(obj.Path)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"path": obj.Path,
				}).Error("Error evaluating realpath")
				return lvmDet, err
			}
		}
		lvmDet = append(lvmDet, obj)
	}
	return lvmDet, nil
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
		"host":       brick.Host,
		"id":         brick.ID,
		"brick_path": brick.Path,
		"volume":     brick.VolumeName,
		"subvolume":  subvol,
		"vg_name":    stat.VGName,
		"lv_path":    stat.Path,
	}
}

func lvmUsage(path string) (stats []LVMStat, err error) {
	mountPoints, err := parseProcMounts()
	if err != nil {
		return stats, err
	}
	lvs, err := getLVS()
	if err != nil {
		return stats, err
	}
	for _, lv := range lvs {
		for _, mount := range mountPoints {
			dev, err := filepath.EvalSymlinks(mount.Device)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"path": mount.Device,
				}).Error("Error evaluating realpath")
				continue
			}
			if lv.Device == dev {
				stats = append(stats, lv)
			}
		}
	}
	return stats, nil
}

// BlockDeviceStat represents block device stats
type BlockDeviceStat struct {
	ReadIos      uint64
	ReadMerges   uint64
	ReadSectors  uint64
	ReadTicks    uint64
	WriteIos     uint64
	WriteMerges  uint64
	WriteSectors uint64
	WriteTicks   uint64
	InFlight     uint64
	TotalTicks   uint64
	TimeInqueue  uint64
}

func blockStat(blockDev string) (stat BlockDeviceStat) {
	fileObj, err := os.Open(fmt.Sprintf("/sys/block/%s/stat", blockDev))
	if err != nil {
		log.WithError(err).Error("Error reading block stats")
		return
	}
	i, err := fmt.Fscanf(fileObj, "%d %d %d %d %d %d %d %d %d %d %d",
		&stat.ReadIos,
		&stat.ReadMerges,
		&stat.ReadSectors,
		&stat.ReadTicks,
		&stat.WriteIos,
		&stat.WriteMerges,
		&stat.WriteSectors,
		&stat.WriteTicks,
		&stat.InFlight,
		&stat.TotalTicks,
		&stat.TimeInqueue,
	)
	if i == 0 || err != nil {
		log.WithError(err).Error("Error parsing block stats")
		return
	}
	return
}

func brickUtilization() error {
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
		if volume.State != glusterutils.VolumeStateStarted {
			// Export brick metrics only if the Volume
			// is is in Started state
			continue
		}
		subvols := volume.SubVolumes
		for _, subvol := range subvols {
			bricks := subvol.Bricks
			var maxBrickUsed float64
			for _, brick := range bricks {
				if brick.PeerID == localPeerID {
					usage, err := diskUsage(brick.Path)
					if err != nil {
						log.WithError(err).WithFields(log.Fields{
							"volume":     volume.Name,
							"brick_path": brick.Path,
						}).Error("Error getting disk usage")
						continue
					}
					var lbls = getGlusterBrickLabels(brick, subvol.Name)
					// Update the metrics
					glusterBrickCapacityUsed.With(lbls).Set(usage.Used)
					glusterBrickCapacityFree.With(lbls).Set(usage.Free)
					glusterBrickCapacityTotal.With(lbls).Set(usage.All)
					glusterBrickInodesTotal.With(lbls).Set(usage.InodesAll)
					glusterBrickInodesFree.With(lbls).Set(usage.InodesFree)
					glusterBrickInodesUsed.With(lbls).Set(usage.InodesUsed)
					// Skip exporting utilization data in case of arbiter
					// brick to avoid wrong values when both the data bricks
					// are down
					if brick.Type != glusterutils.BrickTypeArbiter && usage.Used >= maxBrickUsed {
						maxBrickUsed = usage.Used
					}
					// Get lvm usage details
					stats, err := lvmUsage(brick.Path)
					if err != nil {
						log.WithError(err).WithFields(log.Fields{
							"volume":     volume.Name,
							"brick_path": brick.Path,
						}).Error("Error getting lvm usage")
						continue
					}
					// Add metrics
					for _, stat := range stats {
						var lvmLbls = getGlusterLVMLabels(brick, subvol.Name, stat)
						// Convert to bytes
						glusterBrickLVSize.With(lvmLbls).Set(stat.Size * 1024 * 1024)
						glusterBrickLVPercent.With(lvmLbls).Set(stat.DataPercent)
						// Convert to bytes
						glusterBrickLVMetadataSize.With(lvmLbls).Set(stat.MetadataSize * 1024 * 1024)
						glusterBrickLVMetadataPercent.With(lvmLbls).Set(stat.MetadataPercent)
					}

					bDev := getBrickDevice(brick.Path)
					disks := getPhysicalDisks(bDev)
					for _, disk := range disks {
						stat := blockStat(disk)
						var brickDiskLbls = getBrickDiskLabels(
							brick,
							subvol.Name,
							volume.Name,
							bDev,
							disk)
						glusterDiskReadIOs.With(brickDiskLbls).
							Set(float64(stat.ReadIos))
						glusterDiskReadMerges.With(brickDiskLbls).
							Set(float64(stat.ReadMerges))
						glusterDiskReadSectors.With(brickDiskLbls).
							Set(float64(stat.ReadSectors))
						glusterDiskReadTicks.With(brickDiskLbls).
							Set(float64(stat.ReadTicks))
						glusterDiskWriteIOs.With(brickDiskLbls).
							Set(float64(stat.WriteIos))
						glusterDiskWriteMerges.With(brickDiskLbls).
							Set(float64(stat.WriteMerges))
						glusterDiskWriteSectors.With(brickDiskLbls).
							Set(float64(stat.WriteSectors))
						glusterDiskWriteTicks.With(brickDiskLbls).
							Set(float64(stat.WriteTicks))
						glusterDiskInflight.With(brickDiskLbls).
							Set(float64(stat.InFlight))
						glusterDiskTotalTicks.With(brickDiskLbls).
							Set(float64(stat.TotalTicks))
						glusterDiskTimeInqueue.With(brickDiskLbls).
							Set(float64(stat.TimeInqueue))
					}
				}
			}
			effectiveCapacity := maxBrickUsed
			var subvolLabels = getGlusterSubvolLabels(volume.Name, subvol.Name)
			if subvol.Type == glusterutils.SubvolTypeDisperse {
				// In disperse volume data bricks contribute to the sub
				// volume size
				effectiveCapacity = maxBrickUsed * float64(subvol.DisperseDataCount)
			}

			// Export the metric only if available. it will be zero if the subvolume
			// contains only arbiter brick on current node or no local bricks on
			// this node
			if effectiveCapacity > 0 {
				glusterSubvolCapacityUsed.With(subvolLabels).Set(effectiveCapacity)
			}
		}
	}
	return nil
}

func init() {
	registerMetric("gluster_brick", brickUtilization)
}
