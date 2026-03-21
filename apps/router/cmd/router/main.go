package main

import (
	"encoding/json"
	"net/http"

	"github.com/deapn/logger"
	"github.com/deapn/protocol"
	"github.com/deapn/router/internal/api"
	"github.com/deapn/router/internal/registry"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func main() {
	log := logger.New("router")
	reg := registry.New(log)
	handler := api.NewHandler(reg, log)

	http.HandleFunc("/v1/", handler.HandleProxy)

	http.HandleFunc("/ws/register", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Error("Upgrade error", "error", err)
			return
		}

		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Error("Failed to read register message", "error", err)
			conn.Close()
			return
		}

		var msg protocol.Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Error("Failed to unmarshal register message", "error", err)
			conn.Close()
			return
		}

		var payload protocol.RegisterPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			log.Error("Failed to unmarshal register payload", "error", err)
			conn.Close()
			return
		}

		reg.Register(payload.NodeID, payload.WalletAddress, conn)
		defer reg.Unregister(payload.NodeID)

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Warn("Node disconnected", "id", payload.NodeID, "error", err)
				break
			}
			var nodeMsg protocol.Message
			if err := json.Unmarshal(message, &nodeMsg); err == nil {
				if nodeMsg.Type == protocol.MessageTypeProxyRes {
					handler.HandleResponse(nodeMsg)
				}
			}
		}
	})

	log.Info("Router starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Error("Router failed", "error", err)
	}
}
