package registry

import (
	"log/slog"
	"sync"

	"github.com/gorilla/websocket"
)

type Node struct {
	ID            string
	WalletAddress string
	Conn          *websocket.Conn
	mu            sync.Mutex
}

type Registry struct {
	nodes map[string]*Node
	mu    sync.RWMutex
	log   *slog.Logger
}

func New(logger *slog.Logger) *Registry {
	return &Registry{
		nodes: make(map[string]*Node),
		log:   logger,
	}
}

func (r *Registry) Register(id string, wallet string, conn *websocket.Conn) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// TODO: Call Smart Contract to check if wallet has enough $PROXY staked
	// For now, log the wallet address for trust tracking
	r.log.Info("Checking node eligibility", "wallet", wallet)

	if old, ok := r.nodes[id]; ok {
		old.Conn.Close()
	}

	r.nodes[id] = &Node{
		ID:            id,
		WalletAddress: wallet,
		Conn:          conn,
	}
	r.log.Info("Node registered", "id", id, "wallet", wallet)
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
