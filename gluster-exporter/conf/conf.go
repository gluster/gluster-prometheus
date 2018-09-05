package conf

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// Collectors maintain the global collectors configurations
var Collectors = make(map[string]CollectorConfig)

// SystemConfig maintains the global system configurations
var SystemConfig = make(map[string]string)

// CollectorConfig struct defines the structure of collectors configuration
type CollectorConfig struct {
	Name         string `json:"name"`
	SyncInterval uint64 `json:"sync_interval"`
	Disabled     bool   `json:"disabled"`
}

// LoadCollectorsConfig loads the collectors configuration
func LoadCollectorsConfig(confFilePath string) string {
	fileObj, err := os.Open(confFilePath)
	if err != nil {
		msg := fmt.Sprintf("Failed loading config file: %s", confFilePath)
		return msg
	}
	defer fileObj.Close()
	data, _ := ioutil.ReadAll(fileObj)
	var confs []CollectorConfig
	if err := json.Unmarshal(data, &confs); err != nil {
		msg := fmt.Sprintf("Error parsing the content of config file: %s", confFilePath)
		return msg
	}
	for _, conf := range confs {
		Collectors[conf.Name] = conf
	}
	return ""
}

// LoadConfig loads the global system configuration
func LoadConfig(confFilePath string) string {
	fileObj, err := os.Open(confFilePath)
	if err != nil {
		msg := fmt.Sprintf("Failed loading config file: %s", confFilePath)
		return msg
	}
	defer fileObj.Close()
	data, _ := ioutil.ReadAll(fileObj)
	if err := json.Unmarshal(data, &SystemConfig); err != nil {
		msg := fmt.Sprintf("Error parsing the content of config file: %s", confFilePath)
		return msg
	}
	return ""
}
