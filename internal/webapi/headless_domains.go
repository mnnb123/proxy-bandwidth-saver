package webapi

import "proxy-bandwidth-saver/internal/database"

// GetDomainStatsJSON returns bandwidth stats grouped by domain.
func (a *HeadlessApp) GetDomainStatsJSON(period string) interface{} {
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

	query := `
		SELECT
			domain,
			SUM(total_bytes) as total_bytes,
			COUNT(*) as requests,
			COALESCE(
				(SELECT route_decision FROM bandwidth_log b2
				 WHERE b2.domain = bandwidth_log.domain
				 GROUP BY route_decision ORDER BY COUNT(*) DESC LIMIT 1),
				'direct'
			) as route,
			CASE WHEN COUNT(*) > 0
				THEN CAST(SUM(cache_hit) AS REAL) / COUNT(*) * 100
				ELSE 0
			END as cache_hit_pct
		FROM bandwidth_log
		WHERE timestamp > ` + timeFilter + `
		GROUP BY domain
		ORDER BY total_bytes DESC
		LIMIT 100
	`

	rows, err := a.db.Reader.Query(query)
	if err != nil {
		return []database.DomainStat{}
	}
	defer rows.Close()

	var stats []database.DomainStat
	for rows.Next() {
		var s database.DomainStat
		if err := rows.Scan(&s.Domain, &s.TotalBytes, &s.Requests, &s.Route, &s.CacheHitPct); err != nil {
			continue
		}
		stats = append(stats, s)
	}
	if stats == nil {
		stats = []database.DomainStat{}
	}
	return stats
}
