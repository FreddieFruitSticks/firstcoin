package peer

import (
	"blockchain/coin"
	"blockchain/utils"
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
)

type Server struct {
	Peers      *Peers
	ThisPeer   string
	Client     *Client
	Account    *coin.Account
	Blockchain coin.Blockchain
}

type blockData struct {
	Data string `json:"data"`
}

func NewServer(p *Peers, c *Client, b coin.Blockchain, t string) *Server {
	return &Server{Peers: p, Client: c, Blockchain: b, ThisPeer: t}
}

func (s *Server) HandleServer(port string) {
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "pong from, %q", html.EscapeString(r.URL.Path))
	})

	http.HandleFunc("/create-block", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			b, err := ioutil.ReadAll(r.Body)
			utils.CheckError(err)

			data := blockData{}
			err = json.Unmarshal(b, &data)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				err = json.NewEncoder(w).Encode(err)
				return
			}

			block := s.Blockchain.GenerateNextBlock(data.Data)

			valid := block.IsValidBlock(s.Blockchain.GetLastBlock())
			if !valid {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				err = json.NewEncoder(w).Encode("Block not valid")
				return

			}

			s.Blockchain.AddBlock(block)
			s.Client.BroadcastBlock(block, s.ThisPeer)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
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

			fmt.Println(string(b))

			data := coin.Block{}
			err = json.Unmarshal(b, &data)
			utils.CheckError(err)
			fmt.Println(data.IsValidBlock(s.Blockchain.GetLastBlock()))

			if data.IsValidBlock(s.Blockchain.GetLastBlock()) {
				s.Blockchain.AddBlock(data)
			}
		}
	})

	http.HandleFunc("/block-chain", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			err := json.NewEncoder(w).Encode(s.Blockchain.Blocks)
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

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
