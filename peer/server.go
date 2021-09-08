package peer

import (
	"blockchain/coin"
	"blockchain/service"
	"blockchain/utils"
	"blockchain/wallet"
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
)

// TODO remove all non-server related stuff to a new package - need refactor
type Server struct {
	Peers             *Peers
	Client            *Client
	BlockchainService service.BlockchainService
}

type BlockDataControl struct {
	Data string `json:"data"`
}

type HostName struct {
	Hostname string `json:"hostName"`
}

type CreateTransactionControl struct {
	Address string `json:"address"`
	Amount  int    `json:"amount"`
}

type errorMessage struct {
	ErrorMessage string `json:"errorMessage"`
}

func NewServer(s service.BlockchainService, p *Peers, c *Client) *Server {
	return &Server{
		BlockchainService: s,
		Peers:             p,
		Client:            c,
	}
}

func (s *Server) HandleServer(port string) {
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "pong from, %q", html.EscapeString(r.URL.Path))
	})

	// control endpoint
	http.HandleFunc("/create-block", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			b, err := ioutil.ReadAll(r.Body)
			utils.CheckError(err)

			data := BlockDataControl{}
			err = json.Unmarshal(b, &data)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				err = json.NewEncoder(w).Encode(err)
				return
			}

			valid, block, blockchain, uTxOs := s.BlockchainService.CreateNextBlock()
			if !valid {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				err = json.NewEncoder(w).Encode("Block not valid")

				return
			}

			s.Client.BroadcastBlock(*block, s.Client.ThisPeer)

			payload := struct {
				Blocks              []coin.Block                        `json:"blocks"`
				UnspentTransactions map[string]map[string]wallet.UTxOut `json:"unspentTransactions"`
			}{
				Blocks:              blockchain.Blocks,
				UnspentTransactions: *uTxOs,
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)

			// byte arrays base64 encode and decode (on the other end). So the txOut address encodes as b64
			err = json.NewEncoder(w).Encode(payload)
		}
	})

	http.HandleFunc("/block", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":

			b, err := ioutil.ReadAll(r.Body)
			utils.CheckError(err)

			block := coin.Block{}
			err = json.Unmarshal(b, &block)
			utils.CheckError(err)

			hasUpdated := s.BlockchainService.AddBlockToBlockchain(block)

			if hasUpdated {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				err = json.NewEncoder(w).Encode("Could not update blockchain")
			}

		}
	})

	http.HandleFunc("/block-chain", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			err := json.NewEncoder(w).Encode(s.BlockchainService.Blockchain)
			utils.CheckError(err)

		}
	})

	http.HandleFunc("/peers", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			err := json.NewEncoder(w).Encode(s.Peers.Hostnames)
			utils.CheckError(err)
		}
	})

	http.HandleFunc("/notify", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)

			b, err := ioutil.ReadAll(r.Body)
			utils.CheckError(err)

			fmt.Println(string(b))

			t := HostName{}
			err = json.Unmarshal(b, &t)
			utils.CheckError(err)

			s.Peers.AddHostname(t.Hostname)

			err = json.NewEncoder(w).Encode(s.Peers.Hostnames)
			utils.CheckError(err)
		}
	})

	http.HandleFunc("/latest-block", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			latestBlock := s.BlockchainService.Blockchain.Blocks[len(s.BlockchainService.Blockchain.Blocks)-1]

			err := json.NewEncoder(w).Encode(latestBlock)
			utils.CheckError(err)
		}
	})

	// control endpoint
	http.HandleFunc("/create-transaction", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			w.Header().Set("Content-Type", "application/json")

			b, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				err = json.NewEncoder(w).Encode(err)
				return
			}

			createTransactionControl := CreateTransactionControl{}
			err = json.Unmarshal(b, &createTransactionControl)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				err = json.NewEncoder(w).Encode(err)
				return
			}

			transaction := s.BlockchainService.CreateTransaction([]byte(createTransactionControl.Address), createTransactionControl.Amount)

			s.Client.BroadcastTransaction(transaction)

			w.WriteHeader(http.StatusOK)
			err = json.NewEncoder(w).Encode(createTransactionControl)
			utils.CheckError(err)
		}
	})

	http.HandleFunc("/transaction", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			b, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				err = json.NewEncoder(w).Encode(err)
				return
			}

			t := wallet.Transaction{}
			err = json.Unmarshal(b, &t)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				err = json.NewEncoder(w).Encode(err)
				return
			}

			*s.BlockchainService.UnconfirmedTransactionPool = append(*s.BlockchainService.UnconfirmedTransactionPool, t)

			err = json.NewEncoder(w).Encode(t)
			utils.CheckError(err)
		}
	})

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
