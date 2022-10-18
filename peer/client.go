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
	"time"

	"github.com/cenkalti/backoff/v4"
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

func (c *Client) BroadcastBlock(block coin.Block) (coin.Block, error) {
	for _, peer := range c.Peers.Hostnames {
		if peer != c.ThisPeer {
			resp, err := httpPostWithBackoff(fmt.Sprintf("http://%s/block", peer), block)
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

	return block, nil
}

func (c *Client) getBlockchain(address string) (*coin.Blockchain, error) {
	resp, err := httpGetWithBackoff(fmt.Sprintf("http://%s/block-chain", address))
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
	resp, err := httpGetWithBackoff(fmt.Sprintf("http://%s/latest-block", peer))
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

func (c *Client) SpendCoin(spendCoinRelay SpendCoinRelay) (*http.Response, error) {

	ct := CreateTransactionControl{
		Address: spendCoinRelay.Address,
		Amount:  spendCoinRelay.Amount,
	}

	j, err := json.Marshal(ct)
	if err != nil {
		return nil, err
	}

	b := bytes.NewReader(j)

	resp, err := http.Post(fmt.Sprintf("http://%s/spend-coin", spendCoinRelay.Host), "application/json", b)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) GetTxPoolFromPeer(peer string) (map[repository.TxIDType]repository.Transaction, error) {
	resp, err := httpGetWithBackoff(fmt.Sprintf("http://%s/txpool", peer))
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

func (c *Client) GetPeers(hostName string) (map[string]string, error) {
	resp, err := httpGetWithBackoff(fmt.Sprintf("http://%s/peers", hostName))
	respBody, err := ioutil.ReadAll(resp.Body)
	var peers map[string]string

	err = json.Unmarshal(respBody, &peers)
	if err != nil {
		return nil, err
	}

	return peers, nil
}

func (c *Client) GetHosts(hostName string, excludedHosts map[string]Details) ([]Details, error) {
	resp, err := httpPostWithBackoff(fmt.Sprintf("http://%s/hosts", hostName), excludedHosts)
	respBody, err := ioutil.ReadAll(resp.Body)
	var peers []Details

	err = json.Unmarshal(respBody, &peers)
	if err != nil {
		return nil, err
	}

	return peers, nil
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

	fmt.Println("Notifying these hosts: ", h)

	for _, hostname := range c.Peers.Hostnames {
		_, err := httpPostWithBackoff(fmt.Sprintf("http://%s/notify", hostname), h)
		if err != nil {
			utils.ErrorLogger.Println(err)
		}
	}
}

func (c *Client) BroadcastTransaction(tx repository.Transaction) error {
	for _, peer := range c.Peers.Hostnames {
		if peer != c.ThisPeer {
			resp, err := httpPostWithBackoff(fmt.Sprintf("http://%s/transaction", peer), tx)
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
	if body == nil {
		return ""
	}

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

var (
	DefaultInitialInterval     = 500 * time.Millisecond
	DefaultRandomizationFactor = 0.5
	DefaultMultiplier          = 1.5
	DefaultMaxInterval         = 5 * time.Second
	DefaultMaxElapsedTime      = 2 * time.Minute
)

func NewExponentialBackOff() *backoff.ExponentialBackOff {
	b := &backoff.ExponentialBackOff{
		InitialInterval:     DefaultInitialInterval,
		RandomizationFactor: DefaultRandomizationFactor,
		Multiplier:          DefaultMultiplier,
		MaxInterval:         DefaultMaxInterval,
		MaxElapsedTime:      DefaultMaxElapsedTime,
		Stop:                backoff.Stop,
		Clock:               backoff.SystemClock,
	}
	b.Reset()
	return b
}

func httpPostWithBackoff(url string, body interface{}) (*http.Response, error) {
	var resp *http.Response
	var err error

	j, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	b := bytes.NewReader(j)

	op := func() error {
		resp, err = http.Post(url, "application/json", b)
		if err != nil {
			utils.ErrorLogger.Printf("%s", err)
			return err
		}
		if resp.StatusCode > 399 {
			utils.ErrorLogger.Printf("received response code %d", resp.StatusCode)
			// DefaultMaxElapsedTime = 0 * time.Second
			return fmt.Errorf(readResponseBody(resp.Body))
		}

		return nil
	}

	err = backoff.Retry(op, backoff.NewExponentialBackOff())
	if err != nil {
		utils.ErrorLogger.Printf("%s", err)
		return nil, err
	}

	return resp, nil
}

func httpGetWithBackoff(url string) (*http.Response, error) {
	var resp *http.Response
	var err error

	op := func() error {
		resp, err = http.Get(url)
		if err != nil {
			utils.ErrorLogger.Printf("%s", err)

			return err
		}

		if resp.StatusCode > 399 {
			utils.ErrorLogger.Printf("%s", err)
			return fmt.Errorf(readResponseBody(resp.Body))
		}

		return nil
	}

	err = backoff.Retry(op, backoff.NewExponentialBackOff())
	if err != nil {
		utils.ErrorLogger.Printf("%s", err)
		return nil, err
	}

	return resp, nil
}
