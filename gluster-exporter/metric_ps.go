package main

import (
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	glusterProcs = []string{
		"glusterd",
		"glusterfsd",
		"glusterd2",
		// TODO: Add more processes
	}

	labels = []string{
		"volume",
		"peerid",
		"brick_path",
		"name",
	}

	glusterCPUPercentage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "gluster",
			Name:      "cpu_percentage",
			Help:      "CPU Percentage used by Gluster processes",
		},
		labels,
	)

	glusterMemoryPercentage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "gluster",
			Name:      "memory_percentage",
			Help:      "Memory Percentage used by Gluster processes",
		},
		labels,
	)

	glusterResidentMemory = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "gluster",
			Name:      "resident_memory",
			Help:      "Resident Memory of Gluster processes",
		},
		labels,
	)

	glusterVirtualMemory = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "gluster",
			Name:      "virtual_memory",
			Help:      "Virtual Memory of Gluster processes",
		},
		labels,
	)

	glusterElapsedTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "gluster",
			Name:      "elapsed_time_seconds",
			Help:      "Elapsed Time of Gluster processes",
		},
		labels,
	)
)

func getCmdLine(pid string) ([]string, error) {
	var args []string

	out, err := ioutil.ReadFile(filepath.Clean("/proc/" + pid + "/cmdline"))
	if err != nil {
		return args, err
	}

	return strings.Split(strings.Trim(string(out), "\x00"), "\x00"), nil
}

func getGlusterdLabels(cmd string, args []string) (prometheus.Labels, error) {

	peerID, err := gluster.LocalPeerID()
	if err != nil {
		return nil, err
	}

	return prometheus.Labels{
		"name":       cmd,
		"volume":     "",
		"peerid":     peerID,
		"brick_path": "",
	}, nil
}

func getGlusterFsdLabels(cmd string, args []string) (prometheus.Labels, error) {
	bpath := ""
	volume := ""

	prevArg := ""

	// TODO: Handle error
	peerID, err := gluster.LocalPeerID()
	if err != nil {
		return nil, err
	}

	for _, a := range args {
		if prevArg == "--brick-name" {
			bpath = a
		} else if prevArg == "--volfile-id" {
			volume = strings.Split(a, ".")[0]
		}
		prevArg = a
	}

	return prometheus.Labels{
		"name":       cmd,
		"volume":     volume,
		"peerid":     peerID,
		"brick_path": bpath,
	}, nil
}

func getUnknownLabels(cmd string, args []string) (prometheus.Labels, error) {

	// TODO: Handle error
	peerID, err := gluster.LocalPeerID()
	if err != nil {
		return nil, err
	}

	return prometheus.Labels{
		"name":       cmd,
		"volume":     "",
		"peerid":     peerID,
		"brick_path": "",
	}, nil
}

func ps() {
	args := []string{
		"--no-header", // No header in the output
		"-ww",         // To set unlimited width to avoid crop
		"-o",          // Output Format
		"pid,pcpu,pmem,rsz,vsz,etimes,comm",
		"-C",
		strings.Join(glusterProcs, ","),
	}

	out, err := exec.Command("ps", args...).Output()

	if err != nil {
		// TODO: Log error
		// Return without exporting metrics in this cycle
		return
	}

	for _, line := range strings.Split(string(out), "\n") {
		// Sample data:
		// 6959  0.0  0.6 12840 713660  504076 glusterfs
		lineDataTmp := strings.Split(line, " ")
		lineData := []string{}
		for _, d := range lineDataTmp {
			if strings.Trim(d, " ") == "" {
				continue
			}
			lineData = append(lineData, d)
		}

		if len(lineData) < 7 {
			continue
		}
		cmdlineArgs, err := getCmdLine(lineData[0])
		if err != nil {
			// TODO: Log error
			continue
		}

		if len(cmdlineArgs) == 0 {
			// No cmdline file, may be that process died
			continue
		}

		var lbls prometheus.Labels
		switch lineData[6] {
		case "glusterd":
			lbls, err = getGlusterdLabels(lineData[6], cmdlineArgs)
		case "glusterd2":
			lbls, err = getGlusterdLabels(lineData[6], cmdlineArgs)
		case "glusterfsd":
			lbls, err = getGlusterFsdLabels(lineData[6], cmdlineArgs)
		default:
			lbls, err = getUnknownLabels(lineData[6], cmdlineArgs)
		}

		if err != nil {
			// TODO: Log error
			continue
		}

		pcpu, err := strconv.ParseFloat(lineData[1], 64)
		if err != nil {
			// TODO: Log Error
			continue
		}

		pmem, err := strconv.ParseFloat(lineData[2], 64)
		if err != nil {
			// TODO: Log Error
			continue
		}
		rsz, err := strconv.ParseFloat(lineData[3], 64)
		if err != nil {
			// TODO: Log Error
			continue
		}

		vsz, err := strconv.ParseFloat(lineData[4], 64)
		if err != nil {
			// TODO: Log Error
			continue
		}

		etimes, err := strconv.ParseFloat(lineData[5], 64)
		if err != nil {
			// TODO: Log Error
			continue
		}

		// Update the Metrics
		glusterCPUPercentage.With(lbls).Set(pcpu)
		glusterMemoryPercentage.With(lbls).Set(pmem)
		glusterResidentMemory.With(lbls).Set(rsz)
		glusterVirtualMemory.With(lbls).Set(vsz)
		glusterElapsedTime.With(lbls).Set(etimes)
	}
}

func init() {
	prometheus.MustRegister(glusterCPUPercentage)
	prometheus.MustRegister(glusterMemoryPercentage)
	prometheus.MustRegister(glusterResidentMemory)
	prometheus.MustRegister(glusterVirtualMemory)
	prometheus.MustRegister(glusterElapsedTime)

	// Register to update this every 2 seconds
	// Name, Callback Func, Interval Seconds
	registerMetric("gluster_ps", ps)
}
