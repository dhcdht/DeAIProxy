package forwarder

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/deapn/protocol"
	"github.com/gorilla/websocket"
)

type Forwarder struct {
	client *http.Client
	conn   *websocket.Conn
	log    *slog.Logger
}

func New(conn *websocket.Conn, logger *slog.Logger) *Forwarder {
	return &Forwarder{
		client: &http.Client{},
		conn:   conn,
		log:    logger,
	}
}

func (f *Forwarder) Forward(reqID string, payloadBytes []byte) {
	var payload protocol.ProxyRequestPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		f.log.Error("Failed to unmarshal request payload", "error", err)
		return
	}

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
		return
	}
	defer resp.Body.Close()

	buffer := make([]byte, 4096)
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			f.sendResponse(reqID, resp, buffer[:n], false)
		}
		if err == io.EOF {
			f.sendResponse(reqID, resp, nil, true)
			break
		}
		if err != nil {
			break
		}
	}
}

func (f *Forwarder) sendResponse(reqID string, resp *http.Response, body []byte, isLast bool) {
	payload := protocol.ProxyResponsePayload{
		StatusCode: resp.StatusCode,
		Header:     resp.Header,
		Body:       body,
		IsLast:     isLast,
	}
	payloadBytes, _ := json.Marshal(payload)
	msg := protocol.Message{
		Type:    protocol.MessageTypeProxyRes,
		ID:      reqID,
		Payload: payloadBytes,
	}
	f.conn.WriteJSON(msg)
}
