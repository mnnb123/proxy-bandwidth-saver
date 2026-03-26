package webapi

import (
	"fmt"
	"strings"

	"proxy-bandwidth-saver/internal/database"
	"proxy-bandwidth-saver/internal/upstream"
)

func (a *HeadlessApp) GetProxiesJSON() interface{} {
	proxies := a.getProxies()
	// Mask passwords in API response
	for i := range proxies {
		if proxies[i].Password != "" {
			proxies[i].Password = "****"
		}
	}
	return proxies
}

func (a *HeadlessApp) getProxies() []database.Proxy {
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

func (a *HeadlessApp) AddProxy(address, username, password, proxyType, category string) error {
	if a.db == nil || a.upstream == nil {
		return fmt.Errorf("not initialized")
	}
	if err := a.upstream.AddProxy(a.db.Writer, address, username, password, proxyType, category); err != nil {
		return err
	}
	a.remapAllProxies()
	return nil
}

func (a *HeadlessApp) DeleteProxy(id int) error {
	if a.db == nil || a.upstream == nil {
		return fmt.Errorf("not initialized")
	}
	if a.portMapper != nil {
		a.portMapper.UnmapProxy(id)
	}
	return a.upstream.DeleteProxy(a.db.Writer, id)
}

func (a *HeadlessApp) ImportProxies(text string) int {
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

func (a *HeadlessApp) GetOutputProxiesJSON() interface{} {
	if a.portMapper == nil {
		return []database.OutputProxy{}
	}
	mappings := a.portMapper.GetMappings()
	result := make([]database.OutputProxy, len(mappings))
	for i, m := range mappings {
		result[i] = database.OutputProxy{
			ProxyID:   m.ProxyID,
			LocalAddr: m.LocalAddr,
			LocalPort: m.LocalPort,
			Protocol:  m.Protocol,
			Upstream:  m.Upstream,
			Type:      m.Type,
		}
	}
	return result
}
