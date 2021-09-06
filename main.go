package main

import (
	"blockchain/coin"
	"blockchain/peer"
	"blockchain/wallet"
	"fmt"
	"os"
)

// For now seed host is identified as being on port 8080
func isSeedHost(port string) bool {
	if port == "8080" {
		return true
	}

	return false
}

const (
	seedDifficultyLevel = 5
)

func main() {
	args := os.Args[1:]
	port := args[0]

	unspentTxOuts := make(map[string]map[string]wallet.UTxOut, 0)
	blocks := make([]coin.Block, 0)
	blockchain := coin.NewBlockchain(blocks)

	thisPeer := fmt.Sprintf("localhost:%s", port)
	peers := peer.NewPeers()
	client := peer.NewClient(peers, blockchain, thisPeer)
	account := wallet.NewAccount()
	account.GenerateKeyPair()
	transactionPool := make([]wallet.Transaction, 0)

	if isSeedHost(port) {
		genesisTransactionPool := make([]wallet.Transaction, 0)

		// coinbase transaction is the first transaction included by the miner
		coinbaseTransaction := wallet.CreateCoinbaseTransaction(*account, 0)
		genesisTransactionPool = append(genesisTransactionPool, coinbaseTransaction)

		blockchain.AddBlock(coin.GenesisBlock(seedDifficultyLevel, genesisTransactionPool))
	} else {
		p := client.GetPeers()
		err := client.QueryPeers(p)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(p)
		client.BroadcastOnline(thisPeer)
	}

	peers.AddHostname(thisPeer)

	server := peer.NewServer(peers, client, blockchain, account, thisPeer, &unspentTxOuts, &transactionPool)

	server.HandleServer(args[0])
}
