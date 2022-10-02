package peer

import (
	"bytes"
	"encoding/json"
	"firstcoin/coin"
	"firstcoin/repository"
	"firstcoin/service"
	"firstcoin/utils"
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
	utils.PanicError(err)

	for _, peer := range c.Peers.Hostnames {
		if peer != c.ThisPeer {
			body := bytes.NewReader(j)
			resp, err := http.Post(fmt.Sprintf("http://%s/block", peer), "application/json", body)
			if err != nil {
				utils.ErrorLogger.Println(fmt.Sprintf("error when posting block %s", err))

				// Remove host if error for now
				c.Peers.RemoveHostname(peer)
				continue
			}
			if resp.StatusCode >= 400 {
				utils.ErrorLogger.Println(fmt.Sprintf("error from peer when broadcasting block %s", readResponseBody(resp.Body)))
			}

		}
	}

	return block
}

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

	// if err := bc.IsValidBlockchain(); err != nil {
	// 	c.Blockchain = nil
	// 	return nil, fmt.Errorf("invalid firstcoin. error: %s", err.Error())
	// }

	return &bc, nil
}

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

func (c *Client) SpendCoin(spendCoinRelay SpendCoinRelay) error {
	ct := CreateTransactionControl{
		Address: spendCoinRelay.Address,
		Amount:  spendCoinRelay.Amount,
	}

	j, err := json.Marshal(ct)
	if err != nil {
		return err
	}

	body := bytes.NewReader(j)

	_, err = http.Post(fmt.Sprintf("http://%s/spend-coin", spendCoinRelay.Host), "application/json", body)

	if err != nil {
		return err
	}

	return nil
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

func (c *Client) GetPeers(hostName string) map[string]string {
	resp, err := http.Get(fmt.Sprintf("http://%s/peers", hostName))
	utils.PanicError(err)

	respBody, err := ioutil.ReadAll(resp.Body)
	var peers map[string]string

	err = json.Unmarshal(respBody, &peers)
	utils.PanicError(err)

	return peers
}

func (c *Client) GetHosts(hostName string, excludedHosts map[string]Details) []Details {
	j, err := json.Marshal(excludedHosts)
	utils.PanicError(err)
	body := bytes.NewReader(j)

	resp, err := http.Post(fmt.Sprintf("http://%s/hosts", hostName), "application/json", body)
	utils.PanicError(err)

	respBody, err := ioutil.ReadAll(resp.Body)
	var peers []Details

	err = json.Unmarshal(respBody, &peers)
	utils.PanicError(err)

	return peers
}

// TODO: This will not work - cant simply take the longest chain - malice could have one block longer - should take the one that is 2 or 3 blocks longer
func (c *Client) QueryPeersForBlockchain(peers map[string]string) error {
	for address, _ := range peers {
		if address == c.ThisPeer {
			continue
		}

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

			c.Blockchain.ReplaceBlockchain(*forkChain)
			// if err := bc.IsValidBlockchain(); err == nil {
			// } else {
			// 	return err
			// }

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

	err := replayBlockChainTransactions(*c.Blockchain)
	if err != nil {
		utils.ErrorLogger.Println(err)
		repository.ClearUTxOSet()
	}

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
	utils.PanicError(err)

	fmt.Println("Notifying these hosts: ", string(j))

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

func replayBlockChainTransactions(bc coin.Blockchain) error {
	if err := bc.Blocks[0].IsGenesisBlock(); err != nil {
		return fmt.Errorf("Invalid firstcoin: %s. error: %s", "invalid genesis block", err.Error())
	}

	err := service.CommitBlockTransactions(bc.Blocks[0])
	if err != nil {
		return err
	}

	for _, block := range bc.Blocks[1:] {
		err := service.CommitBlockTransactions(block)
		if err != nil {
			return err
		}
	}

	return nil
}
