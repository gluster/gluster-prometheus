package conf

import (
	"io/ioutil"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Globals maintains the global system configurations
type Globals struct {
	GlusterMgmt       string   `toml:"gluster-mgmt"`
	GlusterdDir       string   `toml:"glusterd-dir"`
	GlusterBinaryPath string   `toml:"gluster-binary-path"`
	GD2RESTEndpoint   string   `toml:"gd2-rest-endpoint"`
	Port              int      `toml:"port"`
	MetricsPath       string   `toml:"metrics-path"`
	LogDir            string   `toml:"log-dir"`
	LogFile           string   `toml:"log-file"`
	LogLevel          string   `toml:"log-level"`
	CacheTTL          uint64   `toml:"cache-ttl-in-sec"`
	CacheEnabledFuncs []string `toml:"cache-enabled-funcs"`
}

// Collectors struct defines the structure of collectors configuration
type Collectors struct {
	Name         string `toml:"name"`
	SyncInterval uint64 `toml:"sync-interval"`
	Disabled     bool   `toml:"disabled"`
}

// Config struct defines overall configurations
type Config struct {
	GlobalConf     Globals               `toml:"globals"`
	CollectorsConf map[string]Collectors `toml:"collectors"`
}

// LoadConfig loads the configuration file
func LoadConfig(confFilePath string) (*Config, error) {
	var conf Config
	b, err := ioutil.ReadFile(filepath.Clean(confFilePath))
	if err != nil {
		return nil, err
	}
	if _, err := toml.Decode(string(b), &conf); err != nil {
		return nil, err
	}
	return &conf, nil
}
