package glusterutils

import "strings"

// HealInfo gets heal info from glusterd2 using rest api
func (g GD2) HealInfo(vol string) ([]HealEntry, error) {
	client, err := initRESTClient(g.config)
	if err != nil {
		return nil, err
	}
	healinfo, herr := client.SelfHealInfo(vol)
	if herr != nil {
		return nil, herr
	}
	brickheal := make([]HealEntry, len(healinfo))
	for hidx, heal := range healinfo {
		hostPath := strings.Split(heal.Name, ":")
		entry := HealEntry{PeerID: heal.HostID, Hostname: hostPath[0],
			Brick: hostPath[1], Connected: heal.Status,
			NumHealEntries: int(*(heal.Entries))}
		brickheal[hidx] = entry
	}
	return brickheal, herr

}
