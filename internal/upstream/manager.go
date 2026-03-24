package upstream

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

type ProxyEntry struct {
	ID           int
	Address      string
	Username     string
	Password     string
	Type         string // http | socks5
	Category     string // residential | datacenter
	Enabled      bool
	Weight       int
	AvgLatencyMs int64
	TotalBytes   int64
	FailCount    int
	Healthy      atomic.Bool
	LastCheckAt  time.Time
	LastError    string
}

func (p *ProxyEntry) URL() *url.URL {
	scheme := "http"
	if p.Type == "socks5" {
		scheme = "socks5"
	}
	u := &url.URL{Scheme: scheme, Host: p.Address}
	if p.Username != "" {
		u.User = url.UserPassword(p.Username, p.Password)
	}
	return u
}

type Manager struct {
	mu        sync.RWMutex
	proxies   []*ProxyEntry
	strategy  RotationStrategy
	rrIndex   atomic.Int64
	stickyMap sync.Map
	stickyTTL time.Duration

	db         *sql.DB
	cancelFunc context.CancelFunc
}

func NewManager(db *sql.DB) *Manager {
	return &Manager{
		db:        db,
		strategy:  StrategyRoundRobin,
		stickyTTL: 5 * time.Minute,
	}
}

// LoadProxies loads proxy pool from database.
func (m *Manager) LoadProxies() error {
	rows, err := m.db.Query(
		"SELECT id, address, username, password, type, category, enabled, weight, avg_latency_ms, total_bytes_up + total_bytes_down, fail_count FROM proxies WHERE enabled = 1",
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	var proxies []*ProxyEntry
	for rows.Next() {
		p := &ProxyEntry{}
		var enabled int
		if err := rows.Scan(&p.ID, &p.Address, &p.Username, &p.Password, &p.Type, &p.Category, &enabled, &p.Weight, &p.AvgLatencyMs, &p.TotalBytes, &p.FailCount); err != nil {
			continue
		}
		p.Enabled = enabled == 1
		p.Healthy.Store(true)
		proxies = append(proxies, p)
	}

	m.mu.Lock()
	m.proxies = proxies
	m.mu.Unlock()

	log.Printf("Loaded %d proxies", len(proxies))
	return nil
}

func (m *Manager) SetStrategy(s string) {
	m.strategy = RotationStrategy(s)
}

// SelectProxy picks the next proxy based on strategy and category.
func (m *Manager) SelectProxy(category string, domain string) *ProxyEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var candidates []*ProxyEntry
	for _, p := range m.proxies {
		if p.Category == category && p.Healthy.Load() && p.Enabled {
			candidates = append(candidates, p)
		}
	}
	if len(candidates) == 0 {
		return nil
	}

	switch m.strategy {
	case StrategySticky:
		return m.selectSticky(candidates, domain)
	case StrategyLeastBandwidth:
		return m.selectLeastBandwidth(candidates)
	case StrategyLeastLatency:
		return m.selectLeastLatency(candidates)
	case StrategyWeighted:
		return m.selectWeighted(candidates)
	default:
		return m.selectRoundRobin(candidates)
	}
}

// HealthCheck tests a single proxy.
func (m *Manager) HealthCheck(p *ProxyEntry) (latencyMs int64, ip string, err error) {
	client := &http.Client{
		Transport: &http.Transport{Proxy: http.ProxyURL(p.URL())},
		Timeout:   10 * time.Second,
	}

	start := time.Now()
	resp, err := client.Get("http://httpbin.org/ip")
	if err != nil {
		p.Healthy.Store(false)
		p.FailCount++
		p.LastError = err.Error()
		return 0, "", err
	}
	defer resp.Body.Close()
	latencyMs = time.Since(start).Milliseconds()

	p.Healthy.Store(true)
	p.AvgLatencyMs = latencyMs
	p.FailCount = 0
	p.LastCheckAt = time.Now()

	return latencyMs, "", nil
}

// StartHealthChecks runs periodic health checks.
func (m *Manager) StartHealthChecks(interval time.Duration) {
	ctx, cancel := context.WithCancel(context.Background())
	m.cancelFunc = cancel

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				m.mu.RLock()
				proxies := make([]*ProxyEntry, len(m.proxies))
				copy(proxies, m.proxies)
				m.mu.RUnlock()

				for _, p := range proxies {
					go m.HealthCheck(p)
				}
			}
		}
	}()
}

func (m *Manager) Stop() {
	if m.cancelFunc != nil {
		m.cancelFunc()
	}
}

// CRUD helpers

func (m *Manager) AddProxy(db *sql.DB, address, username, password, proxyType, category string) error {
	_, err := db.Exec(
		"INSERT INTO proxies (address, username, password, type, category) VALUES (?, ?, ?, ?, ?)",
		address, username, password, proxyType, category,
	)
	if err == nil {
		m.LoadProxies()
	}
	return err
}

func (m *Manager) DeleteProxy(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM proxies WHERE id = ?", id)
	if err == nil {
		m.LoadProxies()
	}
	return err
}

func (m *Manager) GetProxies() []*ProxyEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*ProxyEntry, len(m.proxies))
	copy(result, m.proxies)
	return result
}
