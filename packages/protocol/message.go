package protocol

import "encoding/json"

type MessageType string

const (
	MessageTypeRegister MessageType = "Register"
	MessageTypePing     MessageType = "Ping"
	MessageTypePong     MessageType = "Pong"
	MessageTypeProxyReq MessageType = "ProxyRequest"
	MessageTypeProxyRes MessageType = "ProxyResponse"
)

type Message struct {
	Type    MessageType     `json:"type"`
	ID      string          `json:"id,omitempty"`
	NodeID  string          `json:"node_id,omitempty"`
	Payload json.RawMessage `json:"payload"`
}

type RegisterPayload struct {
	NodeID        string `json:"node_id"`
	WalletAddress string `json:"wallet_address"`
}

type ProxyRequestPayload struct {
	Method string              `json:"method"`
	URL    string              `json:"url"`
	Header map[string][]string `json:"header"`
	Body   []byte              `json:"body"`
}

type ProxyResponsePayload struct {
	StatusCode int                 `json:"status_code"`
	Header     map[string][]string `json:"header"`
	Body       []byte              `json:"body"`
	IsLast     bool                `json:"is_last"`
}
