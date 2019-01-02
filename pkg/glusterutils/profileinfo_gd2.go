package glusterutils

import (
	"strconv"
)

// VolumeProfileInfo returns profile info details for the volume
func (g *GD2) VolumeProfileInfo(vol string) ([]ProfileInfo, error) {
	client, err := initRESTClient(g.config)
	if err != nil {
		return nil, err
	}
	details, err := client.VolumeProfileInfo(vol, "info-cumulative")
	if err != nil {
		return nil, err
	}
	profileinfo := make([]ProfileInfo, len(details))
	for idx, info := range details {
		var duration, reads, writes int64
		if duration, err = strconv.ParseInt(info.CumulativeStats.Duration, 10, 64); err != nil {
			duration = 0
		}
		if reads, err = strconv.ParseInt(info.CumulativeStats.DataRead, 10, 64); err != nil {
			reads = 0
		}
		if writes, err = strconv.ParseInt(info.CumulativeStats.DataWrite, 10, 64); err != nil {
			writes = 0
		}
		obj := ProfileInfo{
			BrickName:   info.BrickName,
			Duration:    uint64(duration),
			TotalReads:  uint64(reads),
			TotalWrites: uint64(writes),
		}
		if info.CumulativeStats.StatsInfo != nil {
			// length field, in 'make' method, is initialized to ZERO
			// so that the append operation adds data from the start
			fopStats := make([]FopStat, 0, len(info.CumulativeStats.StatsInfo))
			for fopName, stat := range info.CumulativeStats.StatsInfo {
				var hits int
				var avgLatency, minLatency, maxLatency float64
				if hits, err = strconv.Atoi(stat["hits"]); err != nil {
					hits = 0
				}
				if avgLatency, err = strconv.ParseFloat(stat["avglatency"], 10); err != nil {
					avgLatency = 0.0
				}
				if minLatency, err = strconv.ParseFloat(stat["minlatency"], 10); err != nil {
					minLatency = 0.0
				}
				if maxLatency, err = strconv.ParseFloat(stat["maxlatency"], 10); err != nil {
					maxLatency = 0.0
				}

				fopStat := FopStat{
					Name:       fopName,
					Hits:       hits,
					AvgLatency: avgLatency,
					MinLatency: minLatency,
					MaxLatency: maxLatency,
				}
				fopStats = append(fopStats, fopStat)
			}
			obj.FopStats = fopStats
		}
		if info.IntervalStats.StatsInfo != nil {
			// length field, in 'make' method, is initialized to ZERO
			// so that the append operation adds data from the start
			fopStats := make([]FopStat, 0, len(info.IntervalStats.StatsInfo))
			for fopName, stat := range info.IntervalStats.StatsInfo {
				var hits int
				var avgLatency, minLatency, maxLatency float64
				if hits, err = strconv.Atoi(stat["hits"]); err != nil {
					hits = 0
				}
				if avgLatency, err = strconv.ParseFloat(stat["avglatency"], 10); err != nil {
					avgLatency = 0.0
				}
				if minLatency, err = strconv.ParseFloat(stat["minlatency"], 10); err != nil {
					minLatency = 0.0
				}
				if maxLatency, err = strconv.ParseFloat(stat["maxlatency"], 10); err != nil {
					maxLatency = 0.0
				}

				fopStat := FopStat{
					Name:       fopName,
					Hits:       hits,
					AvgLatency: avgLatency,
					MinLatency: minLatency,
					MaxLatency: maxLatency,
				}
				fopStats = append(fopStats, fopStat)
			}
			obj.FopStatsInt = fopStats
		}
		profileinfo[idx] = obj
	}

	return profileinfo, nil
}
