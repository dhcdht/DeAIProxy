package tunnel

import (
	"encoding/json"
	"log/slog"
	"net/url"
	"time"

	"github.com/deapn/protocol"
	"github.com/gorilla/websocket"
)

type Client struct {
	nodeID    string
	routerURL string
	log       *slog.Logger
	conn      *websocket.Conn
}

func NewClient(nodeID, routerURL string, logger *slog.Logger) *Client {
	return &Client{
		nodeID:    nodeID,
		routerURL: routerURL,
		log:       logger,
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

		c.log.Info("Connected and registered")
		c.listen()
		c.log.Warn("Disconnected, retrying...")
		time.Sleep(5 * time.Second)
	}
}

func (c *Client) register() error {
	payload, _ := json.Marshal(protocol.RegisterPayload{NodeID: c.nodeID})
	msg := protocol.Message{
		Type:    protocol.MessageTypeRegister,
		NodeID:  c.nodeID,
		Payload: payload,
	}
	return c.conn.WriteJSON(msg)
}

func (c *Client) listen() {
	defer c.conn.Close()
	for {
		var msg protocol.Message
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			c.log.Error("Read error", "error", err)
			break
		}
		c.log.Debug("Received message", "type", msg.Type)
	}
}
