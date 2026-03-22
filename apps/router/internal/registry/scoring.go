package registry

import (
	"math/rand"
	"time"
)

func (r *Registry) SelectNode() (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var activeNodes []string
	var totalScore float64

	for id, node := range r.nodes {
		if node.IsActive {
			score := r.computeScore(id)
			activeNodes = append(activeNodes, id)
			totalScore += score
		}
	}

	if len(activeNodes) == 0 {
		return "", nil
	}

	target := rand.Float64() * totalScore
	current := 0.0
	for _, id := range activeNodes {
		current += r.computeScore(id)
		if current >= target {
			return id, nil
		}
	}

	return activeNodes[0], nil
}

func (r *Registry) computeScore(nodeID string) float64 {
	metrics, ok := r.metrics[nodeID]
	if !ok {
		return 1.0
	}

	score := 1.0

	if metrics.TTFT > 500*time.Millisecond {
		score -= 0.2
	}

	if metrics.SuccessRate < 0.95 {
		score -= 0.5
	}

	if score < 0.1 {
		score = 0.1
	}

	return score
}

func (r *Registry) ReportMetrics(nodeID string, metrics HealthMetrics) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.metrics == nil {
		r.metrics = make(map[string]HealthMetrics)
	}
	r.metrics[nodeID] = metrics
}
