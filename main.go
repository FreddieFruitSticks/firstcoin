package main

import (
	"blockchain/coin"
	"blockchain/peer"
	"blockchain/service"
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

	var client *peer.Client

	unspentTxOuts := make(map[string]map[string]wallet.UTxOut, 0)
	blocks := make([]coin.Block, 0)
	blockchain := coin.NewBlockchain(blocks)

	thisPeer := fmt.Sprintf("localhost:%s", port)
	peers := peer.NewPeers()
	account := wallet.NewAccount()
	account.GenerateKeyPair()
	transactionPool := make([]wallet.Transaction, 0)

	if isSeedHost(port) {
		*blockchain = createGenesisBlockchain(unspentTxOuts, *account, *blockchain)
		client = peer.NewClient(peers, blockchain, thisPeer)
	} else {
		client = peer.NewClient(peers, blockchain, thisPeer)

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

	service := service.NewBlockchainService(account, blockchain, &transactionPool, &unspentTxOuts)

	server := peer.NewServer(service, peers, client)

	server.HandleServer(args[0])
}

func createGenesisBlockchain(unspentTxOuts map[string]map[string]wallet.UTxOut, account wallet.Account, blockchain coin.Blockchain) coin.Blockchain {
	genesisTransactionPool := make([]wallet.Transaction, 0)

	// coinbase transaction is the first transaction included by the miner
	coinbaseTransaction := wallet.CreateCoinbaseTransaction(account, 0)
	genesisTransactionPool = append(genesisTransactionPool, coinbaseTransaction)

	uTxO := wallet.UTxOut{
		ID:      coinbaseTransaction.ID,
		Index:   0,
		Address: coinbaseTransaction.TxOuts[0].Address,
		Amount:  coinbaseTransaction.TxOuts[0].Amount,
	}
	txIDMap := make(map[string]wallet.UTxOut)
	txIDMap[string(coinbaseTransaction.ID)] = uTxO
	unspentTxOuts[string(account.PublicKey)] = txIDMap

	blockchain.AddBlock(coin.GenesisBlock(seedDifficultyLevel, genesisTransactionPool))

	return blockchain
}
