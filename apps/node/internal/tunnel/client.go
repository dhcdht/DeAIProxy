package tunnel

import (
	"crypto/ecdsa"
	"encoding/json"
	"log/slog"
	"net/url"
	"time"

	"github.com/deapn/node/internal/forwarder"
	"github.com/deapn/protocol"
	"github.com/gorilla/websocket"
)

type Client struct {
	nodeID        string
	walletAddress string
	routerURL     string
	privKey       *ecdsa.PrivateKey
	prover        protocol.TLSProver
	log           *slog.Logger
	conn          *websocket.Conn
}

func NewClient(nodeID, wallet, routerURL string, privKey *ecdsa.PrivateKey, prover protocol.TLSProver, logger *slog.Logger) *Client {
	return &Client{
		nodeID:        nodeID,
		walletAddress: wallet,
		routerURL:     routerURL,
		privKey:       privKey,
		prover:        prover,
		log:           logger,
	}
}

func (c *Client) Start() {
	u, err := url.Parse(c.routerURL)
	if err != nil {
		c.log.Error("Invalid router URL", "error", err)
		return
	}

	for {
		c.log.Info("Connecting to router...", "url", u.String())
		conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			c.log.Error("Failed to connect", "error", err)
			time.Sleep(5 * time.Second)
			continue
		}

		c.conn = conn
		if err := c.register(); err != nil {
			c.log.Error("Failed to register", "error", err)
			conn.Close()
			time.Sleep(5 * time.Second)
			continue
		}

		c.log.Info("Connected and registered", "node_id", c.nodeID, "wallet", c.walletAddress)
		
		fwd := forwarder.New(c.conn, c.log, c.privKey, c.prover)
		c.listen(fwd)
		
		c.log.Warn("Disconnected, retrying...")
		time.Sleep(5 * time.Second)
	}
}

func (c *Client) register() error {
	payload, _ := json.Marshal(protocol.RegisterPayload{
		NodeID:        c.nodeID,
		WalletAddress: c.walletAddress,
	})
	msg := protocol.Message{
		Type:    protocol.MessageTypeRegister,
		NodeID:  c.nodeID,
		Payload: payload,
	}
	return c.conn.WriteJSON(msg)
}

func (c *Client) listen(fwd *forwarder.Forwarder) {
	defer c.conn.Close()
	for {
		var msg protocol.Message
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			c.log.Error("Read error", "error", err)
			break
		}

		switch msg.Type {
		case protocol.MessageTypeProxyReq:
			go fwd.Forward(msg.ID, msg.Payload)
		case protocol.MessageTypeDisputeChallenge:
			go c.handleDispute(msg)
		}
	}
}

func (c *Client) handleDispute(msg protocol.Message) {
	var payload protocol.DisputeChallengePayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return
	}

	c.log.Info("Received dispute challenge", "req_id", payload.RequestID)
	proof, err := c.prover.GenerateProof(payload.RequestID)
	if err != nil {
		c.log.Error("Failed to generate proof for dispute", "error", err)
		return
	}

	resPayload, _ := json.Marshal(protocol.DisputeProofPayload{
		RequestID: payload.RequestID,
		Proof:     proof,
	})
	resMsg := protocol.Message{
		Type:    protocol.MessageTypeDisputeProof,
		ID:      msg.ID,
		NodeID:  c.nodeID,
		Payload: resPayload,
	}
	c.conn.WriteJSON(resMsg)
}
