package glusterutils

const (
	// VolumeTypeDistReplicate represents Gluster distributed replicate volume
	VolumeTypeDistReplicate = "Distributed Replicate"
	// VolumeTypeDistDisperse represents Gluster distributed disperse volume
	VolumeTypeDistDisperse = "Distributed Disperse"
	// VolumeTypeReplicate represents Gluster replicate volume
	VolumeTypeReplicate = "Replicate"
	// VolumeTypeDisperse represents Gluster disperse volume
	VolumeTypeDisperse = "Disperse"
	// VolumeTypeDistribute represents Gluster distribute volume
	VolumeTypeDistribute = "Distribute"
	// SubvolTypeDistribute represents Gluster distribute sub volume
	SubvolTypeDistribute = VolumeTypeDistribute
	// SubvolTypeReplicate represents Gluster replicate sub volume
	SubvolTypeReplicate = VolumeTypeReplicate
	// SubvolTypeDisperse represents Gluster disperse sub volume
	SubvolTypeDisperse = VolumeTypeDisperse
	// BrickTypeDefault represents default brick type
	BrickTypeDefault = "Brick"
	// BrickTypeArbiter represents arbiter brick type
	BrickTypeArbiter = "Arbiter"

	// MgmtGlusterd represents glusterd
	MgmtGlusterd = "glusterd"
	// MgmtGlusterd2 represents glusterd
	MgmtGlusterd2 = "glusterd2"

	// VolumeStateCreated represents Volume Created state
	VolumeStateCreated = "Created"
	// VolumeStateStarted represents Volume started state
	VolumeStateStarted = "Started"
	// VolumeStateStopped represents Volume stopped state
	VolumeStateStopped = "Stopped"
)

// Config represents Glusterd1/Glusterd2 configurations
type Config struct {
	GlusterMgmt       string
	Glusterd2Endpoint string
	GlusterCmd        string
	GlusterdWorkdir   string
	Glusterd2User     string
	Glusterd2Secret   string
	Glusterd2Cacert   string
	Glusterd2Insecure bool
	Timeout           int64
}

// Peer represents a Gluster Peer
type Peer struct {
	ID            string   `json:"id"`
	PeerAddresses []string `json:"peer-addresses"`
	Online        bool     `json:"online"`
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
	Hostname string
	PeerID   string
	Status   int
	PID      int
	Path     string
	Volume   string
}

// GInterface should be implemented in GD1 and GD2 structs
type GInterface interface {
	Peers() ([]Peer, error)
	LocalPeerID() (string, error)
	IsLeader() (bool, error)
	HealInfo(vol string) ([]HealEntry, error)
	VolumeInfo() ([]Volume, error)
	Snapshots() ([]Snapshot, error)
	VolumeProfileInfo(vol string) ([]ProfileInfo, error)
	VolumeBrickStatus(vol string) ([]BrickStatus, error)
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
	BrickName   string
	Duration    uint64
	TotalReads  uint64
	TotalWrites uint64
	FopStats    []FopStat
}

// GD1 enables users to interact with gd1 version
type GD1 struct {
	config *Config
}

// GD2 is struct to interact with Glusterd2 using REST API
type GD2 struct {
	config *Config
}
