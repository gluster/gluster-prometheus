package glusterutils

import (
	"encoding/xml"
)

// Snapshots returns snaphosts list for the cluster
func (g *GD1) Snapshots() ([]Snapshot, error) {
	// Run Gluster snapshot list
	out, err := g.execGluster("snapshot", "info")
	if err != nil {
		return nil, err
	}

	var snaps gd1Snapshots
	err = xml.Unmarshal(out, &snaps)
	if err != nil {
		return nil, err
	}

	outsnaps := make([]Snapshot, len(snaps.List))
	for idx, snap := range snaps.List {
		outsnap := Snapshot{
			Name:       snap.Name,
			VolumeName: snap.SnapVolume.OriginVolume.Name,
		}
		if snap.SnapVolume.Status == "Started" {
			outsnap.Started = true
		}
		outsnaps[idx] = outsnap
	}
	return outsnaps, nil
}
