package glusterutils

import (
	"os/exec"

	log "github.com/sirupsen/logrus"
)

// EnableVolumeProfiling enables profiling for a volume
func (g *GD1) EnableVolumeProfiling(volume Volume) error {
	value, exists := volume.Options[CountFOPHitsGD1]
	if !exists {
		// Enable profiling for the volumes as its not set
		_, err := exec.Command(g.config.GlusterCmd, "volume", "profile", volume.Name, "start").Output()
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
