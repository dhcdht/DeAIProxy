package main

import (
	"flag"
	"fmt"

	"github.com/deapn/logger"
	"github.com/deapn/node/internal/tunnel"
)

func main() {
	nodeID := flag.String("id", "node-1", "Node ID")
	wallet := flag.String("wallet", "0x0000000000000000000000000000000000000000", "Wallet Address for Staking")
	router := flag.String("router", "ws://localhost:8080/ws/register", "Router URL")
	flag.Parse()

	log := logger.New("node")
	log.Info("Starting DeAPN Node", "id", *nodeID, "wallet", *wallet)

	routerURL := fmt.Sprintf("%s?node_id=%s", *router, *nodeID)
	client := tunnel.NewClient(*nodeID, *wallet, routerURL, log)
	client.Start()
}
