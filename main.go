package main

import (
	"blockchain/coin"
	"blockchain/peer"
	"blockchain/repository"
	"blockchain/service"
	"blockchain/utils"
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

// func main() {
// 	m := make(map[string]map[string]Person)
// 	p1 := make(map[string]Person)
// 	p := Person{
// 		height: 2,
// 	}
// 	p1["1"] = p
// 	m["1"] = p1

// 	newMap := make(map[string]map[string]Person)
// 	for k, v := range m {
// 		newMap[k] = v
// 	}

// 	myFunc(newMap)

// 	fmt.Println("out func")

// 	fmt.Println(m)
// }

// type Person struct {
// 	height int
// }

// func myFunc(person map[string]map[string]Person) {
// 	p1 := make(map[string]Person)
// 	p := Person{
// 		height: 3,
// 	}
// 	p1["1"] = p

// 	person["1"] = p1
// 	fmt.Println("in func")
// 	fmt.Println(person)
// }

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

	fmt.Println(string(repository.Base64Encode(crypt.PublicKey)))

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
		err = client.QueryNetworkForUnconfirmedTxPool(p)
		if err != nil {
			return
		}

		utils.InfoLogger.Println(p)
		client.BroadcastOnline(thisPeer)
	}

	peers.AddHostname(thisPeer)

	service := service.NewBlockchainService(blockchain, userWallet)
	coinServerHandler := peer.NewCoinServerHandler(service, client, peers)

	server := peer.NewServer(*coinServerHandler)

	server.HandleServer(args[0])
}
