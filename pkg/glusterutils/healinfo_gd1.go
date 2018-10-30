package glusterutils

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
)

// HealInfo gets gluster vol heal info (GD1)
func (g GD1) HealInfo(vol string) ([]HealEntry, error) {
	cmd := fmt.Sprintf("%s vol heal %s info --xml", g.config.GlusterCmd, vol)
	out, err := executeCmd(cmd)
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
			entries, err := strconv.Atoi(entry.NumHealEntries)
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
