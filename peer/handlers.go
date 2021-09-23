package peer

import (
	"blockchain/coin"
	"blockchain/service"
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

func (c *CoinServerHandler) transaction(r *http.Request) (*HTTPResponse, *HTTPError) {
	switch r.Method {
	case "POST":
		t := wallet.Transaction{}

		err := readBody(r, &t)
		if err != nil {
			return nil, &HTTPError{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			}
		}

		err = wallet.IsValidTransaction(t, c.BlockchainService.UTxOSet)
		if err != nil {
			return nil, &HTTPError{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			}
		}

		*c.BlockchainService.UnconfirmedTransactionPool = append(*c.BlockchainService.UnconfirmedTransactionPool, t)

		return &HTTPResponse{
			StatusCode: http.StatusCreated,
			Body:       t,
		}, nil
	}

	return nil, &HTTPError{
		Code: http.StatusMethodNotAllowed,
	}
}

func (c *CoinServerHandler) spendMoney(r *http.Request) (*HTTPResponse, *HTTPError) {
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

		transaction, err := c.BlockchainService.SpendMoney(cc.Address, cc.Amount)
		if err != nil {
			return nil, &HTTPError{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			}
		}

		tID := wallet.GenerateTransactionID(*transaction)
		if !reflect.DeepEqual(tID, transaction.ID) {
			return nil, &HTTPError{
				Code:    http.StatusBadRequest,
				Message: "unequal tsek",
			}
		}

		err = wallet.IsValidTransaction(*transaction, c.BlockchainService.UTxOSet)
		if err != nil {
			return nil, &HTTPError{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			}
		}

		err = c.Client.BroadcastTransaction(*transaction)
		if err != nil {
			return nil, &HTTPError{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			}
		}

		return &HTTPResponse{
			StatusCode: http.StatusCreated,
			Body:       transaction,
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

		err = c.BlockchainService.ValidateAndAddBlockToBlockchain(block)

		if err != nil {
			return nil, &HTTPError{
				Code:    http.StatusInternalServerError,
				Message: fmt.Sprintf("Could not update blockchain. error: %s", err.Error()),
			}
		}

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

		block, blockchain, uTxOSet, err := c.BlockchainService.CreateNextBlock()
		if err != nil {
			return nil, &HTTPError{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			}
		}

		c.Client.BroadcastBlock(*block, c.Client.ThisPeer)

		payload := struct {
			Blocks              []coin.Block                                                    `json:"blocks"`
			UnspentTransactions map[wallet.PublicKeyAddressType]map[wallet.TxIDType]wallet.UTxO `json:"unspentTransactions"`
		}{
			Blocks:              blockchain.Blocks,
			UnspentTransactions: *uTxOSet,
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
