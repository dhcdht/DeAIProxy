package main

import (
	"github.com/deapn/logger"
	"github.com/deapn/node/internal/tunnel"
)

func main() {
	log := logger.New("node")
	client := tunnel.NewClient("node-1", "ws://localhost:8080/ws/register?node_id=node-1", log)
	client.Start()
}
