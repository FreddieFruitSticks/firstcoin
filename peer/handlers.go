package peer

import (
	"blockchain/coin"
	"blockchain/repository"
	"blockchain/service"
	"blockchain/utils"
	"blockchain/wallet"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
)

type CoinServerHandler struct {
	Peers             *Peers
	Client            *Client
	BlockchainService service.BlockchainService
}

func NewCoinServerHandler(s service.BlockchainService, c *Client, p *Peers) *CoinServerHandler {
	return &CoinServerHandler{
		Peers:             p,
		Client:            c,
		BlockchainService: s,
	}
}

func (c *CoinServerHandler) latestBlock(r *http.Request) (*HTTPResponse, *HTTPError) {
	switch r.Method {
	case "GET":
		latestBlock := c.BlockchainService.Blockchain.Blocks[len(c.BlockchainService.Blockchain.Blocks)-1]

		return &HTTPResponse{
			StatusCode: http.StatusCreated,
			Body:       latestBlock,
		}, nil
	}

	return nil, &HTTPError{
		Code: http.StatusMethodNotAllowed,
	}
}

func (c *CoinServerHandler) receiveTransaction(r *http.Request) (*HTTPResponse, *HTTPError) {
	switch r.Method {
	case "POST":
		tx := repository.Transaction{}

		err := readBody(r, &tx)
		if err != nil {
			return nil, &HTTPError{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			}
		}

		_, ok := repository.GetTxFromTxPool(tx.ID)
		if ok {
			utils.InfoLogger.Println(fmt.Sprintf("tx exists in txPool %s", tx))
			return &HTTPResponse{
				StatusCode: http.StatusNotModified,
				Body:       tx,
			}, nil
		}

		err = wallet.IsValidTransaction(tx)
		if err != nil {
			utils.ErrorLogger.Println(fmt.Sprintf("received tx is invalid. error: %s", err.Error()))
			return nil, &HTTPError{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			}
		}

		_, err = service.ValidateTxPoolDryRun(c.BlockchainService.Blockchain.GetLastBlock().Index, &tx)
		if err != nil {
			utils.ErrorLogger.Println(fmt.Sprintf("txPool is invalid. error: %s", err.Error()))
			return nil, &HTTPError{
				Code:    http.StatusBadRequest,
				Message: fmt.Sprintf("txPool is invalid. error: %s", err.Error()),
			}
		}

		ok = c.BlockchainService.AddTxToTxPool(tx)
		if ok {
			utils.InfoLogger.Println("Received new Tx and added to pool. Relaying tx pool")
			// c.Client.BroadcastTransaction(tx)
		}

		return &HTTPResponse{
			StatusCode: http.StatusCreated,
			Body:       tx,
		}, nil
	}

	return nil, &HTTPError{
		Code: http.StatusMethodNotAllowed,
	}
}

func (c *CoinServerHandler) spendCoin(r *http.Request) (*HTTPResponse, *HTTPError) {
	switch r.Method {
	case "POST":
		cc := CreateTransactionControl{}

		err := readBody(r, &cc)
		if err != nil {
			return nil, &HTTPError{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			}
		}

		tx, err := c.BlockchainService.CreateTx(cc.Address, cc.Amount)
		if err != nil {
			return nil, &HTTPError{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			}
		}

		tID := wallet.GenerateTransactionID(*tx)
		if !reflect.DeepEqual(tID, tx.ID) {
			return nil, &HTTPError{
				Code:    http.StatusBadRequest,
				Message: "unequal tsek",
			}
		}

		err = wallet.IsValidTransaction(*tx)
		if err != nil {
			return nil, &HTTPError{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			}
		}

		_, err = service.ValidateTxPoolDryRun(c.BlockchainService.Blockchain.GetLastBlock().Index, tx)
		if err != nil {
			utils.ErrorLogger.Println(fmt.Sprintf("txPool is invalid. error: %s", err.Error()))
			return nil, &HTTPError{
				Code:    http.StatusBadRequest,
				Message: fmt.Sprintf("txPool is invalid. error: %s", err.Error()),
			}
		}

		repository.AddTxToTxPool(*tx)

		err = c.Client.BroadcastTransaction(*tx)
		if err != nil {
			return nil, &HTTPError{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			}
		}

		return &HTTPResponse{
			StatusCode: http.StatusCreated,
			Body:       tx,
		}, nil

	}

	return nil, &HTTPError{
		Code: http.StatusMethodNotAllowed,
	}
}

func (c *CoinServerHandler) getTxPool(r *http.Request) (*HTTPResponse, *HTTPError) {
	switch r.Method {
	case "GET":

		return &HTTPResponse{
			StatusCode: http.StatusOK,
			Body:       repository.GetTxPool(),
		}, nil

	}

	return nil, &HTTPError{
		Code: http.StatusMethodNotAllowed,
	}
}

func (c *CoinServerHandler) getTxSet(r *http.Request) (*HTTPResponse, *HTTPError) {
	switch r.Method {
	case "GET":

		return &HTTPResponse{
			StatusCode: http.StatusOK,
			Body:       repository.GetEntireUTxOSet(),
		}, nil

	}

	return nil, &HTTPError{
		Code: http.StatusMethodNotAllowed,
	}
}

func (c *CoinServerHandler) getBlockchain(r *http.Request) (*HTTPResponse, *HTTPError) {
	switch r.Method {
	case "GET":

		return &HTTPResponse{
			StatusCode: http.StatusOK,
			Body:       c.BlockchainService.Blockchain.Blocks,
		}, nil

	}

	return nil, &HTTPError{
		Code: http.StatusMethodNotAllowed,
	}
}

func (c *CoinServerHandler) peers(r *http.Request) (*HTTPResponse, *HTTPError) {
	switch r.Method {
	case "POST":
		t := HostName{}

		err := readBody(r, &t)
		if err != nil {
			return nil, &HTTPError{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			}
		}

		c.Peers.AddHostname(t.Hostname)

		return &HTTPResponse{
			StatusCode: http.StatusOK,
		}, nil

	case "GET":
		return &HTTPResponse{
			StatusCode: http.StatusOK,
			Body:       c.Peers.Hostnames,
		}, nil
	}

	return nil, &HTTPError{
		Code: http.StatusMethodNotAllowed,
	}
}

func (c *CoinServerHandler) blockChain(r *http.Request) (*HTTPResponse, *HTTPError) {
	switch r.Method {
	case "GET":
		return &HTTPResponse{
			StatusCode: http.StatusOK,
			Body:       c.BlockchainService.Blockchain,
		}, nil
	}

	return nil, &HTTPError{
		Code: http.StatusMethodNotAllowed,
	}
}

func (c *CoinServerHandler) addBlockToBlockchain(r *http.Request) (*HTTPResponse, *HTTPError) {
	switch r.Method {
	case "POST":
		block := coin.Block{}
		err := readBody(r, &block)
		if err != nil {
			return nil, &HTTPError{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			}
		}

		if block.Index == c.BlockchainService.Blockchain.GetLastBlock().Index {
			utils.InfoLogger.Println("Block already exists in blockchain")
			return &HTTPResponse{
				StatusCode: http.StatusAlreadyReported,
				Body:       c.BlockchainService.Blockchain,
			}, nil
		}

		if err := block.IsValidBlock(c.BlockchainService.Blockchain.GetLastBlock()); err != nil {
			utils.ErrorLogger.Println(err)
			return nil, &HTTPError{
				Code:    http.StatusBadRequest,
				Message: fmt.Sprintf("Could not update blockchain. error: %s", err.Error()),
			}
		}

		// TODO: This needs to be added to a fork (need to implement forks first). It is not a given that the block should be accepted
		//just because it has valid POW, and "fits" on to the chain.
		c.BlockchainService.Blockchain.AddBlock(block)
		service.CommitBlockTransactions(block)

		// this is relaying an accepted block to the network. Right now it simply sends to all the peers. The node that originally sent
		// the block only adds it block to its own chain if it receives it back from the network.
		c.Client.BroadcastBlock(block)

		return &HTTPResponse{
			StatusCode: http.StatusCreated,
			Body:       c.BlockchainService.Blockchain,
		}, nil
	}

	return nil, &HTTPError{
		Code: http.StatusMethodNotAllowed,
	}
}

func (c *CoinServerHandler) createBlock(r *http.Request) (*HTTPResponse, *HTTPError) {
	switch r.Method {
	case "POST":
		data := BlockDataControl{}
		err := readBody(r, &data)
		if err != nil {
			return nil, &HTTPError{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			}
		}

		block, blockchain, err := c.BlockchainService.CreateNextBlock()
		if err != nil {
			return nil, &HTTPError{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			}
		}
		c.Client.BroadcastBlock(*block)

		payload := struct {
			Blocks              []coin.Block                                                                `json:"blocks"`
			UnspentTransactions map[repository.PublicKeyAddressType]map[repository.TxIDType]repository.UTxO `json:"unspentTransactions"`
		}{
			Blocks:              blockchain.Blocks,
			UnspentTransactions: repository.GetEntireUTxOSet(),
		}

		return &HTTPResponse{
			StatusCode: http.StatusCreated,
			Body:       payload,
		}, nil
	}

	return nil, &HTTPError{
		Code: http.StatusMethodNotAllowed,
	}
}

func readBody(request *http.Request, params interface{}) error {
	reqBody, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return fmt.Errorf("cant read body %s. error: %s", request.Body, err.Error())
	}

	unmarshalErr := json.Unmarshal(reqBody, params)
	if unmarshalErr != nil {
		return fmt.Errorf("cant unmarshal body %+v in to %s. error: %s", request.Body, reflect.TypeOf(params), unmarshalErr.Error())
	}

	return nil
}

type BlockDataControl struct {
	Data string `json:"data"`
}

type HostName struct {
	Hostname string `json:"hostName"`
}

type CreateTransactionControl struct {
	Address []byte `json:"address"`
	Amount  int    `json:"amount"`
}
