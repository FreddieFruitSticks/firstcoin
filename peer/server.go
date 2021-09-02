package peer

import (
	"blockchain/coin"
	"blockchain/utils"
	"blockchain/wallet"
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
)

type Server struct {
	Peers           *Peers
	ThisPeer        string
	Client          *Client
	Account         *wallet.Account
	Blockchain      *coin.Blockchain
	TransactionPool *[]wallet.Transaction
	UTxOs           *map[string]wallet.UTxOut
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

func NewServer(p *Peers, c *Client, b *coin.Blockchain, a *wallet.Account, t string, uTxO *map[string]wallet.UTxOut, tp *[]wallet.Transaction) *Server {
	return &Server{Peers: p, Client: c, Blockchain: b, Account: a, ThisPeer: t, UTxOs: uTxO, TransactionPool: tp}
}

func (s *Server) HandleServer(port string) {
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "pong from, %q", html.EscapeString(r.URL.Path))
	})

	// control node
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

			block := s.Blockchain.GenerateNextBlock(s.TransactionPool)

			valid := block.IsValidBlock(s.Blockchain.GetLastBlock())
			if !valid {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				err = json.NewEncoder(w).Encode("Block not valid")
				return

			}

			s.Blockchain.AddBlock(block)
			s.Client.BroadcastBlock(block, s.ThisPeer)

			updateUnspentTxOutputs(s.TransactionPool, s.UTxOs)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)

			// byte arrays base64 encode and decode (on the other end). So the txOut address encodes as b64
			err = json.NewEncoder(w).Encode(s.Blockchain.Blocks)
		}
	})

	http.HandleFunc("/block", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)

			b, err := ioutil.ReadAll(r.Body)
			utils.CheckError(err)

			block := coin.Block{}
			err = json.Unmarshal(b, &block)
			utils.CheckError(err)

			if block.IsValidBlock(s.Blockchain.GetLastBlock()) && block.ValidTimestampToNow() {
				s.Blockchain.AddBlock(block)
				updateUnspentTxOutputs(&block.Transactions, s.UTxOs)
				// TODO: when this node receives a valid block, it must remove transactions from its own pool that exist in the blocks transactions data
			}
		}
	})

	http.HandleFunc("/block-chain", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			err := json.NewEncoder(w).Encode(s.Blockchain)
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
			latestBlock := s.Blockchain.Blocks[len(s.Blockchain.Blocks)-1]

			err := json.NewEncoder(w).Encode(latestBlock)
			utils.CheckError(err)
		}
	})

	// control node
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

			unspentTransaction := wallet.UTxOut{
				Address: []byte("1235asasdasda"),
				Amount:  1000,
			}

			transaction := wallet.CreateTransaction([]byte(createTransactionControl.Address), createTransactionControl.Amount, &unspentTransaction, s.Account)
			s.Client.BroadcastTransaction(transaction)
			*s.TransactionPool = append(*s.TransactionPool, transaction)

			fmt.Println((*s.TransactionPool))

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

			*s.TransactionPool = append(*s.TransactionPool, t)

			err = json.NewEncoder(w).Encode(t)
			utils.CheckError(err)
		}
	})

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

func updateUnspentTxOutputs(newTransactions *[]wallet.Transaction, unspentTxOuts *map[string]wallet.UTxOut) {
	// in here loop through transactions and update unspentTxOuts
}
