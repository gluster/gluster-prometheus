package glusterutils

import (
	"github.com/gluster/gluster-prometheus/gluster-exporter/conf"
)

// Peer represents a Gluster Peer
type Peer struct {
	ID            string   `json:"id"`
	PeerAddresses []string `json:"peer-addresses"`
	Online        bool     `json:"online"`
	Gd1State      int      // GD1 only
}

// Brick represents Gluster Brick
type Brick struct {
	Host       string `json:"host"`
	ID         string `json:"id"`
	Path       string `json:"path"`
	PeerID     string `json:"peer-id"`
	Type       string `json:"type"`
	VolumeID   string `json:"volume-id"`
	VolumeName string `json:"volume-name"`
}

// SubVolume represents Gluster SubVolume
type SubVolume struct {
	ArbiterCount            int     `json:"arbiter-count"`
	Bricks                  []Brick `json:"bricks"`
	DisperseCount           int     `json:"disperse-count"`
	DisperseDataCount       int     `json:"disperse-data-count"`
	DisperseRedundancyCount int     `json:"disperse-redundancy-count"`
	Name                    string  `json:"name"`
	ReplicaCount            int     `json:"replica-count"`
	Type                    string  `json:"type"`
}

// Volume represents Gluster Volume
type Volume struct {
	DistributeCount         int               `json:"distribute-count"`
	ID                      string            `json:"id"`
	Metadata                map[string]string `json:"metadata"`
	Name                    string            `json:"name"`
	Options                 map[string]string `json:"options"`
	SnapList                []string          `json:"snap-list"`
	State                   string            `json:"state"`
	SubVolumes              []SubVolume       `json:"subvols"`
	Transport               string            `json:"transport"`
	Type                    string            `json:"type"`
	DisperseCount           int               `json:"disperse-count"`
	DisperseDataCount       int               `json:"disperse-data-count"`
	DisperseRedundancyCount int               `json:"disperse-redundancy-count"`
	ReplicaCount            int               `json:"replica-count"`
}

// VolumeStatus represents the detailed status of a Gluster volume
type VolumeStatus struct {
	Name  string
	Nodes []BrickStatus
}

// HealEntry describe gluster heal info for each brick
type HealEntry struct {
	PeerID         string
	Hostname       string
	Brick          string
	Connected      string
	NumHealEntries int64
}

// Snapshot represents a Volume snapshot
type Snapshot struct {
	Name       string
	VolumeName string
	Started    bool
}

// BrickStatus describes the status details of volume brick
type BrickStatus struct {
	Hostname       string
	PeerID         string
	Status         int
	PID            int
	Port           int
	Path           string
	Volume         string
	Capacity       uint64
	Free           uint64
	Gd1InodesFree  int64 // only valid with GD1, -1 with GD2
	Gd1InodesTotal int64 // only valid with GD1, -1 with GD2
}

// GInterface should be implemented in GD1 and GD2 structs
type GInterface interface {
	Peers() ([]Peer, error)
	LocalPeerID() (string, error)
	IsLeader() (bool, error)
	HealInfo(vol string) ([]HealEntry, error)
	SplitBrainHealInfo(vol string) ([]HealEntry, error)
	VolumeInfo() ([]Volume, error)
	Snapshots() ([]Snapshot, error)
	VolumeProfileInfo(vol string) ([]ProfileInfo, error)
	VolumeBrickStatus(vol string) ([]BrickStatus, error)
	EnableVolumeProfiling(volinfo Volume) error
	VolumeStatus() ([]VolumeStatus, error)
}

// GDConfigInterface returns the configuration of the GD
type GDConfigInterface interface {
	Config() *Config
}

// FopStat defines file ops related details
type FopStat struct {
	Name       string
	Hits       int
	AvgLatency float64
	MinLatency float64
	MaxLatency float64
}

// ProfileInfo represents volume profile info brickwise
type ProfileInfo struct {
	BrickName      string
	Duration       uint64
	TotalReads     uint64
	TotalWrites    uint64
	FopStats       []FopStat
	DurationInt    uint64
	TotalReadsInt  uint64
	TotalWritesInt uint64
	FopStatsInt    []FopStat
}

// GD1 enables users to interact with gd1 version
type GD1 struct {
	config *conf.GConfig
}

// GD2 is struct to interact with Glusterd2 using REST API
type GD2 struct {
	config *conf.GConfig
}
