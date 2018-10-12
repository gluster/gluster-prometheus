package glusterutils

import (
	"encoding/xml"
	"os/exec"
)

type peerGlusterd1 struct {
	PeerID    string   `xml:"uuid"`
	Hostname  []string `xml:"hostname"`
	Connected int      `xml:"connected"`
	State     int      `xml:"state"`
	StateStr  string   `xml:"stateStr"`
}

type peersGlusterd1 struct {
	XMLName xml.Name        `xml:"cliOutput"`
	List    []peerGlusterd1 `xml:"peerStatus>peer"`
}

// Peers returns the list of peers ( for GlusterD1 )
func (g *GD1) Peers() ([]Peer, error) {
	var gd1Peers peersGlusterd1
	var peersgd1 []Peer
	out, err := exec.Command(g.config.GlusterCmd, "pool", "list", "--xml").Output()
	if err != nil {
		return peersgd1, err
	}
	err = xml.Unmarshal(out, &gd1Peers)
	if err != nil {
		return peersgd1, err
	}
	peersgd1 = make([]Peer, len(gd1Peers.List))
	var online bool
	// Convert to required format
	for pidx, peergd1 := range gd1Peers.List {
		if peergd1.Connected == 1 {
			online = true
		} else {
			online = false
		}
		peersgd1[pidx] = Peer{
			ID:            peergd1.PeerID,
			PeerAddresses: peergd1.Hostname,
			Online:        online,
		}
	}

	return peersgd1, nil
}
