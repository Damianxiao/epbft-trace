package main

import (
	"os"
	"pbft-pra/network"
)

func main() {
	nodeId := os.Args[1]
	server := network.NewServer(nodeId)
	server.Start()
}
