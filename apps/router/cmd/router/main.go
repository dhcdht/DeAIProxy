package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/deapn/logger"
	"github.com/deapn/protocol"
	"github.com/deapn/router/internal/api"
	"github.com/deapn/router/internal/dispute"
	"github.com/deapn/router/internal/eth"
	"github.com/deapn/router/internal/registry"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func main() {
	log := logger.New("router")

	rpcURL := "http://localhost:8545"
	stakingAddr := "0x5FbDB2315678afecb367f032d93F642f64180aa3" 
	ethClient, err := eth.New(rpcURL, stakingAddr)
	if err != nil {
		log.Warn("Failed to initialize eth client, starting without on-chain verification", "error", err)
	}

	reg := registry.New(log, ethClient)
	handler := api.NewHandler(reg, log)
	
	verifier := &protocol.StubTLSVerifier{}
	arbiter := dispute.NewArbiter(reg, verifier, log)

	http.HandleFunc("/v1/", handler.HandleProxy)

	http.HandleFunc("/v1/report_fraud", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			NodeID       string `json:"node_id"`
			RequestID    string `json:"request_id"`
			ExpectedData []byte `json:"expected_data"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 180*time.Second)
		defer cancel()

		valid, err := arbiter.Dispute(ctx, req.NodeID, req.RequestID, req.ExpectedData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if valid {
			w.Write([]byte(`{"status": "ok", "message": "Node proved innocent"}`))
		} else {
			w.Write([]byte(`{"status": "fraud", "message": "Node failed proof - SLASHING TRIGGERED"}`))
		}
	})

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

		if err := reg.Register(payload.NodeID, payload.WalletAddress, conn); err != nil {
			log.Error("Registration rejected", "error", err)
			conn.Close()
			return
		}
		defer reg.Unregister(payload.NodeID)

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Warn("Node disconnected", "id", payload.NodeID, "error", err)
				break
			}
			var nodeMsg protocol.Message
			if err := json.Unmarshal(message, &nodeMsg); err == nil {
				switch nodeMsg.Type {
				case protocol.MessageTypeProxyRes:
					handler.HandleResponse(nodeMsg)
				case protocol.MessageTypeDisputeProof:
					arbiter.HandleProof(nodeMsg)
				}
			}
		}
	})

	log.Info("Router starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Error("Router failed", "error", err)
	}
}
