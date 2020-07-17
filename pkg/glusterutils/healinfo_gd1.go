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
			entries, err := strconv.ParseInt(entry.NumHealEntries, 10, 64)
			if err != nil {
				return nil, err
			}
			hostPath := strings.Split(entry.Brickname, ":")
			heal := HealEntry{PeerID: entry.HostUUID, Hostname: hostPath[0],
				Brick:          hostPath[1],
				Connected:      entry.Connected,
				NumHealEntries: entries}
			heals[hidx] = heal
		}
	}

	return heals, nil
}

// HealInfo gets gluster vol heal info (GD1)
func (g GD1) HealInfo(vol string) ([]HealEntry, error) {
	// Get the overall heal count
	cmd := fmt.Sprintf("vol heal %s info --nolog", vol)
	heals, err := g.getHealDetails(cmd)
	if err != nil {
		return nil, err
	}

	return heals, nil
}

// SplitBrainHealInfo gets gluster vol heal info (GD1)
func (g GD1) SplitBrainHealInfo(vol string) ([]HealEntry, error) {
	cmd := fmt.Sprintf("vol heal %s info split-brain --nolog", vol)
	splitBrainHeals, err := g.getHealDetails(cmd)
	if err != nil {
		return nil, err
	}

	return splitBrainHeals, nil
}
