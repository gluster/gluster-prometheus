package main

import (
	"github.com/gluster/gluster-prometheus/pkg/glusterutils"
	// "github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// MyTestFunc is for testing
func MyTestFunc(gd glusterutils.GInterface) {
	var conf *glusterutils.Config
	if gdConf, ok := gd.(glusterutils.GDConfigInterface); ok {
		conf = gdConf.Config()
	}
	if conf != nil {
		log.Println("GD Management: ", conf.GlusterMgmt)
	}
}
