package glusterutils

import (
	"bufio"
	"errors"
	"os"
	"regexp"
	"strings"
)

var (
	peerIDPattern = regexp.MustCompile("[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}")
)

// VolumeInfo gets Volume info from Gd1/Gd2 based on config provided
func VolumeInfo(config *Config) ([]Volume, error) {
	setDefaultConfig(config)

	if config.GlusterMgmt == "glusterd2" {
		return gd2VolumeInfo(config)
	}

	return gd1VolumeInfo(config)
}

// LocalPeerID returns local peer ID of glusterd/glusterd2
func LocalPeerID(config *Config) (string, error) {
	setDefaultConfig(config)

	keywordID := "UUID"
	filepath := config.GlusterdWorkdir + "/glusterd.info"
	if config.GlusterMgmt == "glusterd2" {
		keywordID = "peer-id"
		filepath = config.GlusterdWorkdir + "/uuid.toml"
	}
	fileStream, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer fileStream.Close()

	scanner := bufio.NewScanner(fileStream)
	for scanner.Scan() {
		lines := strings.Split(scanner.Text(), "\n")
		for _, line := range lines {
			if strings.Contains(line, keywordID) {
				parts := strings.Split(string(line), "=")
				unformattedPeerID := parts[1]
				peerID := peerIDPattern.FindString(unformattedPeerID)
				if peerID == "" {
					return "", errors.New("unable to find peer address")
				}
				return peerID, nil
			}
		}
	}
	return "", errors.New("unable to find peer address")
}
