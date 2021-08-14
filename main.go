package main

import (
	"blockchain/coin"
	"blockchain/peer"
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

func main() {
	args := os.Args[1:]
	port := args[0]

	blockchain := coin.NewBlockchain()

	thisPeer := fmt.Sprintf("localhost:%s", port)
	peers := peer.NewPeers()
	client := peer.NewClient(peers, blockchain)

	if isSeedHost(port) {
		blockchain.AddBlock(coin.GenesisBlock())
	} else {
		p := client.GetPeers()
		client.QueryPeers(p)

		fmt.Println(p)
		client.BroadcastOnline(thisPeer)
	}

	peers.AddHostname(thisPeer)

	server := peer.NewServer(peers, client, blockchain, thisPeer)

	server.HandleServer(args[0])
}
