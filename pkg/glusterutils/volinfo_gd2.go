package glusterutils

import (
	"github.com/gluster/glusterd2/pkg/api"
)

// temporary hack till glusterd2 supports SubvolType.String()
var subvolTypeValueToName = map[api.SubvolType]string{
	api.SubvolDistribute: "Distribute",
	api.SubvolReplicate:  "Replicate",
	api.SubvolDisperse:   "Disperse",
}

func gd2VolumeInfo(config *Config) ([]Volume, error) {
	client, err := initRESTClient(config)
	if err != nil {
		return nil, err
	}
	vols, err := client.Volumes("")
	if err != nil {
		return nil, err
	}
	volumes := make([]Volume, len(vols))

	// Convert to required format
	for vidx, vol := range vols {
		volumes[vidx] = Volume{
			DistributeCount:         vol.DistCount,
			ID:                      vol.ID.String(),
			Metadata:                vol.Metadata,
			Name:                    vol.Name,
			Options:                 vol.Options,
			ReplicaCount:            vol.ReplicaCount,
			SnapList:                vol.SnapList,
			State:                   vol.State.String(),
			Transport:               vol.Transport,
			SubVolumes:              make([]SubVolume, len(vol.Subvols)),
			Type:                    vol.Type.String(),
			DisperseCount:           vol.DisperseCount,
			DisperseDataCount:       vol.DisperseDataCount,
			DisperseRedundancyCount: vol.DisperseRedundancyCount,
		}
		for sidx, sv := range vol.Subvols {
			volumes[vidx].SubVolumes[sidx] = SubVolume{
				ArbiterCount:            sv.ArbiterCount,
				DisperseCount:           sv.DisperseCount,
				DisperseDataCount:       sv.DisperseDataCount,
				DisperseRedundancyCount: sv.DisperseRedundancyCount,
				ReplicaCount:            sv.ReplicaCount,
				Name:                    sv.Name,
				Type:                    subvolTypeValueToName[sv.Type],
				Bricks:                  make([]Brick, len(sv.Bricks)),
			}
			for bidx, brick := range sv.Bricks {
				volumes[vidx].SubVolumes[sidx].Bricks[bidx] = Brick{
					Host:       brick.Hostname,
					ID:         brick.ID.String(),
					Path:       brick.Path,
					PeerID:     brick.PeerID.String(),
					Type:       brick.Type.String(),
					VolumeID:   brick.VolumeID.String(),
					VolumeName: brick.VolumeName,
				}
			}
		}
	}

	return volumes, nil
}
