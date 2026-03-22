package api

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"log/slog"
	"math/big"
	"net/http"
	"strconv"
	"sync"

	"github.com/deapn/protocol"
	"github.com/deapn/router/internal/registry"
	"github.com/deapn/router/internal/settlement"
	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
)

type Handler struct {
	reg     *registry.Registry
	log     *slog.Logger
	settler *settlement.Settler
	pending sync.Map 
	nodeMap sync.Map 
}

func NewHandler(reg *registry.Registry, settler *settlement.Settler, logger *slog.Logger) *Handler {
	return &Handler{
		reg:     reg,
		settler: settler,
		log:     logger,
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

	billAmountStr := r.Header.Get("X-DeAPN-Bill-Amount")
	billNonceStr := r.Header.Get("X-DeAPN-Bill-Nonce")
	billSigB64 := r.Header.Get("X-DeAPN-Bill-Sig")

	if billAmountStr != "" && billSigB64 != "" {
		amount, _ := new(big.Int).SetString(billAmountStr, 10)
		nonce, _ := strconv.ParseUint(billNonceStr, 10, 64)
		sig, _ := base64.StdEncoding.DecodeString(billSigB64)
		
		buyer := common.HexToAddress("0x0000000000000000000000000000000000000000") 
		seller := common.HexToAddress(node.WalletAddress)

		err := h.settler.ProcessBill(buyer, seller, amount, nonce, sig, 1337) 
		if err != nil {
			h.log.Error("Payment verification failed", "error", err)
			http.Error(w, "Payment required or invalid", http.StatusPaymentRequired)
			return
		}
	}

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
		NodeID:  nodeID,
		Payload: payload,
	}

	respChan := make(chan *protocol.ProxyResponsePayload, 100)
	h.pending.Store(reqID, respChan)
	h.nodeMap.Store(reqID, nodeID)
	defer h.pending.Delete(reqID)
	defer h.nodeMap.Delete(reqID)

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

		if len(resp.Body) > 0 {
			w.Write(resp.Body)
		}
		
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

	if payload.SignedChunk != nil && len(payload.Body) > 0 {
		nodeIDVal, ok := h.nodeMap.Load(msg.ID)
		if !ok {
			return
		}
		nodeID := nodeIDVal.(string)
		node := h.reg.GetNode(nodeID)
		if node != nil {
			valid, err := protocol.VerifyChunk(node.WalletAddress, payload.SignedChunk, payload.Body)
			if err != nil || !valid {
				h.log.Error("CRITICAL: Stream chunk signature verification failed", "id", msg.ID, "node", nodeID)
				return
			}
		}
	}

	if val, ok := h.pending.Load(msg.ID); ok {
		respChan := val.(chan *protocol.ProxyResponsePayload)
		respChan <- &payload
		if payload.IsLast {
			close(respChan)
		}
	}
}
