package peer

import (
	"blockchain/coin"
	"blockchain/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Client struct {
	Peers *Peers
}

func NewClient(p *Peers) *Client {
	return &Client{
		Peers: p,
	}
}

func (c *Client) BroadcastBlock(block coin.Block, thisPeer string) coin.Block {
	j, err := json.Marshal(block)
	utils.CheckError(err)

	for _, peer := range c.Peers.Hostnames {
		if peer != thisPeer {
			body := bytes.NewReader(j)
			_, err := http.Post(fmt.Sprintf("http://%s/block", peer), "application/json", body)
			utils.CheckError(err)
		}
	}

	return block
}

func (c *Client) GetPeers() map[string]string {
	resp, err := http.Get(fmt.Sprintf("http://%s/peers", seedHost))
	utils.CheckError(err)

	respBody, err := ioutil.ReadAll(resp.Body)
	var peers map[string]string

	err = json.Unmarshal(respBody, &peers)
	utils.CheckError(err)

	c.Peers.Hostnames = peers

	return peers
}

func (c *Client) BroadcastOnline(thisHostname string) {
	h := HostName{Hostname: thisHostname}

	j, err := json.Marshal(h)
	utils.CheckError(err)

	fmt.Println(string(j))

	for _, hostname := range c.Peers.Hostnames {
		body := bytes.NewReader(j)
		_, err := http.Post(fmt.Sprintf("http://%s/notify", hostname), "application/json", body)
		utils.CheckError(err)
	}
}

type HostName struct {
	Hostname string `json:"hostName"`
}
