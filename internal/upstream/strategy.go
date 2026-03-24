package upstream

import (
	"math/rand"
	"time"
)

type RotationStrategy string

const (
	StrategyRoundRobin     RotationStrategy = "round-robin"
	StrategyLeastBandwidth RotationStrategy = "least-bandwidth"
	StrategyLeastLatency   RotationStrategy = "least-latency"
	StrategySticky         RotationStrategy = "sticky"
	StrategyWeighted       RotationStrategy = "weighted"
)

func (m *Manager) selectRoundRobin(candidates []*ProxyEntry) *ProxyEntry {
	idx := m.rrIndex.Add(1) - 1
	return candidates[idx%int64(len(candidates))]
}

func (m *Manager) selectLeastBandwidth(candidates []*ProxyEntry) *ProxyEntry {
	best := candidates[0]
	for _, p := range candidates[1:] {
		if p.TotalBytes < best.TotalBytes {
			best = p
		}
	}
	return best
}

func (m *Manager) selectLeastLatency(candidates []*ProxyEntry) *ProxyEntry {
	best := candidates[0]
	for _, p := range candidates[1:] {
		if p.AvgLatencyMs < best.AvgLatencyMs {
			best = p
		}
	}
	return best
}

func (m *Manager) selectSticky(candidates []*ProxyEntry, domain string) *ProxyEntry {
	if val, ok := m.stickyMap.Load(domain); ok {
		proxyID := val.(int)
		for _, p := range candidates {
			if p.ID == proxyID {
				return p
			}
		}
	}
	selected := m.selectRoundRobin(candidates)
	m.stickyMap.Store(domain, selected.ID)
	time.AfterFunc(m.stickyTTL, func() {
		m.stickyMap.Delete(domain)
	})
	return selected
}

func (m *Manager) selectWeighted(candidates []*ProxyEntry) *ProxyEntry {
	totalWeight := 0
	for _, p := range candidates {
		totalWeight += p.Weight
	}
	if totalWeight == 0 {
		return candidates[0]
	}
	r := rand.Intn(totalWeight)
	for _, p := range candidates {
		r -= p.Weight
		if r < 0 {
			return p
		}
	}
	return candidates[0]
}
