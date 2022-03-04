package glusterutils

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
)

func (g *GD1) getHealDetails(cmd string) ([]HealEntry, error) {
	args := strings.Fields(cmd)
	out, err := g.execGluster(args...)
	if err != nil {
		return nil, err
	}
	var healop healBricks
	err = xml.Unmarshal(out, &healop)
	if err != nil {
		return nil, err
	}
	heals := make([]HealEntry, len(healop.Healentries))
	for hidx, entry := range healop.Healentries {
		if entry.Connected == "Connected" {
			numHealPending, err := strconv.ParseInt(entry.NumHealPending, 10, 64)
			if err != nil {
				return nil, err
			}
			numSplitBrain, err := strconv.ParseInt(entry.NumSplitBrain, 10, 64)
			if err != nil {
				return nil, err
			}
			hostPath := strings.Split(entry.Brickname, ":")
			heal := HealEntry{PeerID: entry.HostUUID, Hostname: hostPath[0],
				Brick:          hostPath[1],
				Connected:      entry.Connected,
				NumHealPending: numHealPending,
				NumSplitBrain:  numSplitBrain}
			heals[hidx] = heal
		}
	}

	return heals, nil
}

// HealInfo gets gluster vol heal info (GD1)
func (g GD1) HealInfo(vol string) ([]HealEntry, error) {
	// Get the overall heal count
	cmd := fmt.Sprintf("vol heal %s info summary --nolog", vol)
	heals, err := g.getHealDetails(cmd)
	if err != nil {
		return nil, err
	}

	return heals, nil
}
