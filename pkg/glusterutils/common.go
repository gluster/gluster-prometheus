package glusterutils

import (
	"time"

	"github.com/gluster/gluster-prometheus/gluster-exporter/conf"
	"github.com/gluster/glusterd2/pkg/restclient"
)

func initRESTClient(config *conf.GConfig) (*restclient.Client, error) {
	client, err := restclient.New(
		config.Glusterd2Endpoint,
		config.Glusterd2User,
		config.Glusterd2Secret,
		config.Glusterd2Cacert,
		config.Glusterd2Insecure,
	)
	if err != nil {
		return nil, err
	}
	client.SetTimeout(time.Duration(config.Timeout) * time.Second)
	return client, nil
}

func setDefaultConfig(config *conf.GConfig) {
	if config.Timeout == 0 {
		config.Timeout = 30
	}
	if config.GlusterMgmt == "" {
		config.GlusterMgmt = "glusterd"
	}
	if config.GlusterCmd == "" {
		config.GlusterCmd = "gluster"
	}
	if config.GlusterRemoteHost == "" {
		config.GlusterRemoteHost = "localhost"
	}
	if config.Glusterd2Endpoint == "" {
		config.Glusterd2Endpoint = "http://localhost:24007"
	}
}
