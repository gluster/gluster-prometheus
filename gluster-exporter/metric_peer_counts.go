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
	PVCount int
	LVMap   map[string]int
	VGCount int
}

type myVGDetails struct {
	LVUUID  string `json:"lv_uuid"`
	LVName  string `json:"lv_name"`
	PoolLV  string `json:"pool_lv"`
	VGName  string `json:"vg_name"`
	LVPath  string `json:"lv_path"`
	LVCount string `json:"lv_count"`
	PVCount string `json:"pv_count"`
}

// NewPeerMetrics : provides a way to get the consolidated metrics (such PV, LV, VG counts)
func NewPeerMetrics() (*PeerMetrics, error) {
	cmdStr := "lvm vgs --noheading --reportformat=json -o lv_uuid,lv_name,pool_lv,vg_name,lv_path,lv_count,pv_count"
	outBs, err := execCommand(cmdStr)
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(bytes.NewReader(outBs))
	var vgs []myVGDetails
	var vgMap = make(map[string]myVGDetails)
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
	for _, vg := range vgs {
		// collect the unique vgs into the map
		if _, ok := vgMap[vg.VGName]; !ok {
			vgMap[vg.VGName] = vg
		}
	}
	pMetrics := &PeerMetrics{VGCount: len(vgMap), LVMap: make(map[string]int)}
	for k, vg := range vgMap {
		log.Printf("%s: %+v\n", k, vg)
		if count, convErr := strconv.Atoi(strings.TrimSpace(vg.LVCount)); convErr == nil {
			pMetrics.LVMap[vg.VGName] = count
		}
		if count, convErr := strconv.Atoi(strings.TrimSpace(vg.PVCount)); convErr == nil {
			pMetrics.PVCount += count
		}
	}
	return pMetrics, nil
}

var (
	gnrlMetricLabels = []MetricLabel{
		{
			Name: "name",
			Help: "Name of the metric, whose count is collected (PV or LV or VG etc)",
		},
		{
			Name: "peerID",
			Help: "Peer ID of the host on which the metrics are collected",
		},
	}
	// an additional information of 'vgName' is added
	// this specifies which Volume Group the LV count belongs
	lvMetricLabels = []MetricLabel{
		{
			Name: "name",
			Help: "Logical Volume Metric details",
		},
		{
			Name: "peerID",
			Help: "Peer ID of the host on which the LV details are picked",
		},
		{
			Name: "vgName",
			Help: "Volume Group Name, shows no: of lvs under the specified Volume Group",
		},
	}

	glusterPVCountMetric = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "pv_count",
		Help:      "No: of physical volumes, got through pvs command",
		LongHelp:  "",
		Labels:    gnrlMetricLabels,
	})
	glusterLVCountMetric = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "lv_count",
		Help:      "No: of logical volumes, got through lvs command",
		LongHelp:  "",
		Labels:    lvMetricLabels,
	})
	glusterVGCountMetric = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "vg_count",
		Help:      "No: of volume groups, got through vgs command",
		LongHelp:  "",
		Labels:    gnrlMetricLabels,
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
		return err
	}
	log.Printf("Peer Metrics: %+v\n", pMetrics)
	genrlLbls := prometheus.Labels{
		"name":   "Physical_Volumes",
		"peerID": peerID,
	}
	glusterPVCountMetric.With(genrlLbls).Set(float64(pMetrics.PVCount))
	genrlLbls = prometheus.Labels{
		"name":   "Volume_Groups",
		"peerID": peerID,
	}
	glusterVGCountMetric.With(genrlLbls).Set(float64(pMetrics.VGCount))
	// logical volume counts are added specific to each VGs
	for vgName, lvCount := range pMetrics.LVMap {
		genrlLbls = prometheus.Labels{
			"name":   "Logical_Volumes",
			"peerID": peerID,
			"vgName": vgName,
		}
		glusterLVCountMetric.With(genrlLbls).Set(float64(lvCount))
	}
	return
}

func init() {
	registerMetric("gluster_peer_counts", peerCounts)
}
