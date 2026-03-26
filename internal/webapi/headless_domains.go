package webapi

import (
	"fmt"
	"log"

	"proxy-bandwidth-saver/internal/database"
)

// GetDomainStatsJSON returns bandwidth stats grouped by domain.
// proxyID=0 means all proxies, proxyID>0 means specific output proxy port.
func (a *HeadlessApp) GetDomainStatsJSON(period string) interface{} {
	return a.getDomainStats(period, 0)
}

// GetDomainStatsByPortJSON returns bandwidth stats for a specific proxy port.
func (a *HeadlessApp) GetDomainStatsByPortJSON(period string, proxyID int) interface{} {
	return a.getDomainStats(period, proxyID)
}

func (a *HeadlessApp) getDomainStats(period string, proxyID int) []database.DomainStat {
	if a.db == nil {
		return []database.DomainStat{}
	}

	var timeFilter string
	switch period {
	case "1h":
		timeFilter = "datetime('now', '-1 hour')"
	case "24h":
		timeFilter = "datetime('now', '-1 day')"
	case "7d":
		timeFilter = "datetime('now', '-7 days')"
	case "30d":
		timeFilter = "datetime('now', '-30 days')"
	default:
		timeFilter = "datetime('now', '-1 day')"
	}

	proxyFilter := ""
	if proxyID > 0 {
		proxyFilter = fmt.Sprintf(" AND proxy_id = %d", proxyID)
	}

	query := `
		SELECT
			domain,
			SUM(total_bytes) as total_bytes,
			COUNT(*) as requests,
			COALESCE(route_decision, 'residential') as route,
			CASE WHEN COUNT(*) > 0
				THEN CAST(SUM(cache_hit) AS REAL) / COUNT(*) * 100
				ELSE 0
			END as cache_hit_pct,
			COALESCE(proxy_id, 0) as proxy_id
		FROM bandwidth_log
		WHERE timestamp > ` + timeFilter + proxyFilter + `
		GROUP BY domain, proxy_id
		ORDER BY total_bytes DESC
		LIMIT 200
	`

	rows, err := a.db.Reader.Query(query)
	if err != nil {
		return []database.DomainStat{}
	}
	defer rows.Close()

	var stats []database.DomainStat
	for rows.Next() {
		var s database.DomainStat
		if err := rows.Scan(&s.Domain, &s.TotalBytes, &s.Requests, &s.Route, &s.CacheHitPct, &s.ProxyID); err != nil {
			continue
		}
		stats = append(stats, s)
	}
	if stats == nil {
		stats = []database.DomainStat{}
	}
	return stats
}

// ClearDomainStats deletes all bandwidth_log entries.
func (a *HeadlessApp) ClearDomainStats() error {
	if a.db == nil {
		return fmt.Errorf("not initialized")
	}
	_, err := a.db.Writer.Exec("DELETE FROM bandwidth_log")
	if err != nil {
		return err
	}
	log.Println("Domain stats cleared")
	return nil
}

// AutoClearDomainStats deletes bandwidth_log entries older than the given minutes.
func (a *HeadlessApp) AutoClearDomainStats(minutes int) error {
	if a.db == nil {
		return fmt.Errorf("not initialized")
	}
	_, err := a.db.Writer.Exec(
		fmt.Sprintf("DELETE FROM bandwidth_log WHERE timestamp < datetime('now', '-%d minutes')", minutes),
	)
	return err
}
