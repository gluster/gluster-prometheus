package glusterutils

import (
	"github.com/gluster/glusterd2/pkg/api"
)

// Snapshots returns snaphosts list for the cluster
func (g *GD2) Snapshots() ([]Snapshot, error) {
	client, err := initRESTClient(g.config)
	if err != nil {
		return nil, err
	}
	snapListResp, err := client.SnapshotList("")
	if err != nil {
		return nil, err
	}
	var outsnaps []Snapshot

	// Convert to required format
	for _, entry := range snapListResp {
		for _, snapInfo := range entry.SnapList {
			outsnap := Snapshot{
				Name:       snapInfo.VolInfo.Name,
				VolumeName: entry.ParentName,
			}
			if snapInfo.VolInfo.State == api.VolStarted {
				outsnap.Started = true
			}
			outsnaps = append(outsnaps, outsnap)
		}
	}
	return outsnaps, nil
}
