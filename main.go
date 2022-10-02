package main

import (
	"firstcoin/coin"
	"firstcoin/peer"
	"firstcoin/repository"
	"firstcoin/service"
	"firstcoin/utils"
	"firstcoin/wallet"
	"fmt"
	"os"
)

const seedHost = "firstcoin-node1:8080"

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
	hostname := os.Getenv("HOST_NAME")
	thisPeer := fmt.Sprintf("%s:%s", hostname, port)
	fmt.Printf("This peer: %s\n", thisPeer)

	peers := peer.NewPeers()
	peers.ThisHost = thisPeer

	crypt := wallet.NewCryptographic()
	crypt.GenerateKeyPair()

	fmt.Printf("Address of this node: %s\n", string(string(crypt.FirstcoinAddress)))
	fmt.Printf("Address of this node: %s\n", string(repository.Base64Encode(crypt.FirstcoinAddress)))

	userWallet := wallet.NewWallet(*crypt)

	if isSeedHost(port) {
		*blockchain, _ = service.CreateGenesisBlockchain(*crypt, *blockchain)
		client = peer.NewClient(peers, blockchain, thisPeer)
	} else {
		if len(args) > 1 {
			specificPeerToConnectTo := args[1]
			peers.Hostnames[specificPeerToConnectTo] = specificPeerToConnectTo

			client = peer.NewClient(peers, blockchain, thisPeer)
		} else {
			client = peer.NewClient(peers, blockchain, thisPeer)
			newPeers := client.GetPeers(seedHost)
			peers.Hostnames = newPeers
		}

		err := client.QueryPeersForBlockchain(client.Peers.Hostnames)
		if err != nil {
			fmt.Println(err)
			return
		}

		err = client.QueryNetworkForUnconfirmedTxPool(client.Peers.Hostnames)
		if err != nil {
			return
		}

		utils.InfoLogger.Println(client.Peers.Hostnames)
		client.BroadcastOnline(thisPeer)
	}

	// peers.AddHostname(thisPeer)

	service := service.NewBlockchainService(blockchain, userWallet)
	coinServerHandler := peer.NewCoinServerHandler(service, client, peers)

	server := peer.NewServer(*coinServerHandler)

	server.HandleServer(args[0], port == "8080")
}
