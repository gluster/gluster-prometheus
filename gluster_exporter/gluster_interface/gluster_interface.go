package gluster_interface

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strings"
)

type ginterface interface {
	GetOnlinePeers() ([]Peer, error)
	GetHealInfo(vol string) ([]HealEntries, error)
	GetVolStatus(vol string) ([]Volume, error)
}

type PeerStatus struct {
	XMLName xml.Name `xml:"cliOutput"`
	Peers   []Peer   `xml:"peerStatus>peer"`
}
type Peer struct {
	XMLName   xml.Name `xml:"peer"`
	UUID      string   `xml:"uuid"`
	Hostname  string   `xml:"hostname"`
	Connected int      `xml:"connected"`
}

type HealBricks struct {
	XMLName     xml.Name      `xml:"cliOutput"`
	Healentries []HealEntries `xml:"healInfo>bricks>brick"`
}
type HealEntries struct {
	XMLName        xml.Name `xml:"brick"`
	HostUUID       string   `xml:"hostUuid,attr"`
	Brickname      string   `xml:"name"`
	Connected      string   `xml:"status"`
	NumHealEntries float64  `xml:"numberOfEntries"`
}

type VolumeStatus struct {
	XMLName xml.Name `xml:"cliOutput"`
	Volumes []Volume `xml:"volStatus>volumes>volume"`
}

type Volume struct {
	XMLName   xml.Name `xml:"volume"`
	VolName   string   `xml:"volName"`
	NodeCount int      `xml:"nodeCount"`
	Node      []Node   `xml:"node"`
	Status    string
}
type Node struct {
	XMLName  xml.Name `xml:"node"`
	Hostname string   `xml:"hostname"`
	Path     string   `xml:"path"`
	PeerId   string   `xml:"peerid"`
	Status   string   `xml:"status"`
	Port     string   `xml:"port"`
	Ports    struct {
		TCP  string `xml:"tcp"`
		RDMA string `xml:"rdma"`
	} `xml:"ports"`
	Pid string `xml:"pid"`
}

type GD1 struct {
}

func ExecuteCmd(cmd string) ([]byte, error) {
	cmd_fields := strings.Fields(cmd)
	cmd_str := cmd_fields[0]
	args := cmd_fields[1:]
	fmt.Println("executing", cmd_str, args)
	out, err := exec.Command(cmd_str, args...).Output()
	return out, err
}

func (g GD1) GetOnlinePeers() ([]Peer, error) {
	cmd := "gluster peer status --xml"
	out, err := ExecuteCmd(cmd)
	if err != nil {
		return nil, err
	}
	var result PeerStatus
	err = xml.Unmarshal(out, &result)
	var onlinePeers []Peer
	for _, peer := range result.Peers {
		if peer.Connected == 1 {
			onlinePeers = append(onlinePeers, peer)
		}
	}
	return onlinePeers, err
}
func (g GD1) GetHealInfo(vol string) ([]HealEntries, error) {
	cmd := fmt.Sprintf("gluster vol heal %s info --xml", vol)
	out, err := ExecuteCmd(cmd)
	if err != nil {
		return nil, err
	}
	var healop HealBricks
	err = xml.Unmarshal(out, &healop)
	return healop.Healentries, err
}
func (g GD1) GetVolStatus(vol string) ([]Volume, error) {

	cmd := fmt.Sprintf("gluster vol status %s --xml", vol)

	out, err := ExecuteCmd(cmd)
	if err != nil {
		return nil, err
	}
	var volstatus VolumeStatus
	err = xml.Unmarshal(out, &volstatus)
	if err == nil {
		Volumes := volstatus.Volumes
		for index := 0; index < len(Volumes); index++ {
			// get brick status
			vol := &Volumes[index]
			for bindex := 0; bindex < len(vol.Node); bindex++ {
				brick := &vol.Node[bindex]
				if brick.Status == "1" {
					vol.Status = "Started"
					break
				}

			}
		}
	}
	return volstatus.Volumes, err
}

type GD2 struct {
	Host string
	Port string
}

func MakeGD2(host string, port string) GD2 {
	gd2 := GD2{host, port}
	return gd2
}

func (g GD2) GetOnlinePeers() ([]Peer, error) {
	url := "http://" + g.Host + ":" + g.Port + "/v1/peers"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	body, readErr := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if readErr != nil {
		return nil, readErr
	}
	type M map[string]interface{}
	var peers []M
	err = json.Unmarshal(body, &peers)
	if err != nil {
		return nil, err
	}
	var online_peers []Peer

	for _, peer := range peers {
		UUID := peer["id"].(string)
		online := peer["online"].(bool)
		if online {
			peer_addresses := peer["peer-addresses"].([]interface{})
			var list_addresses []string
			for _, i := range peer_addresses {
				addr := i.(string)
				list_addresses = append(list_addresses, addr)
			}
			addresses := strings.Join(list_addresses, ",")
			online_peer := Peer{UUID: UUID, Hostname: addresses, Connected: 1}
			online_peers = append(online_peers, online_peer)
		}
	}
	return online_peers, nil
}

func (g GD2) GetAllVolStatus() ([]Volume, error) {
	url := "http://" + g.Host + ":" + g.Port + "/v1/volumes"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	body, readErr := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if readErr != nil {
		return nil, readErr
	}
	type M map[string]interface{}
	var volumes []M
	err = json.Unmarshal(body, &volumes)
	if err != nil {
		return nil, err
	}
	var Volumes []Volume
	for _, volume := range volumes {
		v := GetVolumeObj(volume)
		Volumes = append(Volumes, v)
	}
	return Volumes, nil
}

func GetVolumeObj(volume map[string]interface{}) Volume {
	var name string
	var state string
	var NodeMap = make(map[string]string)
	var List_Nodes []Node
	for key, val := range volume {
		switch key {
		case "name":
			name = val.(string)
		case "state":
			state = val.(string)
		case "subvols":
			// get bricks
			subvols := val.([]interface{})
			for _, i := range subvols {
				subvol := i.(map[string]interface{})
				bricks := subvol["bricks"].([]interface{})
				for _, j := range bricks {
					brick := j.(map[string]interface{})
					host := brick["host"].(string)
					peer := brick["peer-id"].(string)
					path := brick["path"].(string)
					keystr := host + ":" + path
					_, present := NodeMap[host]
					if !present {
						NodeMap[keystr] = peer
						N1 := Node{Hostname: host, Path: path, PeerId: peer,
							Status: "NA", Port: "NA", Pid: "NA"}
						List_Nodes = append(List_Nodes, N1)
					}
				}
			}
		} // end key
	}
	Count := len(NodeMap)
	vol := Volume{VolName: name, NodeCount: Count,
		Node: List_Nodes, Status: state}
	return vol
}

func (g GD2) GetVolStatus(vol string) ([]Volume, error) {
	if vol == "" {
		return g.GetAllVolStatus()
	}
	url := "http://" + g.Host + ":" + g.Port + "/v1/volumes/" + vol
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	body, readErr := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if readErr != nil {
		return nil, readErr
	}
	type M map[string]interface{}
	var volume M
	err = json.Unmarshal(body, &volume)
	if err != nil {
		return nil, err
	}
	var list_volumes []Volume
	v := GetVolumeObj(volume)
	list_volumes = append(list_volumes, v)
	return list_volumes, nil
}

func (g GD2) GetHealInfo(vol string) ([]HealEntries, error) {
	url := "http://" + g.Host + ":" + g.Port + "/v1/volumes/" + vol +
		"/heal-info"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	body, readErr := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if readErr != nil {
		return nil, readErr
	}
	type jsonobj map[string]interface{}
	var HealInfo []jsonobj
	err = json.Unmarshal(body, &HealInfo)
	if err != nil {
		return nil, err
	}
	var List_Heals []HealEntries
	for _, heal := range HealInfo {
		uuid := heal["host-id"].(string)
		name := heal["name"].(string)
		entries := heal["entries"].(float64)
		connected := heal["status"].(string)
		HealEntry := HealEntries{HostUUID: uuid, Brickname: name,
			Connected: connected, NumHealEntries: entries}
		List_Heals = append(List_Heals, HealEntry)
	}
	return List_Heals, nil

}
