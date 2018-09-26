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

// Peers return a list of peers
func Peers(config *Config) ([]Peer, error) {
	setDefaultConfig(config)

	if config.GlusterMgmt == "glusterd2" {
		return peersGD2(config)
	}

	return peersGD1(config)
}

// IsLeader returns true or false based on whether the node is the leader of the cluster or not
func IsLeader(config *Config) (bool, error) {
	setDefaultConfig(config)
	peerList, err := Peers(config)
	if err != nil {
		return false, err
	}
	peerID, err := LocalPeerID(config)
	if err != nil {
		return false, err
	}
	if config.GlusterMgmt == "glusterd2" {
		//The following lines checks and returns true if the local PeerID is equal to that of the first PeerID in the list
		for _, pr := range peerList {
			if pr.Online == true {
				if peerID == peerList[0].ID {
					return true, nil
				}
				return false, nil
			}
		}
		// This would imply none of the peers are online and return false
		return false, nil
	}
	var maxPeerID string
	//This for loop iterates among all the peers and finds the peer with the maximum UUID (lexicographically)
	for i, pr := range peerList {
		if pr.Online == true {
			if peerList[i].ID > maxPeerID {
				maxPeerID = peerList[i].ID
			}
		}
	}
	//Checks and returns true if maximum peerID is equal to the local peerID
	if maxPeerID == peerID {
		return true, nil
	}
	return false, nil
}

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
