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
