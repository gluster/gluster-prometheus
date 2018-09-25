package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"syscall"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	brickLabels = []string{
		"host",
		"id",
		"brick_path",
		"volume",
		"subvolume",
	}

	subvolLabels = []string{
		"volume",
		"subvolume",
	}

	glusterBrickCapacityUsed = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "gluster",
			Name:      "brick_capacity_used",
			Help:      "Used capacity of gluster bricks",
		},
		brickLabels,
	)

	glusterBrickCapacityFree = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "gluster",
			Name:      "brick_capacity_free",
			Help:      "Free capacity of gluster bricks",
		},
		brickLabels,
	)

	glusterBrickCapacityTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "gluster",
			Name:      "brick_capacity_total",
			Help:      "Total capacity of gluster bricks",
		},
		brickLabels,
	)

	glusterBrickInodesTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "gluster",
			Name:      "brick_inodes_total",
			Help:      "Total no of inodes of gluster brick disk",
		},
		brickLabels,
	)

	glusterBrickInodesFree = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "gluster",
			Name:      "brick_inodes_free",
			Help:      "Free no of inodes of gluster brick disk",
		},
		brickLabels,
	)

	glusterBrickInodesUsed = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "gluster",
			Name:      "brick_inodes_used",
			Help:      "Used no of inodes of gluster brick disk",
		},
		brickLabels,
	)

	glusterSubvolCapacityUsed = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "gluster",
			Name:      "subvol_capacity_used",
			Help:      "Effective used capacity of gluster subvolume",
		},
		subvolLabels,
	)
)

func getGlusterBrickLabels(brick GlusterBrick, subvol string) prometheus.Labels {
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

// DiskStatus represents Disk usage
type DiskStatus struct {
	All        float64 `json:"all"`
	Used       float64 `json:"used"`
	Free       float64 `json:"free"`
	InodesAll  float64 `json:"inodesall"`
	InodesFree float64 `json:"inodesfree"`
	InodesUsed float64 `json:"inodesused"`
}

// GlusterBrick represents brick info
type GlusterBrick struct {
	Host       string `json:"host"`
	ID         string `json:"id"`
	Path       string `json:"path"`
	PeerID     string `json:"peer-id"`
	Type       string `json:"type"`
	VolumeID   string `json:"volume-id"`
	VolumeName string `json:"volume-name"`
}

// GlusterSubVolume represents sub volume of distribution set
type GlusterSubVolume struct {
	ArbiterCount      uint           `json:"arbiter-count"`
	Bricks            []GlusterBrick `json:"bricks"`
	DisperseCount     uint           `json:"disperse-count"`
	DisperseDataCount uint           `json:"disperse-data-count"`
	Name              string         `json:"name"`
	ReplicaCount      uint           `json:"replica-count"`
	Type              string         `json:"type"`
}

// GlusterVolume represents Gluster Volume
type GlusterVolume struct {
	DistributeCount uint               `json:"distribute-count"`
	ID              string             `json:"id"`
	Metadata        map[string]string  `json:"metadata"`
	Name            string             `json:"name"`
	Options         map[string]string  `json:"options"`
	ReplicCount     uint               `json:"replica-count"`
	SnapList        []string           `json:"snap-list"`
	State           string             `json:"state"`
	SubVolumes      []GlusterSubVolume `json:"subvols"`
	Transport       string             `json:"transport"`
	Type            string             `json:"type"`
}

func diskUsage(path string) (disk DiskStatus) {
	fs := syscall.Statfs_t{}
	err := syscall.Statfs(path, &fs)
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

func brickUtilization() {
	fileObj, err := os.Open(getVolInfoFile())
	if err != nil {
		// TODO: log the error
		return
	}
	defer fileObj.Close()
	data, _ := ioutil.ReadAll(fileObj)
	var volumes []GlusterVolume
	if err := json.Unmarshal(data, &volumes); err != nil {
		return
	}
	for _, volume := range volumes {
		subvols := volume.SubVolumes
		for _, subvol := range subvols {
			bricks := subvol.Bricks
			var maxBrickUsed float64
			for _, brick := range bricks {

				// TODO: Handle error
				peerID, _ = getPeerID()

				if brick.PeerID == peerID {
					usage := diskUsage(brick.Path)
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
					if brick.Type != "Arbiter" && usage.Used >= maxBrickUsed {
						maxBrickUsed = usage.Used
					}
				}
			}
			effectiveCapacity := maxBrickUsed
			var subvolLabels = getGlusterSubvolLabels(volume.Name, subvol.Name)
			if subvol.Type == "Disperse" {
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
}

func init() {
	prometheus.MustRegister(glusterBrickCapacityUsed)
	prometheus.MustRegister(glusterBrickCapacityFree)
	prometheus.MustRegister(glusterBrickCapacityTotal)
	prometheus.MustRegister(glusterBrickInodesTotal)
	prometheus.MustRegister(glusterBrickInodesFree)
	prometheus.MustRegister(glusterBrickInodesUsed)
	prometheus.MustRegister(glusterSubvolCapacityUsed)

	// Register to update this every 2 seconds
	// Name, Callback Func, Interval Seconds
	registerMetric("gluster_brick", brickUtilization, 5)
}
