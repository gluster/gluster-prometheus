package glusterutils

// VolumeBrickStatus gets brick status info from glusterd2 using rest api
func (g GD2) VolumeBrickStatus(vol string) ([]BrickStatus, error) {
	client, err := initRESTClient(g.config)
	if err != nil {
		return nil, err
	}
	brickstatusinfo, err := client.BricksStatus(vol)
	if err != nil {
		return nil, err
	}
	brickstatus := make([]BrickStatus, len(brickstatusinfo))
	for idx, info := range brickstatusinfo {
		brickStatusObj := BrickStatus{
			Hostname: info.Info.Hostname,
			PeerID:   info.Info.PeerID.String(),
			PID:      info.Pid,
			Path:     info.Info.Path,
			Volume:   vol,
		}
		if info.Online {
			brickStatusObj.Status = 1
		} else {
			brickStatusObj.Status = 0
		}
		brickstatus[idx] = brickStatusObj
	}
	return brickstatus, nil
}
