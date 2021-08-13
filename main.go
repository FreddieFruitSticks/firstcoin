package main

import (
	"blockchain/coin"
	"blockchain/peer"
	"fmt"
	"os"
)

var GenesisBlock = coin.GenesisBlock()

func main() {
	args := os.Args[1:]

	blockchain := coin.NewBlockchain()
	blockchain.AddBlock(GenesisBlock)

	thisPeer := fmt.Sprintf("localhost:%s", args[0])

	peers := peer.NewPeers()
	client := peer.NewClient(peers)

	if args[0] != "8080" {
		p := client.GetPeers()
		fmt.Println(p)
		client.BroadcastOnline(thisPeer)
	}

	peers.AddHostname(thisPeer)

	server := peer.NewServer(peers, client, blockchain, thisPeer)

	server.HandleServer(args[0])
}
