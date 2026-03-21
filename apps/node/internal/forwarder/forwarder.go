package forwarder

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"sync/atomic"

	"github.com/deapn/protocol"
	"github.com/gorilla/websocket"
)

type Forwarder struct {
	client  *http.Client
	conn    *websocket.Conn
	log     *slog.Logger
	privKey *ecdsa.PrivateKey
	counter uint64
}

func New(conn *websocket.Conn, logger *slog.Logger, privKey *ecdsa.PrivateKey) *Forwarder {
	return &Forwarder{
		client:  &http.Client{},
		conn:    conn,
		log:     logger,
		privKey: privKey,
	}
}

func (f *Forwarder) Forward(reqID string, payloadBytes []byte) {
	var payload protocol.ProxyRequestPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		f.log.Error("Failed to unmarshal request payload", "error", err)
		return
	}

	f.log.Info("Forwarding request", "url", payload.URL, "method", payload.Method)

	req, err := http.NewRequest(payload.Method, payload.URL, bytes.NewReader(payload.Body))
	if err != nil {
		f.log.Error("Failed to create request", "error", err)
		return
	}

	for k, vs := range payload.Header {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}

	resp, err := f.client.Do(req)
	if err != nil {
		f.sendError(reqID, err.Error())
		return
	}
	defer resp.Body.Close()

	buffer := make([]byte, 4096)
	seqNum := atomic.AddUint64(&f.counter, 1)
	
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			f.sendResponse(reqID, resp, buffer[:n], seqNum, false)
			seqNum++
		}
		if err == io.EOF {
			f.sendResponse(reqID, resp, nil, seqNum, true)
			break
		}
		if err != nil {
			f.log.Error("Read error", "error", err)
			break
		}
	}
}

func (f *Forwarder) sendResponse(reqID string, resp *http.Response, body []byte, seqNum uint64, isLast bool) {
	var signedChunk *protocol.SignedChunk
	if f.privKey != nil && len(body) > 0 {
		var err error
		signedChunk, err = protocol.SignChunk(f.privKey, reqID, seqNum, body)
		if err != nil {
			f.log.Error("Failed to sign chunk", "error", err)
		}
	}

	payload := protocol.ProxyResponsePayload{
		StatusCode:  resp.StatusCode,
		Header:      resp.Header,
		Body:        body,
		IsLast:      isLast,
		SignedChunk: signedChunk,
	}
	payloadBytes, _ := json.Marshal(payload)
	msg := protocol.Message{
		Type:    protocol.MessageTypeProxyRes,
		ID:      reqID,
		Payload: payloadBytes,
	}
	f.conn.WriteJSON(msg)
}

func (f *Forwarder) sendError(reqID string, errMsg string) {
	payload := protocol.ProxyResponsePayload{
		StatusCode: http.StatusBadGateway,
		Body:       []byte(errMsg),
		IsLast:     true,
	}
	payloadBytes, _ := json.Marshal(payload)
	msg := protocol.Message{
		Type:    protocol.MessageTypeProxyRes,
		ID:      reqID,
		Payload: payloadBytes,
	}
	f.conn.WriteJSON(msg)
}
