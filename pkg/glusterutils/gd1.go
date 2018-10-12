package glusterutils

import (
	"encoding/xml"
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
