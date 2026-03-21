package protocol

import "encoding/json"

type MessageType string

const (
	MessageTypeRegister    MessageType = "Register"
	MessageTypePing        MessageType = "Ping"
	MessageTypePong        MessageType = "Pong"
	MessageTypeProxyReq    MessageType = "ProxyRequest"
	MessageTypeProxyRes    MessageType = "ProxyResponse"
	MessageTypeSignedChunk MessageType = "SignedChunk"
)

type Message struct {
	Type      MessageType     `json:"type"`
	ID        string          `json:"id,omitempty"` // Correlation ID
	NodeID    string          `json:"node_id,omitempty"`
	Signature []byte          `json:"signature,omitempty"`
	Payload   json.RawMessage `json:"payload"`
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
	StatusCode  int                 `json:"status_code"`
	Header      map[string][]string `json:"header"`
	Body        []byte              `json:"body"`
	IsLast      bool                `json:"is_last"`
	SignedChunk *SignedChunk        `json:"signed_chunk,omitempty"`
}

type SignedChunk struct {
	RequestID string `json:"request_id"`
	SeqNum    uint64 `json:"seq_num"`
	DataHash  []byte `json:"data_hash"`
	Timestamp int64  `json:"timestamp"`
	Signature []byte `json:"signature"`
}
