package glusterutils

import (
	"encoding/xml"
	"fmt"
	"os/exec"
	"strings"
)

type gd1Brick struct {
	Name      string `xml:"name"`
	PeerID    string `xml:"hostUuid"`
	IsArbiter int    `xml:"IsArbiter"`
}

type gd1Option struct {
	Name  string `xml:"name"`
	Value string `xml:"value"`
}

type gd1Transport string

type gd1Volume struct {
	Name                    string       `xml:"name"`
	ID                      string       `xml:"id"`
	Status                  string       `xml:"statusStr"`
	Type                    string       `xml:"typeStr"`
	Bricks                  []gd1Brick   `xml:"bricks>brick"`
	NumBricks               int          `xml:"brickCount"`
	DistCount               int          `xml:"distCount"`
	ReplicaCount            int          `xml:"replicaCount"`
	DisperseCount           int          `xml:"disperseCount"`
	DisperseRedundancyCount int          `xml:"redundancyCount"`
	StripeCount             int          `xml:"stripeCount"`
	TransportRaw            gd1Transport `xml:"transport"`
	Options                 []gd1Option  `xml:"options>option"`
}

type gd1Volumes struct {
	XMLName xml.Name    `xml:"cliOutput"`
	List    []gd1Volume `xml:"volInfo>volumes>volume"`
}

func (t *gd1Transport) String() string {
	// 0 - tcp
	// 1 - rdma
	// 2 - tcp,rdma
	if *t == "0" {
		return "tcp"
	} else if *t == "1" {
		return "rdma"
	}
	return "tcp,rdma"
}

func getSubvolType(voltype string) string {
	switch voltype {
	case VolumeTypeDistReplicate:
		return SubvolTypeReplicate
	case VolumeTypeDistDisperse:
		return SubvolTypeDisperse
	default:
		return voltype
	}
}

func getSubvolBricksCount(replicaCount int, disperseCount int) int {
	if replicaCount > 0 {
		return replicaCount
	}

	if disperseCount > 0 {
		return disperseCount
	}
	return 1
}

func gd1VolumeInfo(config *Config) ([]Volume, error) {
	// Run Gluster volume info --xml
	out, err := exec.Command(config.GlusterCmd, "volume", "info", "--xml").Output()
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
				brickType := BrickTypeDefault
				if vol.Bricks[sidx+bidx].IsArbiter == 1 {
					brickType = BrickTypeArbiter
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
