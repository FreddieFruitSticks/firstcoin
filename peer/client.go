package peer

import (
	"blockchain/coin"
	"blockchain/repository"
	"blockchain/service"
	"blockchain/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

func (c *Client) BroadcastBlock(block coin.Block) coin.Block {
	j, err := json.Marshal(block)
	utils.CheckError(err)

	for _, peer := range c.Peers.Hostnames {
		if peer != c.ThisPeer {
			body := bytes.NewReader(j)
			resp, err := http.Post(fmt.Sprintf("http://%s/block", peer), "application/json", body)
			if resp.StatusCode >= 400 {
				utils.ErrorLogger.Println(fmt.Sprintf("error from peer when broadcasting block %s", readResponseBody(resp.Body)))
			}
			if err != nil {
				utils.ErrorLogger.Println(fmt.Sprintf("error when posting block %s", err))

				// Remove host if error for now
				c.Peers.RemoveHostname(peer)
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

	if err := bc.IsValidBlockchain(); err != nil {
		c.Blockchain = nil
		return nil, fmt.Errorf("invalid blockchain. error: %s", err.Error())
	}

	return &bc, nil
}

// Fetch latest block from seed host for now
func (c *Client) GetLatestBlockFromPeer(peer string) (*coin.Block, error) {
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

func (c *Client) GetTxPoolFromPeer(peer string) (map[repository.TxIDType]repository.Transaction, error) {
	resp, err := http.Get(fmt.Sprintf("http://%s/txpool", peer))
	if err != nil {
		return nil, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	var txPool map[repository.TxIDType]repository.Transaction

	err = json.Unmarshal(respBody, &txPool)
	if err != nil {
		return nil, err
	}

	return txPool, nil
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

// TODO: This will not work - cant simply take the longest chain - malice could have one block longer - should take the one that is 2 or 3 blocks longer
func (c *Client) QueryPeersForBlockchain(peers map[string]string) error {
	for address, _ := range peers {
		block, err := c.GetLatestBlockFromPeer(address)
		if err != nil {
			return err
		}

		if len(c.Blockchain.Blocks) == 0 {
			bc, err := c.getBlockchain(address)
			if err != nil {
				return err
			}

			forkChain := coin.NewBlockchain(bc.Blocks)

			if err := bc.IsValidBlockchain(); err == nil {
				c.Blockchain.ReplaceBlockchain(*forkChain)
			} else {
				return err
			}

		}

		if len(c.Blockchain.Blocks) > 0 && block.Index > c.Blockchain.GetLastBlock().Index {
			forkChain, err := c.getBlockchain(address)
			if err != nil {
				return err
			}

			if err := forkChain.IsValidBlockchain(); err != nil {
				return err
			}
			c.Blockchain.ReplaceBlockchain(*forkChain)
		}
	}

	replayBlockChainTransactions(*c.Blockchain)

	return nil
}

func (c *Client) QueryNetworkForUnconfirmedTxPool(peers map[string]string) error {
	for address, _ := range peers {
		txPool, err := c.GetTxPoolFromPeer(address)
		if err != nil {
			utils.ErrorLogger.Println(fmt.Sprintf("couldn't fetch txpool from peer %s. error: %s", address, err))
			continue
		}

		repository.SetTxPool(txPool)
		return nil
	}

	return fmt.Errorf("could not find txpool from any of the peers")

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

func (c *Client) BroadcastTransaction(tx repository.Transaction) error {
	transaction, err := json.Marshal(tx)
	if err != nil {
		return err
	}

	for _, peer := range c.Peers.Hostnames {
		if peer != c.ThisPeer {
			body := bytes.NewReader(transaction)
			resp, err := http.Post(fmt.Sprintf("http://%s/transaction", peer), "application/json", body)
			if err != nil {
				utils.ErrorLogger.Println(fmt.Sprintf("peer %s rejected transaction. error: %s", peer, err))
				continue
			}

			if resp.StatusCode >= 400 {
				utils.ErrorLogger.Println(fmt.Sprintf("peer %s rejected transaction. error: %s", peer, readResponseBody(resp.Body)))
			}

		}
	}

	return nil
}

func readResponseBody(body io.ReadCloser) string {
	buf := new(bytes.Buffer)
	buf.ReadFrom(body)
	return buf.String()
}

func replayBlockChainTransactions(bc coin.Blockchain) {
	for _, block := range bc.Blocks {
		service.CommitBlockTransactions(block)
	}
}
