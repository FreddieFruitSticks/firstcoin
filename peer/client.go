package peer

import (
	"blockchain/coin"
	"blockchain/utils"
	"blockchain/wallet"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Client struct {
	Peers      *Peers
	Blockchain *coin.Blockchain
	ThisPeer   string
}

func NewClient(p *Peers, b *coin.Blockchain, t string) *Client {
	return &Client{
		Peers:      p,
		Blockchain: b,
		ThisPeer:   t,
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
func (c *Client) getBlockchain(address string) (*coin.Blockchain, error) {
	resp, err := http.Get(fmt.Sprintf("http://%s/block-chain", address))
	if err != nil {
		return nil, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	var bc coin.Blockchain

	err = json.Unmarshal(respBody, &bc)
	if err != nil {
		return nil, err
	}

	if !bc.IsValidBlockchain() {
		c.Blockchain = nil
		return nil, fmt.Errorf("invalid blockchain")
	}

	return &bc, nil
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

func (c *Client) QueryPeers(peers map[string]string) error {
	for address, _ := range peers {
		block, err := c.GetLatestBlock(address)
		if err != nil {
			return err
		}

		if len(c.Blockchain.Blocks) == 0 {
			bc, err := c.getBlockchain(address)
			if err != nil {
				return err
			}

			forkChain := coin.NewBlockchain(bc.Blocks)

			if bc.IsValidBlockchain() {
				c.Blockchain.ReplaceBlockchain(*forkChain)
			}

			return nil
		}

		if len(c.Blockchain.Blocks) > 0 && block.Index > c.Blockchain.GetLastBlock().Index {
			forkChain, err := c.getBlockchain(address)
			if err != nil {
				return err
			}

			if forkChain.IsValidBlockchain() {
				c.Blockchain.ReplaceBlockchain(*forkChain)
			}
		}
	}

	return nil

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

func (c *Client) BroadcastTransaction(tx wallet.Transaction) error {
	transaction, err := json.Marshal(tx)
	if err != nil {
		return err
	}

	for _, peer := range c.Peers.Hostnames {
		if peer != c.ThisPeer {
			body := bytes.NewReader(transaction)
			resp, err := http.Post(fmt.Sprintf("http://%s/transaction", peer), "application/json", body)
			if err != nil {
				fmt.Println(err)

				// Remove host if error for now - assume its a fail peer
				c.Peers.RemoveHostname(peer)
			} else {
				if resp.StatusCode >= 400 {
					c.Peers.RemoveHostname(peer)
				}
			}

		}
	}

	return nil
}
