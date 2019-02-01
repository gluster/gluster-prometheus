package glusterutils

import (
	"github.com/gluster/gluster-prometheus/pkg/glusterutils/glusterconsts"
	log "github.com/sirupsen/logrus"
)

// EnableVolumeProfiling enables profiling for a volume
func (g *GD1) EnableVolumeProfiling(volume Volume) error {
	value, exists := volume.Options[glusterconsts.CountFOPHitsGD1]
	if !exists {
		// Enable profiling for the volumes as its not set
		_, err := g.execGluster("volume", "profile", volume.Name, "start")
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
