package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/gluster/gluster-prometheus/pkg/glusterutils"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

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
	LVAttr     string `json:"lv_attr"`
}

// NewPeerMetrics : provides a way to get the consolidated metrics (such PV, LV, VG counts)
func NewPeerMetrics() (*PeerMetrics, error) {
	cmdStr := "lvm vgs --noheading --reportformat=json -o lv_uuid,lv_name,pool_lv,vg_name,lv_path,lv_count,pv_count,pool_lv_uuid,lv_attr"
	outBs, err := glusterutils.ExecuteCmd(cmdStr)
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
			// by default set the LV count to Zero
			pMetrics.LVCountMap[vg.VGName] = 0
			if count, convErr := strconv.Atoi(strings.TrimSpace(vg.LVCount)); convErr == nil {
				// if there are no errors, update the LV count
				pMetrics.LVCountMap[vg.VGName] = count
			}
		}
		// before adding into 'thinPoolMap', check the attribute
		// if attribute string starts with 't', consider it as thinpool
		//
		// converting to a rune array because of 'utf-8' enconded
		// string handling in 'Go'
		if len(vg.LVAttr) > 0 && []rune(vg.LVAttr)[0] == 't' {
			// increment the thin pool count for that particular VG
			pMetrics.ThinPoolCountMap[vg.VGName]++
		}
	}
	return pMetrics, nil
}

var (
	// general metric labels
	gnrlMetricLabels = []MetricLabel{
		clusterIDLabel,
		{
			Name: "name",
			Help: "Metric name, for which data is collected",
		},
		{
			Name: "peerID",
			Help: "Peer ID of the host on which this metric is collected",
		},
	}
	// an additional information of 'vgName' is added
	// this specifies which Volume Group the LV or ThinPool count belongs
	withVgMetricLabels = []MetricLabel{
		clusterIDLabel,
		{
			Name: "name",
			Help: "Metric name, for which the data is collected",
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
		Help:      "No: of Physical Volumes",
		LongHelp:  "",
		Labels:    gnrlMetricLabels,
	})
	glusterLVCount = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "lv_count",
		Help:      "No: of Logical Volumes in a Volume Group",
		LongHelp:  "",
		Labels:    withVgMetricLabels,
	})
	glusterVGCount = newPrometheusGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "vg_count",
		Help:      "No: of Volume Groups",
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

func peerCounts(gluster glusterutils.GInterface) (err error) {
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
		log.WithError(err).WithFields(log.Fields{
			"peer": peerID,
		}).Errorln("[Peer_Metric_Count] Error:", err)
		return err
	}
	genrlLbls := prometheus.Labels{
		"cluster_id": clusterID,
		"name":       "Physical_Volumes",
		"peerID":     peerID,
	}
	glusterPVCount.With(genrlLbls).Set(float64(pMetrics.PVCount))
	genrlLbls = prometheus.Labels{
		"cluster_id": clusterID,
		"name":       "Volume_Groups",
		"peerID":     peerID,
	}
	glusterVGCount.With(genrlLbls).Set(float64(pMetrics.VGCount))
	// logical volume counts are added specific to each VG
	for vgName, lvCount := range pMetrics.LVCountMap {
		genrlLbls = prometheus.Labels{
			"cluster_id": clusterID,
			"name":       "Logical_Volumes",
			"peerID":     peerID,
			"vgName":     vgName,
		}
		glusterLVCount.With(genrlLbls).Set(float64(lvCount))
	}
	// similarly thinpool counts are also added per VG
	for vgName, tpCount := range pMetrics.ThinPoolCountMap {
		genrlLbls = prometheus.Labels{
			"cluster_id": clusterID,
			"name":       "ThinPool_Count",
			"peerID":     peerID,
			"vgName":     vgName,
		}
		glusterTPCount.With(genrlLbls).Set(float64(tpCount))
	}
	return
}

func init() {
	registerMetric("gluster_peer_counts", peerCounts)
}
