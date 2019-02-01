package glusterutils

import (
	"encoding/xml"
)

// VolumeBrickStatus gets brick status info from glusterd2 using rest api
func (g GD1) VolumeBrickStatus(vol string) ([]BrickStatus, error) {
	// Run gluster volume status {vol}
	out, err := g.execGluster("volume", "status", vol)
	if err != nil {
		return nil, err
	}
	var volStatus gd1VolumeStatus
	err = xml.Unmarshal(out, &volStatus)
	if err != nil {
		return nil, err
	}

	var brickstatus []BrickStatus
	if len(volStatus.List) > 0 {
		for _, process := range volStatus.List[0].NodeProcesses {
			if process.Hostname != "Self-heal Daemon" {
				brickStatusObj := BrickStatus{
					Hostname: process.Hostname,
					PeerID:   process.PeerID,
					Status:   process.Status,
					PID:      process.PID,
					Path:     process.Path,
					Volume:   vol,
				}
				brickstatus = append(brickstatus, brickStatusObj)
			}
		}
	}
	return brickstatus, nil
}
