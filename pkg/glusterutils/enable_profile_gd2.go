package glusterutils

import (
	"github.com/gluster/glusterd2/pkg/api"
	log "github.com/sirupsen/logrus"
)

// EnableVolumeProfiling enables profiling for a volume
func (g *GD2) EnableVolumeProfiling(volume Volume) error {
	client, err := initRESTClient(g.config)
	if err != nil {
		return err
	}

	value, exists := volume.Options[CountFOPHitsGD2]
	if !exists {
		// Enable profiling for the volumes as its not set
		err := client.VolumeSet(
			volume.Name,
			api.VolOptionReq{
				Options: map[string]string{
					CountFOPHitsGD2:       "on",
					LatencyMeasurementGD2: "on",
				},
				VolOptionFlags: api.VolOptionFlags{
					AllowAdvanced: true,
				},
			},
		)
		if err != nil {
			return err
		}
	} else {
		if value == "off" {
			log.WithFields(log.Fields{
				"volume": volume.Name,
			}).Debug("Volume profiling is explicitly disabled. No profile metrics would be exposed.")
		}
	}
	return nil
}
