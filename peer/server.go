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

// TODO remove all non-server related stuff to a new package - need refactor
type Server struct {
	Peers                      *Peers
	ThisPeer                   string
	Client                     *Client
	Account                    *wallet.Account
	Blockchain                 *coin.Blockchain
	UnconfirmedTransactionPool *[]wallet.Transaction
	UTxOs                      *map[string]map[string]wallet.UTxOut
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

func NewServer(p *Peers, c *Client, b *coin.Blockchain, a *wallet.Account, t string, uTxO *map[string]map[string]wallet.UTxOut, tp *[]wallet.Transaction) *Server {
	return &Server{Peers: p, Client: c, Blockchain: b, Account: a, ThisPeer: t, UTxOs: uTxO, UnconfirmedTransactionPool: tp}
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

			// coinbase transaction is the first transaction included by the miner
			coinbaseTransaction := wallet.CreateCoinbaseTransaction(*s.Account, s.Blockchain.GetLastBlock().Index+1)
			transactionPool := make([]wallet.Transaction, 0)
			transactionPool = append(transactionPool, coinbaseTransaction)
			transactionPool = append(transactionPool, *s.UnconfirmedTransactionPool...)

			block := s.Blockchain.GenerateNextBlock(&transactionPool)

			valid := block.IsValidBlock(s.Blockchain.GetLastBlock())
			if !valid {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				err = json.NewEncoder(w).Encode("Block not valid")

				return
			}

			s.Blockchain.AddBlock(block)
			s.Client.BroadcastBlock(block, s.ThisPeer)

			s.UpdateUnspentTxOutputs(block)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)

			payload := struct {
				Blocks              []coin.Block                        `json:"blocks"`
				UnspentTransactions map[string]map[string]wallet.UTxOut `json:"unspentTransactions"`
			}{
				Blocks:              s.Blockchain.Blocks,
				UnspentTransactions: *s.UTxOs,
			}

			// byte arrays base64 encode and decode (on the other end). So the txOut address encodes as b64
			err = json.NewEncoder(w).Encode(payload)
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
				s.UpdateUnspentTxOutputs(block)
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
			*s.UnconfirmedTransactionPool = append(*s.UnconfirmedTransactionPool, transaction)

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

			*s.UnconfirmedTransactionPool = append(*s.UnconfirmedTransactionPool, t)

			err = json.NewEncoder(w).Encode(t)
			utils.CheckError(err)
		}
	})

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

func (s *Server) UpdateUnspentTxOutputs(block coin.Block) {
	// At this point assume each transaction is valid - checked previously

	s.UpdateUTxOWithCoinbaseTransaction(block)

	for _, _ = range (block.Transactions)[1:] {
	}
	// in here loop through transactions and update unspentTxOuts
}

func (s *Server) UpdateUTxOWithCoinbaseTransaction(block coin.Block) bool {
	coinbaseTx := (block.Transactions)[0]

	txIDMap := make(map[string]wallet.UTxOut)
	txIDMap[string(coinbaseTx.ID)] = wallet.UTxOut{
		ID:      coinbaseTx.ID,
		Index:   block.Index,
		Address: coinbaseTx.TxOuts[0].Address,
		Amount:  coinbaseTx.TxOuts[0].Amount,
	}

	(*s.UTxOs)[string(coinbaseTx.TxOuts[0].Address)] = txIDMap

	return true
}
