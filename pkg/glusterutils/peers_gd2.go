package glusterutils

import (
	"github.com/gluster/glusterd2/pkg/api"
)

var (
	peers api.PeerListResp
)

// Peers returns the list of peers ( for GlusterD2 )
func (g *GD2) Peers() ([]Peer, error) {
	var peersgd2 []Peer
	client, err := initRESTClient(g.config)
	if err != nil {
		return nil, err
	}
	peers, err = client.Peers()
	if err != nil {
		return peersgd2, err
	}
	peersgd2 = make([]Peer, len(peers))

	// Convert to required format
	for pidx, peergd2 := range peers {
		peersgd2[pidx] = Peer{
			ID:            peergd2.ID.String(),
			PeerAddresses: peergd2.PeerAddresses,
			Online:        peergd2.Online,
		}
	}
	return peersgd2, nil
}
