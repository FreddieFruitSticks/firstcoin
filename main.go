package main

import (
	"blockchain/coin"
	"blockchain/peer"
	"blockchain/service"
	"blockchain/wallet"
	"encoding/base64"
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

	var client *peer.Client

	blocks := make([]coin.Block, 0)
	blockchain := coin.NewBlockchain(blocks)

	thisPeer := fmt.Sprintf("localhost:%s", port)

	peers := peer.NewPeers()
	crypt := wallet.NewCryptographic()
	crypt.GenerateKeyPair()

	fmt.Println(string(base64Encode(crypt.PublicKey)))

	userWallet := wallet.NewWallet(*crypt)

	if isSeedHost(port) {
		*blockchain, _ = service.CreateGenesisBlockchain(*crypt, *blockchain)
		client = peer.NewClient(peers, blockchain, thisPeer)
	} else {
		client = peer.NewClient(peers, blockchain, thisPeer)

		p := client.GetPeers()
		err := client.QueryPeersForBlockchain(p)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(p)
		client.BroadcastOnline(thisPeer)
	}

	peers.AddHostname(thisPeer)

	service := service.NewBlockchainService(blockchain, userWallet)
	coinServerHandler := peer.NewCoinServerHandler(service, client, peers)

	server := peer.NewServer(*coinServerHandler)

	server.HandleServer(args[0])
}

func base64Encode(message []byte) []byte {
	b := make([]byte, base64.StdEncoding.EncodedLen(len(message)))
	base64.StdEncoding.Encode(b, message)
	return b
}
