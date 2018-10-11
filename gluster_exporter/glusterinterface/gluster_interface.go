package glusterinterface

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

type peerStatus struct {
	XMLName xml.Name `xml:"cliOutput"`
	Peers   []Peer   `xml:"peerStatus>peer"`
}

// Peer struct describe gluster peer
type Peer struct {
	XMLName   xml.Name `xml:"peer"`
	UUID      string   `xml:"uuid"`
	Hostname  string   `xml:"hostname"`
	Connected int      `xml:"connected"`
}

type healBricks struct {
	XMLName     xml.Name      `xml:"cliOutput"`
	Healentries []HealEntries `xml:"healInfo>bricks>brick"`
}

// HealEntries describe gluster heal info for each brick
type HealEntries struct {
	XMLName        xml.Name `xml:"brick"`
	HostUUID       string   `xml:"hostUuid,attr"`
	Brickname      string   `xml:"name"`
	Connected      string   `xml:"status"`
	NumHealEntries float64  `xml:"numberOfEntries"`
}

type volumeStatus struct {
	XMLName xml.Name `xml:"cliOutput"`
	Volumes []Volume `xml:"volStatus>volumes>volume"`
}

// Volume describes gluster volume
type Volume struct {
	XMLName   xml.Name `xml:"volume"`
	VolName   string   `xml:"volName"`
	NodeCount int      `xml:"nodeCount"`
	Node      []Node   `xml:"node"`
	Status    string
}

// Node describes gluster brick node
type Node struct {
	XMLName  xml.Name `xml:"node"`
	Hostname string   `xml:"hostname"`
	Path     string   `xml:"path"`
	PeerID   string   `xml:"peerid"`
	Status   string   `xml:"status"`
	Port     string   `xml:"port"`
	Ports    struct {
		TCP  string `xml:"tcp"`
		RDMA string `xml:"rdma"`
	} `xml:"ports"`
	Pid string `xml:"pid"`
}

// GD1 enables users to interact with gd1 version
type GD1 struct {
}

// ExecuteCmd enables to execute system cmds and returns stdout, err
func ExecuteCmd(cmd string) ([]byte, error) {
	cmdfields := strings.Fields(cmd)
	cmdstr := cmdfields[0]
	args := cmdfields[1:]
	fmt.Println("executing", cmdstr, args)
	out, err := exec.Command(cmdstr, args...).Output()
	return out, err
}

// GetOnlinePeers gets you the online gluster peers, look for Peer struct
func (g GD1) GetOnlinePeers() ([]Peer, error) {
	cmd := "gluster peer status --xml"
	out, err := ExecuteCmd(cmd)
	if err != nil {
		return nil, err
	}
	var result peerStatus
	err = xml.Unmarshal(out, &result)
	var onlinePeers []Peer
	for _, peer := range result.Peers {
		if peer.Connected == 1 {
			onlinePeers = append(onlinePeers, peer)
		}
	}
	return onlinePeers, err
}

// GetHealInfo gets gluster vol heal in []HealEntries
func (g GD1) GetHealInfo(vol string) ([]HealEntries, error) {
	cmd := fmt.Sprintf("gluster vol heal %s info --xml", vol)
	out, err := ExecuteCmd(cmd)
	if err != nil {
		return nil, err
	}
	var healop healBricks
	err = xml.Unmarshal(out, &healop)
	return healop.Healentries, err
}

// GetVolStatus gets the volume status if given vol name,
// otherwise gets all vols status
func (g GD1) GetVolStatus(vol string) ([]Volume, error) {

	cmd := fmt.Sprintf("gluster vol status %s --xml", vol)

	out, err := ExecuteCmd(cmd)
	if err != nil {
		return nil, err
	}
	var volstatus volumeStatus
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

// GD2 is struct to interact with Glusterd2 using REST API
type GD2 struct {
	Host string
	Port string
}

// MakeGD2 returns created gd2 instance, given host and port
func MakeGD2(host string, port string) GD2 {
	gd2 := GD2{host, port}
	return gd2
}

// GetOnlinePeers gets online peers from Glusterd2 using REST
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
	var onlinepeers []Peer

	for _, peer := range peers {
		UUID := peer["id"].(string)
		online := peer["online"].(bool)
		if online {
			peeraddresses := peer["peer-addresses"].([]interface{})
			var listaddresses []string
			for _, i := range peeraddresses {
				addr := i.(string)
				listaddresses = append(listaddresses, addr)
			}
			addresses := strings.Join(listaddresses, ",")
			onlinepeer := Peer{UUID: UUID, Hostname: addresses, Connected: 1}
			onlinepeers = append(onlinepeers, onlinepeer)
		}
	}
	return onlinepeers, nil
}

func (g GD2) getAllVolStatus() ([]Volume, error) {
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
		v := getVolumeObj(volume)
		Volumes = append(Volumes, v)
	}
	return Volumes, nil
}

func getVolumeObj(volume map[string]interface{}) Volume {
	var name string
	var state string
	var NodeMap = make(map[string]string)
	var ListNodes []Node
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
						N1 := Node{Hostname: host, Path: path, PeerID: peer,
							Status: "NA", Port: "NA", Pid: "NA"}
						ListNodes = append(ListNodes, N1)
					}
				}
			}
		} // end key
	}
	Count := len(NodeMap)
	vol := Volume{VolName: name, NodeCount: Count,
		Node: ListNodes, Status: state}
	return vol
}

// GetVolStatus of GD2 gives vol status using REST, if vol name not given
// then it gets status of all vols
func (g GD2) GetVolStatus(vol string) ([]Volume, error) {
	if vol == "" {
		return g.getAllVolStatus()
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
	var ListVolumes []Volume
	v := getVolumeObj(volume)
	ListVolumes = append(ListVolumes, v)
	return ListVolumes, nil
}

// GetHealInfo gets heal info from glusterd2 using rest api
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
	var ListHeals []HealEntries
	for _, heal := range HealInfo {
		uuid := heal["host-id"].(string)
		name := heal["name"].(string)
		entries := heal["entries"].(float64)
		connected := heal["status"].(string)
		HealEntry := HealEntries{HostUUID: uuid, Brickname: name,
			Connected: connected, NumHealEntries: entries}
		ListHeals = append(ListHeals, HealEntry)
	}
	return ListHeals, nil

}
