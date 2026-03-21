package api

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"sync"

	"github.com/deapn/protocol"
	"github.com/deapn/router/internal/registry"
	"github.com/google/uuid"
)

type Handler struct {
	reg     *registry.Registry
	log     *slog.Logger
	pending sync.Map
}

func NewHandler(reg *registry.Registry, logger *slog.Logger) *Handler {
	return &Handler{
		reg: reg,
		log: logger,
	}
}

func (h *Handler) HandleProxy(w http.ResponseWriter, r *http.Request) {
	nodes := h.reg.ListNodes()
	if len(nodes) == 0 {
		http.Error(w, "No active nodes available", http.StatusServiceUnavailable)
		return
	}
	nodeID := nodes[0]
	node := h.reg.GetNode(nodeID)

	body, _ := io.ReadAll(r.Body)
	reqID := uuid.New().String()
	payload, _ := json.Marshal(protocol.ProxyRequestPayload{
		Method: r.Method,
		URL:    "https://api.openai.com" + r.URL.Path,
		Header: r.Header,
		Body:   body,
	})

	msg := protocol.Message{
		Type:    protocol.MessageTypeProxyReq,
		ID:      reqID,
		Payload: payload,
	}

	respChan := make(chan *protocol.ProxyResponsePayload, 100)
	h.pending.Store(reqID, respChan)
	defer h.pending.Delete(reqID)

	if err := node.Conn.WriteJSON(msg); err != nil {
		h.log.Error("Failed to send proxy request", "error", err)
		http.Error(w, "Node communication error", http.StatusInternalServerError)
		return
	}

	first := true
	for resp := range respChan {
		if first {
			for k, vs := range resp.Header {
				for _, v := range vs {
					w.Header().Add(k, v)
				}
			}
			w.WriteHeader(resp.StatusCode)
			first = false
		}

		w.Write(resp.Body)
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}

		if resp.IsLast {
			break
		}
	}
}

func (h *Handler) HandleResponse(msg protocol.Message) {
	var payload protocol.ProxyResponsePayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		h.log.Error("Failed to unmarshal proxy response", "error", err)
		return
	}

	if val, ok := h.pending.Load(msg.ID); ok {
		respChan := val.(chan *protocol.ProxyResponsePayload)
		respChan <- &payload
		if payload.IsLast {
			close(respChan)
		}
	}
}
