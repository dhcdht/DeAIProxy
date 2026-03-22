package registry

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/deapn/router/internal/eth"
	"github.com/gorilla/websocket"
)

type HealthMetrics struct {
	TTFT        time.Duration
	TPS         float64
	DropRate    float64
	SuccessRate float64
	UpdatedAt   time.Time
}

type Node struct {
	ID            string
	WalletAddress string
	Conn          *websocket.Conn
	IsActive      bool
	mu            sync.Mutex
}

type Registry struct {
	nodes   map[string]*Node
	metrics map[string]HealthMetrics
	mu      sync.RWMutex
	log     *slog.Logger
	eth     *eth.Client
}

func New(logger *slog.Logger, ethClient *eth.Client) *Registry {
	return &Registry{
		nodes:   make(map[string]*Node),
		metrics: make(map[string]HealthMetrics),
		log:     logger,
		eth:     ethClient,
	}
}

func (r *Registry) Register(id string, wallet string, conn *websocket.Conn) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.eth != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		eligible, err := r.eth.IsNodeEligible(ctx, wallet)
		if err != nil {
			r.log.Error("Failed to check node eligibility on-chain", "wallet", wallet, "error", err)
			return fmt.Errorf("staking verification failed: %w", err)
		}
		if !eligible {
			r.log.Warn("Node registration rejected: Insufficient Stake", "wallet", wallet)
			return fmt.Errorf("insufficient stake on-chain")
		}
	}

	if old, ok := r.nodes[id]; ok {
		old.Conn.Close()
	}

	r.nodes[id] = &Node{
		ID:            id,
		WalletAddress: wallet,
		Conn:          conn,
		IsActive:      true,
	}
	r.log.Info("Node registered", "id", id, "wallet", wallet)
	return nil
}

func (r *Registry) Unregister(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if node, ok := r.nodes[id]; ok {
		node.Conn.Close()
		delete(r.nodes, id)
		r.log.Info("Node unregistered", "id", id)
	}
}

func (r *Registry) GetNode(id string) *Node {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.nodes[id]
}

func (r *Registry) ListNodes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var ids []string
	for id := range r.nodes {
		ids = append(ids, id)
	}
	return ids
}
