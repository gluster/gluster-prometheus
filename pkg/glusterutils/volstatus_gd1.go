package glusterutils

import (
	"encoding/xml"
	"strconv"
)

// VolumeStatus returns gluster vol status (glusterd)
func (g *GD1) VolumeStatus() ([]VolumeStatus, error) {
	// Run Gluster volume status all detail --xml --mode=script
	out, err := g.execGluster("volume", "status", "all", "detail")
	if err != nil {
		return nil, err
	}

	var vols gd1VolumesDetail
	err = xml.Unmarshal(out, &vols)
	if err != nil {
		return nil, err
	}

	outvols := make([]VolumeStatus, len(vols.List))
	for vidx, vol := range vols.List {
		outvol := VolumeStatus{
			Name: vol.Name,
		}
		outvol.Nodes = make([]BrickStatus, len(vol.Nodes))
		for nidx, node := range vol.Nodes {
			port64, err := strconv.ParseInt(node.Port, 10, 32)
			if err != nil {
				port64 = -1
			}
			port := int(port64)
			outnode := BrickStatus{
				Hostname:       node.Hostname,
				PeerID:         node.PeerID,
				Status:         node.Status,
				Port:           port,
				PID:            node.PID,
				Gd1InodesTotal: int64(node.InodesTotal),
				Gd1InodesFree:  int64(node.InodesFree),
				Capacity:       node.SizeTotal,
				Free:           node.SizeFree,
				Volume:         vol.Name,
				Path:           node.Path,
			}
			outvol.Nodes[nidx] = outnode
		}
		outvols[vidx] = outvol
	}

	return outvols, nil
}
