package main

import (
	"fmt"
	"strings"

	"proxy-bandwidth-saver/internal/database"
	"proxy-bandwidth-saver/internal/proxy"
	"proxy-bandwidth-saver/internal/upstream"
)

// GetProxies returns all upstream proxies from DB.
func (a *App) GetProxies() []database.Proxy {
	if a.db == nil {
		return nil
	}
	rows, err := a.db.Reader.Query(
		"SELECT id, address, username, password, type, category, enabled, weight, total_bytes_up, total_bytes_down, total_requests, fail_count, avg_latency_ms, last_check_at, last_error, created_at FROM proxies ORDER BY id ASC",
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var proxies []database.Proxy
	for rows.Next() {
		var p database.Proxy
		var enabled int
		var lastCheck, lastErr, created *string
		if err := rows.Scan(&p.ID, &p.Address, &p.Username, &p.Password, &p.Type, &p.Category, &enabled, &p.Weight, &p.TotalBytesUp, &p.TotalBytesDown, &p.TotalRequests, &p.FailCount, &p.AvgLatencyMs, &lastCheck, &lastErr, &created); err != nil {
			continue
		}
		p.Enabled = enabled == 1
		if lastErr != nil {
			p.LastError = *lastErr
		}
		proxies = append(proxies, p)
	}
	return proxies
}

func (a *App) AddProxy(address, username, password, proxyType, category string) error {
	if a.db == nil || a.upstream == nil {
		return fmt.Errorf("not initialized")
	}
	if err := a.upstream.AddProxy(a.db.Writer, address, username, password, proxyType, category); err != nil {
		return err
	}
	a.remapAllProxies()
	return nil
}

func (a *App) DeleteProxy(id int) error {
	if a.db == nil || a.upstream == nil {
		return fmt.Errorf("not initialized")
	}
	if a.portMapper != nil {
		a.portMapper.UnmapProxy(id)
	}
	return a.upstream.DeleteProxy(a.db.Writer, id)
}

func (a *App) ClearAllProxies() error {
	if a.db == nil {
		return fmt.Errorf("not initialized")
	}
	if a.portMapper != nil {
		a.portMapper.StopAll()
	}
	if _, err := a.db.Writer.Exec("DELETE FROM proxies"); err != nil {
		return err
	}
	if a.upstream != nil {
		a.upstream.LoadProxies()
	}
	return nil
}

func (a *App) ImportProxies(text string) int {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	count := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		addr, user, pass, pType, err := upstream.ParseProxyLine(line)
		if err != nil {
			continue
		}
		if a.upstream.AddProxy(a.db.Writer, addr, user, pass, pType, "residential") == nil {
			count++
		}
	}
	if count > 0 {
		a.remapAllProxies()
	}
	return count
}

// GetOutputProxies returns the list of local output proxies (one port per upstream).
func (a *App) GetOutputProxies() []database.OutputProxy {
	if a.portMapper == nil {
		return nil
	}
	mappings := a.portMapper.GetMappings()
	result := make([]database.OutputProxy, len(mappings))
	for i, m := range mappings {
		result[i] = database.OutputProxy{
			ProxyID:   m.ProxyID,
			LocalAddr: m.LocalAddr,
			LocalPort: m.LocalPort,
			Upstream:  m.Upstream,
			Type:      m.Type,
		}
	}
	return result
}

// remapAllProxies reads all proxies from DB and ensures each has a local port mapping.
func (a *App) remapAllProxies() {
	if a.portMapper == nil || a.db == nil {
		return
	}
	proxies := a.GetProxies()
	for _, p := range proxies {
		if !p.Enabled {
			continue
		}
		a.portMapper.MapProxy(p.ID, proxy.UpstreamInfo{
			Address:  p.Address,
			Username: p.Username,
			Password: p.Password,
			Type:     p.Type,
		})
	}
}
