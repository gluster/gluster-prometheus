package glusterutils

// VolumeStatus returns gluster vol status (glusterd2)
func (g *GD2) VolumeStatus() ([]VolumeStatus, error) {
	client, err := initRESTClient(g.config)
	if err != nil {
		return nil, err
	}
	// We have to fetch the list of volumes first...
	volumelist, err := client.Volumes("")
	if err != nil {
		return nil, err
	}
	volumestatus := make([]VolumeStatus, len(volumelist))
	for idx, vol := range volumelist {
		// ...and the detailed brick statuses individually for each
		// volume, because the GD2 REST API does not have a "give me
		// detailed status information for all volumes" endpoint.
		brickstatusinfo, err := client.BricksStatus(vol.Name)
		if err != nil {
			return nil, err
		}
		brickstatus := make([]BrickStatus, len(brickstatusinfo))
		for idx, info := range brickstatusinfo {
			brickStatusObj := BrickStatus{
				Hostname:       info.Info.Hostname,
				PeerID:         info.Info.PeerID.String(),
				PID:            info.Pid,
				Port:           info.Port,
				Path:           info.Info.Path,
				Volume:         vol.Name,
				Capacity:       info.Size.Capacity,
				Free:           info.Size.Free,
				Gd1InodesFree:  -1, // Inode data n/a in GD2 response
				Gd1InodesTotal: -1, // Inode data n/a in GD2 response
			}
			if info.Online {
				brickStatusObj.Status = 1
			} else {
				brickStatusObj.Status = 0
			}
			brickstatus[idx] = brickStatusObj
		}
		volumestatus[idx].Name = vol.Name
		volumestatus[idx].Nodes = brickstatus
	}
	return volumestatus, nil
}
