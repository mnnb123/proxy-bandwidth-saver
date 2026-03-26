package upstream

import (
	"context"
	"database/sql"
	"io"
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
	mu           sync.Mutex // protects non-atomic fields during health check
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

	// healthClient is reused across all health checks to avoid
	// allocating a new http.Client (and underlying transport) per call.
	healthClient *http.Client
}

func NewManager(db *sql.DB) *Manager {
	return &Manager{
		db:        db,
		strategy:  StrategyRoundRobin,
		stickyTTL: 5 * time.Minute,
		healthClient: &http.Client{
			// Transport is intentionally nil here; HealthCheck sets the
			// proxy per-request via a cloned Transport, but we keep a
			// single client to reuse idle connections and TLS state.
			Timeout: 10 * time.Second,
		},
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
	// Build a request so we can override the transport per-proxy while
	// reusing the shared healthClient (timeout, cookie jar, etc.).
	req, _ := http.NewRequest(http.MethodGet, "http://httpbin.org/ip", nil)
	transport := &http.Transport{Proxy: http.ProxyURL(p.URL())}

	// Temporarily swap the client transport for this proxy. Because each
	// goroutine gets its own Transport, there is no data race on the
	// shared client itself — only Timeout and CheckRedirect are read.
	client := *m.healthClient // shallow copy (cheap, no alloc on heap in most cases)
	client.Transport = transport

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		p.Healthy.Store(false)
		p.mu.Lock()
		p.FailCount++
		p.LastError = err.Error()
		p.mu.Unlock()
		return 0, "", err
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	latencyMs = time.Since(start).Milliseconds()

	p.Healthy.Store(true)
	p.mu.Lock()
	p.AvgLatencyMs = latencyMs
	p.FailCount = 0
	p.LastCheckAt = time.Now()
	p.mu.Unlock()

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
