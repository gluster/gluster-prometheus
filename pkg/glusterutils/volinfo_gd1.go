package glusterutils

import (
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/gluster/gluster-prometheus/pkg/glusterutils/glusterconsts"
)

// VolumeInfo returns gluster vol info (glusterd)
func (g *GD1) VolumeInfo() ([]Volume, error) {
	// Run Gluster volume info
	out, err := g.execGluster("volume", "info")
	if err != nil {
		return nil, err
	}

	var vols gd1Volumes
	err = xml.Unmarshal(out, &vols)
	if err != nil {
		return nil, err
	}

	outvols := make([]Volume, len(vols.List))
	for vidx, vol := range vols.List {
		outvol := Volume{
			ID:                      vol.ID,
			Name:                    vol.Name,
			Transport:               vol.TransportRaw.String(),
			State:                   vol.Status,
			Type:                    vol.Type,
			DistributeCount:         vol.DistCount,
			ReplicaCount:            vol.ReplicaCount,
			DisperseCount:           vol.DisperseCount,
			DisperseDataCount:       vol.DisperseCount - vol.DisperseRedundancyCount,
			DisperseRedundancyCount: vol.DisperseRedundancyCount,
		}
		outvol.Options = make(map[string]string)
		for _, opt := range vol.Options {
			outvol.Options[opt.Name] = opt.Value
		}

		subvolType := getSubvolType(vol.Type)
		subvolBricksCount := getSubvolBricksCount(vol.ReplicaCount, vol.DisperseCount)
		numberOfSubvols := len(vol.Bricks) / subvolBricksCount
		outvol.SubVolumes = make([]SubVolume, numberOfSubvols)
		for sidx := 0; sidx < numberOfSubvols; sidx++ {
			outvol.SubVolumes[sidx].Type = subvolType
			outvol.SubVolumes[sidx].ReplicaCount = vol.ReplicaCount
			outvol.SubVolumes[sidx].DisperseCount = vol.DisperseCount
			outvol.SubVolumes[sidx].DisperseDataCount = vol.DisperseCount - vol.DisperseRedundancyCount
			outvol.SubVolumes[sidx].DisperseRedundancyCount = vol.DisperseRedundancyCount
			outvol.SubVolumes[sidx].Name = fmt.Sprintf("%s-%s-%d", vol.Name, strings.ToLower(subvolType), sidx)
			for bidx := 0; bidx < subvolBricksCount; bidx++ {
				brickType := glusterconsts.BrickTypeDefault
				if vol.Bricks[sidx+bidx].IsArbiter == 1 {
					brickType = glusterconsts.BrickTypeArbiter
				}
				brickParts := strings.Split(vol.Bricks[sidx+bidx].Name, ":")
				brick := Brick{
					Host:       brickParts[0],
					PeerID:     vol.Bricks[sidx+bidx].PeerID,
					Type:       brickType,
					Path:       brickParts[1],
					VolumeID:   vol.ID,
					VolumeName: vol.Name,
				}
				outvol.SubVolumes[sidx].Bricks = append(outvol.SubVolumes[sidx].Bricks, brick)
			}
		}
		outvols[vidx] = outvol
	}

	return outvols, nil
}
