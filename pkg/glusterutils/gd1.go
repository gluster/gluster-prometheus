package glusterutils

import (
	"encoding/xml"
)

type healBricks struct {
	XMLName     xml.Name         `xml:"cliOutput"`
	Healentries []healEntriesXML `xml:"healInfo>bricks>brick"`
}

type healEntriesXML struct {
	XMLName        xml.Name `xml:"brick"`
	HostUUID       string   `xml:"hostUuid,attr"`
	Brickname      string   `xml:"name"`
	Connected      string   `xml:"status"`
	NumHealEntries string   `xml:"numberOfEntries"`
}

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

type snapshotParentVolume struct {
	Name          string `xml:"name"`
	SnapCount     int    `xml:"snapCount"`
	SnapRemaining int    `xml:"snapRemaining"`
}

type snapshotVolume struct {
	Name         string               `xml:"name"`
	Status       string               `xml:"status"`
	OriginVolume snapshotParentVolume `xml:"originVolume"`
}

type gd1Snapshot struct {
	Name        string         `xml:"name"`
	UUID        string         `xml:"uuid"`
	Description string         `xml:"description"`
	CreateTime  string         `xml:"createTime"`
	VolCount    string         `xml:"volCount"`
	SnapVolume  snapshotVolume `xml:"snapVolume"`
}

type gd1Snapshots struct {
	XMLName xml.Name      `xml:"cliOutput"`
	List    []gd1Snapshot `xml:"snapInfo>snapshots>snapshot"`
}

type blockStat struct {
	Size   uint64 `xml:"size"`
	Reads  uint64 `xml:"reads"`
	Writes uint64 `xml:"writes"`
}

type gd1FopStat struct {
	Name       string  `xml:"name"`
	Hits       int     `xml:"hits"`
	AvgLatency float64 `xml:"avgLatency"`
	MinLatency float64 `xml:"minLatency"`
	MaxLatency float64 `xml:"maxLatency"`
}

type cumulativeStats struct {
	BlkStats   []blockStat  `xml:"blokcStats>block"`
	Duration   uint64       `xml:"duration"`
	TotalRead  uint64       `xml:"totalRead"`
	TotalWrite uint64       `xml:"totalWrite"`
	FopStats   []gd1FopStat `xml:"fopStats>fop"`
}

type brickProfileInfo struct {
	Name  string          `xml:"brickName"`
	Stats cumulativeStats `xml:"cumulativeStats"`
}

type volumeProfile struct {
	VolName    string             `xml:"volname"`
	ProfileOp  int                `xml:"profileOp"`
	BrickCount int                `xml:"brickCount"`
	Bricks     []brickProfileInfo `xml:"brick"`
}

type gd1ProfileInfo struct {
	XMLName    xml.Name      `xml:"cliOutput"`
	VolProfile volumeProfile `xml:"volProfile"`
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
