package main

import (
	"crypto/ecdsa"
	"flag"
	"fmt"

	"github.com/deapn/logger"
	"github.com/deapn/node/internal/tunnel"
	"github.com/deapn/protocol"
	"github.com/ethereum/go-ethereum/crypto"
)

func main() {
	nodeID := flag.String("id", "node-1", "Node ID")
	router := flag.String("router", "ws://localhost:8080/ws/register", "Router URL")
	privateKeyHex := flag.String("key", "", "ECDSA Private Key Hex")
	flag.Parse()

	log := logger.New("node")

	var privKey *ecdsa.PrivateKey
	var wallet string
	var err error

	if *privateKeyHex != "" {
		privKey, err = crypto.HexToECDSA(*privateKeyHex)
		if err != nil {
			log.Error("Invalid private key", "error", err)
			return
		}
		wallet = crypto.PubkeyToAddress(privKey.PublicKey).Hex()
	} else {
		log.Warn("No private key provided, node will register with empty identity")
		wallet = "0x0000000000000000000000000000000000000000"
	}

	prover := protocol.NewStubTLSProver()
	log.Info("Starting DeAPN Node", "id", *nodeID, "wallet", wallet)

	routerURL := fmt.Sprintf("%s?node_id=%s", *router, *nodeID)
	client := tunnel.NewClient(*nodeID, wallet, routerURL, privKey, prover, log)
	client.Start()
}
