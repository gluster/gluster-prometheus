package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os/exec"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// executes the command 'cmdStr'
// returns the byte array as output or error
func execCommand(cmdStr string) ([]byte, error) {
	cmdArr := strings.Fields(strings.TrimSpace(cmdStr))
	mainCmd, err := exec.LookPath(cmdArr[0])
	if err != nil {
		mainCmd = cmdArr[0]
	}
	outB, err := exec.Command(mainCmd, cmdArr[1:]...).Output()
	return outB, err
}

// PeerMetrics : exposes PV, LV, VG counts
type PeerMetrics struct {
	PVCount          int            // total Physical Volume counts
	LVCountMap       map[string]int // collects lv count for each Volume Group
	ThinPoolCountMap map[string]int // collects thinpool count for each Volume Group
	VGCount          int            // no: of Volume Groups
}

type myVGDetails struct {
	LVUUID     string `json:"lv_uuid"`
	LVName     string `json:"lv_name"`
	PoolLV     string `json:"pool_lv"`
	VGName     string `json:"vg_name"`
	LVPath     string `json:"lv_path"`
	LVCount    string `json:"lv_count"`
	PVCount    string `json:"pv_count"`
	PoolLVUUID string `json:"pool_lv_uuid"`
}

// NewPeerMetrics : provides a way to get the consolidated metrics (such PV, LV, VG counts)
func NewPeerMetrics() (*PeerMetrics, error) {
	cmdStr := "lvm vgs --noheading --reportformat=json -o lv_uuid,lv_name,pool_lv,vg_name,lv_path,lv_count,pv_count,pool_lv_uuid"
	outBs, err := execCommand(cmdStr)
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(bytes.NewReader(outBs))
	var vgs []myVGDetails
	// collect the details from the JSON output
	for {
		t, decErr := dec.Token()
		if decErr == io.EOF {
			break
		}
		if decErr != nil {
			err = errors.New("Unable to parse JSON output: " + err.Error())
			return nil, err
		}
		// if the token is 'vg', collect/decode the details into VGDetails array
		if t == "vg" {
			decErr = dec.Decode(&vgs)
		}
		if decErr != nil {
			err = errors.New("JSON output changed, parse failed: " + err.Error())
			return nil, err
		}
	}
	pMetrics := &PeerMetrics{
		PVCount:          0,
		VGCount:          0,
		LVCountMap:       make(map[string]int),
		ThinPoolCountMap: make(map[string]int),
	}
	var vgMap = make(map[string]myVGDetails)
	var thinPoolMap = make(map[string]myVGDetails)
	delim := "<<>>"
	for _, vg := range vgs {
		// collect the unique vgs into the map
		if _, ok := vgMap[vg.VGName]; !ok {
			vgMap[vg.VGName] = vg
			// increment the VG counter, for each new Volume Group
			pMetrics.VGCount++
			// if there is no error while integer conversion, add that number
			if count, convErr := strconv.Atoi(strings.TrimSpace(vg.PVCount)); convErr == nil {
				pMetrics.PVCount += count
			}
			// even if a parse error happens,
			// the function will return Zero by default
			pMetrics.LVCountMap[vg.VGName], _ = strconv.Atoi(strings.TrimSpace(vg.LVCount))
		}
		// logic to collect thinpool count in each Volume Group
		if vg.PoolLVUUID != "" {
			// in a Volume Group (VG), pool ID should be unique,
			// that means, combination of 'VGName + poolID' should be unique
			idsCombined := vg.VGName + delim + vg.PoolLVUUID
			if _, ok := thinPoolMap[idsCombined]; !ok {
				thinPoolMap[idsCombined] = vg
				// whenever a new unique 'VG+poolID' comes
				// increase the thin pool count for that particular VG
				pMetrics.ThinPoolCountMap[vg.VGName]++
			}
		}
	}
	return pMetrics, nil
}

var (
	// general metric labels
	gnrlMetricLabels = []MetricLabel{
		{
			Name: "name",
			Help: "Name of the metric, whose count is collected (PV or LV or VG etc)",
		},
		{
			Name: "peerID",
			Help: "Peer ID of the host on which this metric is collected",
		},
	}
	// an additional information of 'vgName' is added
	// this specifies which Volume Group the LV count belongs
	withVgMetricLabels = []MetricLabel{
		{
			Name: "name",
			Help: "Logical Volume Metric details",
		},
		{
			Name: "peerID",
			Help: "Peer ID of the host on which this metric is collected",
		},
		{
			Name: "vgName",
			Help: "Volume Group Name associated with the metric",
		},
	}

	glusterPVCount = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "pv_count",
		Help:      "No: of physical volumes, got through pvs command",
		LongHelp:  "",
		Labels:    gnrlMetricLabels,
	})
	glusterLVCount = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "lv_count",
		Help:      "No: of logical volumes, got through lvs command",
		LongHelp:  "",
		Labels:    withVgMetricLabels,
	})
	glusterVGCount = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "vg_count",
		Help:      "No: of volume groups, got through vgs command",
		LongHelp:  "",
		Labels:    gnrlMetricLabels,
	})
	glusterTPCount = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "thinpool_count",
		Help:      "No: of thinpools in a Volume Group",
		LongHelp:  "",
		Labels:    withVgMetricLabels,
	})
)

func peerCounts() (err error) {
	var peerID string
	// 'gluster' is initialized inside 'main' function,
	// so it is better to check whether it is available or not
	if gluster != nil {
		if peerID, err = gluster.LocalPeerID(); err != nil {
			return
		}
	}
	pMetrics, err := NewPeerMetrics()
	if err != nil {
		// log the error and then return
		log.Errorln("[Peer_Metric_Count] Error:", err)
		return err
	}
	genrlLbls := prometheus.Labels{
		"name":   "Physical_Volumes",
		"peerID": peerID,
	}
	glusterPVCount.With(genrlLbls).Set(float64(pMetrics.PVCount))
	genrlLbls = prometheus.Labels{
		"name":   "Volume_Groups",
		"peerID": peerID,
	}
	glusterVGCount.With(genrlLbls).Set(float64(pMetrics.VGCount))
	// logical volume counts are added specific to each VG
	for vgName, lvCount := range pMetrics.LVCountMap {
		genrlLbls = prometheus.Labels{
			"name":   "Logical_Volumes",
			"peerID": peerID,
			"vgName": vgName,
		}
		glusterLVCount.With(genrlLbls).Set(float64(lvCount))
	}
	// similarly thinpool counts are also added per VG
	for vgName, tpCount := range pMetrics.ThinPoolCountMap {
		genrlLbls = prometheus.Labels{
			"name":   "ThinPool_Count",
			"peerID": peerID,
			"vgName": vgName,
		}
		glusterTPCount.With(genrlLbls).Set(float64(tpCount))
	}
	return
}

func init() {
	registerMetric("gluster_peer_counts", peerCounts)
}
