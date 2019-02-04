package conf

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/gluster/gluster-prometheus/pkg/glusterutils/glusterconsts"
)

// GConfig represents Glusterd1/Glusterd2 configurations
type GConfig struct {
	GlusterMgmt       string `toml:"gluster-mgmt"`
	Glusterd2Endpoint string `toml:"gd2-rest-endpoint"`
	GlusterCmd        string `toml:"gluster-binary-path"`
	GlusterdWorkdir   string `toml:"glusterd-dir"`
	GlusterClusterID  string `toml:"gluster-cluster-id"`
	Glusterd2User     string
	Glusterd2Secret   string
	Glusterd2Cacert   string
	Glusterd2Insecure bool
	Timeout           int64
}

// Globals maintains the global system configurations
type Globals struct {
	Port              int      `toml:"port"`
	MetricsPath       string   `toml:"metrics-path"`
	LogDir            string   `toml:"log-dir"`
	LogFile           string   `toml:"log-file"`
	LogLevel          string   `toml:"log-level"`
	CacheTTL          uint64   `toml:"cache-ttl-in-sec"`
	CacheEnabledFuncs []string `toml:"cache-enabled-funcs"`
	*GConfig
}

// Collectors struct defines the structure of collectors configuration
type Collectors struct {
	Name         string `toml:"name"`
	SyncInterval uint64 `toml:"sync-interval"`
	Disabled     bool   `toml:"disabled"`
}

// Config struct defines overall configurations
// it embeds 'Globals' configuration
type Config struct {
	*Globals       `toml:"globals"`
	CollectorsConf map[string]Collectors `toml:"collectors"`
}

// GConfig method helps 'Config' objects to implement 'GConfigInterface'
func (conf *Config) GConfig() *GConfig {
	return conf.Globals.GConfig
}

// LoadConfig loads the configuration file
func LoadConfig(confFilePath string) (conf *Config, err error) {
	conf = &Config{}
	if _, err = toml.DecodeFile(filepath.Clean(confFilePath), conf); err != nil {
		conf = nil
		return
	}
	// by default, use glusterd (that is; GD1)
	if conf.GlusterMgmt == "" {
		conf.GlusterMgmt = glusterconsts.MgmtGlusterd
	}
	// If GD2_ENDPOINTS env variable is set, use that info
	// for making REST API calls
	if endpoint := os.Getenv(glusterconsts.EnvGD2Endpoints); endpoint != "" {
		conf.Glusterd2Endpoint = endpoint
	}
	// if there are multiple endpoints, get the first one
	if endpoint := conf.Glusterd2Endpoint; endpoint != "" {
		endpoint = strings.Replace(endpoint, ",", " ", -1)
		endpoint = strings.Fields(endpoint)[0]
		conf.Glusterd2Endpoint = endpoint
	}
	// if GLUSTER_CLUSTER_ID env variable is set, it gets the precedence
	if gClusterID := os.Getenv(glusterconsts.EnvGlusterClusterID); gClusterID != "" {
		conf.GlusterClusterID = gClusterID
	}
	// gluster cluster ID is still empty, put the default
	if conf.GlusterClusterID == "" {
		conf.GlusterClusterID = glusterconsts.DefaultGlusterClusterID
	}
	return
}

// GConfigInterface enables to get configuration,
// with which the gluster management objects are created.
// Should be implemented by both GD1 and GD2.
type GConfigInterface interface {
	GConfig() *GConfig
}

// GConfigFromInterface method returns a 'GConfig' pointer,
// if and only if the argument interface implements 'GConfigInterface'
func GConfigFromInterface(iFace interface{}) (*GConfig, error) {
	if gConfig, ok := iFace.(GConfigInterface); ok {
		return gConfig.GConfig(), nil
	}
	return nil, errors.New("provided interface don't implement 'GConfigInterface'")
}
