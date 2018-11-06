package glusterutils

import (
	"encoding/xml"
	"os/exec"
)

// Snapshots returns snaphosts list for the cluster
func (g *GD1) Snapshots() ([]Snapshot, error) {
	// Run Gluster snapshot list --xml
	out, err := exec.Command(g.config.GlusterCmd, "snapshot", "info", "--xml").Output()
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
