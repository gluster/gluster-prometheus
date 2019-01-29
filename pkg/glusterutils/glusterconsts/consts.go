package glusterconsts

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
	// EnvGD2Endpoints environment variable
	EnvGD2Endpoints = "GD2_ENDPOINTS"
	// EnvGlusterClusterID environment variable for cluster id
	EnvGlusterClusterID = "GLUSTER_CLUSTER_ID"

	// VolumeStateCreated represents Volume Created state
	VolumeStateCreated = "Created"
	// VolumeStateStarted represents Volume started state
	VolumeStateStarted = "Started"
	// VolumeStateStopped represents Volume stopped state
	VolumeStateStopped = "Stopped"

	// CountFOPHitsGD1 represents volume option name for fop hits counts
	CountFOPHitsGD1 = "diagnostics.count-fop-hits"
	// LatencyMeasurementGD1 represents volume option for latency measurement
	LatencyMeasurementGD1 = "diagnostics.latency-measurement"
	// CountFOPHitsGD2 represents volume option name for fop hits counts
	CountFOPHitsGD2 = "debug/io-stats.count-fop-hits"
	// LatencyMeasurementGD2 represents volume option for latency measurement
	LatencyMeasurementGD2 = "debug/io-stats.latency-measurement"

	// DefaultGlusterClusterID provides the default clusnter ID
	DefaultGlusterClusterID = "default"
)
