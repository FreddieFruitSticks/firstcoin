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
	Peers      *Peers
	Blockchain *coin.Blockchain
}

func NewClient(p *Peers, b *coin.Blockchain) *Client {
	return &Client{
		Peers:      p,
		Blockchain: b,
	}
}

func (c *Client) BroadcastBlock(block coin.Block, thisPeer string) coin.Block {
	j, err := json.Marshal(block)
	utils.CheckError(err)

	for _, peer := range c.Peers.Hostnames {
		if peer != thisPeer {
			body := bytes.NewReader(j)
			resp, err := http.Post(fmt.Sprintf("http://%s/block", peer), "application/json", body)
			if err != nil {
				fmt.Println(err)

				// Remove host if error for now
				c.Peers.RemoveHostname(peer)
			} else {
				if resp.StatusCode >= 400 {
					c.Peers.RemoveHostname(peer)
				}
			}

		}
	}

	return block
}

// Fetch from seed host for now
func (c *Client) GetBlockchain() (*coin.Blockchain, error) {
	resp, err := http.Get(fmt.Sprintf("http://%s/block-chain", seedHost))
	if err != nil {
		return nil, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	var blocks []coin.Block

	err = json.Unmarshal(respBody, &blocks)
	if err != nil {
		return nil, err
	}

	c.Blockchain.SetBlockchain(blocks)

	if !c.Blockchain.IsValidBlockchain() {
		c.Blockchain = nil
		return nil, fmt.Errorf("invalid blockchain")
	}

	return c.Blockchain, nil
}

// Fetch latest block from seed host for now
func (c *Client) GetLatestBlock(peer string) (*coin.Block, error) {
	resp, err := http.Get(fmt.Sprintf("http://%s/latest-block", peer))
	if err != nil {
		return nil, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	var block coin.Block

	err = json.Unmarshal(respBody, &block)
	if err != nil {
		return nil, err
	}

	return &block, nil
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

func (c *Client) QueryPeers(peers map[string]string) {
	for address, _ := range peers {
		block, err := c.GetLatestBlock(address)
		if err != nil {
			fmt.Println(err)
		}

		if len(c.Blockchain.Blocks) == 0 {
			bc, err := c.GetBlockchain()
			if err != nil {
				fmt.Println(err)
			}

			c.Blockchain.SetBlockchain(bc.Blocks)

			return
		}

		if len(c.Blockchain.Blocks) > 0 && block.Index != c.Blockchain.GetLastBlock().Index {
			bc, err := c.GetBlockchain()
			if err != nil {
				fmt.Println(err)
			}
			c.Blockchain.SetBlockchain(bc.Blocks)
		}
	}

}

func (c *Client) BroadcastOnline(thisHostname string) {
	h := HostName{Hostname: thisHostname}

	j, err := json.Marshal(h)
	utils.CheckError(err)

	fmt.Println(string(j))

	for _, hostname := range c.Peers.Hostnames {
		body := bytes.NewReader(j)
		_, err := http.Post(fmt.Sprintf("http://%s/notify", hostname), "application/json", body)
		if err != nil {
			fmt.Println(err)
		}
	}
}

type HostName struct {
	Hostname string `json:"hostName"`
}
