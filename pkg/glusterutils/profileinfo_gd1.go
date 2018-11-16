package glusterutils

import (
	"encoding/xml"
	"os/exec"
)

// VolumeProfileInfo returns profile info details for the volume
func (g *GD1) VolumeProfileInfo(vol string) ([]ProfileInfo, error) {
	// Run Gluster volume profile <volname> info umulative --xml
	out, err := exec.Command(g.config.GlusterCmd, "volume", "profile", vol, "info", "cumulative", "--xml").Output()
	if err != nil {
		return nil, err
	}

	var info gd1ProfileInfo
	err = xml.Unmarshal(out, &info)
	if err != nil {
		return nil, err
	}

	profileinfo := make([]ProfileInfo, len(info.VolProfile.Bricks))
	for idx, brick := range info.VolProfile.Bricks {
		obj := ProfileInfo{
			BrickName:   brick.Name,
			Duration:    brick.Stats.Duration,
			TotalReads:  brick.Stats.TotalRead,
			TotalWrites: brick.Stats.TotalWrite,
		}
		if brick.Stats.FopStats != nil {
			fopStats := make([]FopStat, len(brick.Stats.FopStats))
			for idx1, stat := range brick.Stats.FopStats {
				fopStats[idx1] = FopStat(stat)
			}
			obj.FopStats = fopStats
		}
		profileinfo[idx] = obj
	}
	return profileinfo, nil
}
