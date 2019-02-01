package glusterutils

import (
	"encoding/xml"
)

// VolumeProfileInfo returns profile info details for the volume
func (g *GD1) VolumeProfileInfo(vol string) ([]ProfileInfo, error) {
	// Run Gluster volume profile <volname> info
	out, err := g.execGluster("volume", "profile", vol, "info")
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
			BrickName:      brick.Name,
			Duration:       brick.Stats.Duration,
			TotalReads:     brick.Stats.TotalRead,
			TotalWrites:    brick.Stats.TotalWrite,
			DurationInt:    brick.IntStats.Duration,
			TotalReadsInt:  brick.IntStats.TotalRead,
			TotalWritesInt: brick.IntStats.TotalWrite,
		}

		if brick.Stats.FopStats != nil {
			fopStats := make([]FopStat, len(brick.Stats.FopStats))
			for idx1, stat := range brick.Stats.FopStats {
				fopStats[idx1] = FopStat(stat)
			}
			obj.FopStats = fopStats
		}
		if brick.IntStats.FopStats != nil {
			intFopStats := make([]FopStat, len(brick.IntStats.FopStats))
			for idx1, stat := range brick.IntStats.FopStats {
				intFopStats[idx1] = FopStat(stat)
			}
			obj.FopStatsInt = intFopStats
		}
		profileinfo[idx] = obj
	}
	return profileinfo, nil
}
