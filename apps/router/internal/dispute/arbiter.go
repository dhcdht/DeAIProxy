package dispute

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	"github.com/deapn/protocol"
	"github.com/deapn/router/internal/registry"
)

type Arbiter struct {
	reg      *registry.Registry
	log      *slog.Logger
	verifier protocol.TLSVerifier
	pending  sync.Map
}

func NewArbiter(reg *registry.Registry, verifier protocol.TLSVerifier, logger *slog.Logger) *Arbiter {
	return &Arbiter{
		reg:      reg,
		log:      logger,
		verifier: verifier,
	}
}

func (a *Arbiter) Dispute(ctx context.Context, nodeID string, reqID string, expectedData []byte) (bool, error) {
	node := a.reg.GetNode(nodeID)
	if node == nil {
		return false, fmt.Errorf("node %s not found", nodeID)
	}

	payload, _ := json.Marshal(protocol.DisputeChallengePayload{RequestID: reqID})
	msg := protocol.Message{
		Type:    protocol.MessageTypeDisputeChallenge,
		ID:      reqID,
		Payload: payload,
	}

	proofChan := make(chan []byte, 1)
	a.pending.Store(reqID, proofChan)
	defer a.pending.Delete(reqID)

	if err := node.Conn.WriteJSON(msg); err != nil {
		return false, fmt.Errorf("failed to send challenge: %w", err)
	}

	select {
	case proof := <-proofChan:
		return a.verifier.VerifyProof(proof, expectedData)
	case <-ctx.Done():
		return false, fmt.Errorf("dispute timed out")
	}
}

func (a *Arbiter) HandleProof(msg protocol.Message) {
	var payload protocol.DisputeProofPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return
	}

	if val, ok := a.pending.Load(payload.RequestID); ok {
		ch := val.(chan []byte)
		ch <- payload.Proof
	}
}
