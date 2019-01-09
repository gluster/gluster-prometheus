package main

import (
	"errors"

	"github.com/gluster/gluster-prometheus/pkg/glusterutils"
)

func cachedVolumeInfo(gluster glusterutils.GInterface, ttl int) ([]glusterutils.Volume, error) {
	volinfoFn := func() (interface{}, error) {
		return gluster.VolumeInfo()
	}
	data, err := glusterutils.CachedOutput("volumes", volinfoFn, ttl)
	if err != nil {
		volumes, ok := data.([]glusterutils.Volume)
		if ok {
			return volumes, nil
		}
		return nil, errors.New("unable to type cast volume info")
	}
	return nil, err
}

func cachedIsLeader(gluster glusterutils.GInterface, ttl int) (bool, error) {
	leaderFn := func() (interface{}, error) {
		return gluster.IsLeader()
	}
	data, err := glusterutils.CachedOutput("leader", leaderFn, ttl)
	if err != nil {
		isleader, ok := data.(bool)
		if ok {
			return isleader, nil
		}
		return false, errors.New("unable to type cast isLeader output")
	}
	return false, err
}
