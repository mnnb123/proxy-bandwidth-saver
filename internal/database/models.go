package database

import "time"

type Proxy struct {
	ID            int       `json:"id"`
	Address       string    `json:"address"`
	Username      string    `json:"username"`
	Password      string    `json:"password"`
	Type          string    `json:"type"`     // http | socks5
	Category      string    `json:"category"` // residential | datacenter
	Enabled       bool      `json:"enabled"`
	Weight        int       `json:"weight"`
	TotalBytesUp  int64     `json:"totalBytesUp"`
	TotalBytesDown int64    `json:"totalBytesDown"`
	TotalRequests int64     `json:"totalRequests"`
	FailCount     int       `json:"failCount"`
	AvgLatencyMs  int       `json:"avgLatencyMs"`
	LastCheckAt   *time.Time `json:"lastCheckAt"`
	LastError     string    `json:"lastError"`
	CreatedAt     time.Time `json:"createdAt"`
}

type Rule struct {
	ID        int    `json:"id"`
	RuleType  string `json:"ruleType"`  // domain | content_type | url_pattern
	Pattern   string `json:"pattern"`
	Action    string `json:"action"`    // direct | datacenter | residential | block | bypass | bypass_vps
	Priority  int    `json:"priority"`
	Enabled   bool   `json:"enabled"`
	HitCount  int64  `json:"hitCount"`
	BytesSaved int64 `json:"bytesSaved"`
	CreatedAt string `json:"createdAt"`
}

type BandwidthLog struct {
	ID            int       `json:"id"`
	Timestamp     time.Time `json:"timestamp"`
	Domain        string    `json:"domain"`
	URL           string    `json:"url"`
	Method        string    `json:"method"`
	ProxyID       *int      `json:"proxyId"`
	RequestBytes  int64     `json:"requestBytes"`
	ResponseBytes int64     `json:"responseBytes"`
	TotalBytes    int64     `json:"totalBytes"`
	ContentType   string    `json:"contentType"`
	StatusCode    int       `json:"statusCode"`
	CacheHit      bool      `json:"cacheHit"`
	RouteDecision string    `json:"routeDecision"`
	DurationMs    int64     `json:"durationMs"`
}

type BandwidthStats struct {
	ID               int     `json:"id"`
	Period           string  `json:"period"`
	ResidentialBytes int64   `json:"residentialBytes"`
	DatacenterBytes  int64   `json:"datacenterBytes"`
	DirectBytes      int64   `json:"directBytes"`
	CacheSavedBytes  int64   `json:"cacheSavedBytes"`
	TotalRequests    int64   `json:"totalRequests"`
	CacheHitCount    int64   `json:"cacheHitCount"`
	CacheMissCount   int64   `json:"cacheMissCount"`
	CostUSD          float64 `json:"costUsd"`
}

type Setting struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// API response types

type ProxyStatus struct {
	Running     bool   `json:"running"`
	HTTPPort    int    `json:"httpPort"`
	SOCKS5Port  int    `json:"socks5Port"`
	Uptime      int64  `json:"uptime"`
	Connections int    `json:"connections"`
}

type RealtimeStats struct {
	BytesPerSecond    int64   `json:"bytesPerSecond"`
	ResidentialBPS    int64   `json:"residentialBps"`
	TotalToday        int64   `json:"totalToday"`
	ResidentialToday  int64   `json:"residentialToday"`
	CostToday         float64 `json:"costToday"`
	CacheHitRatio     float64 `json:"cacheHitRatio"`
	ActiveConnections int     `json:"activeConnections"`
}

type CostSummary struct {
	CostToday  float64 `json:"costToday"`
	CostWeek   float64 `json:"costWeek"`
	CostMonth  float64 `json:"costMonth"`
	CostTotal  float64 `json:"costTotal"`
	SavedBytes int64   `json:"savedBytes"`
	SavedCost  float64 `json:"savedCost"`
}

type BudgetStatus struct {
	MonthlyBudgetGB float64 `json:"monthlyBudgetGb"`
	UsedGB          float64 `json:"usedGb"`
	UsedPercent     float64 `json:"usedPercent"`
	RemainingGB     float64 `json:"remainingGb"`
	CostPerGB       float64 `json:"costPerGb"`
	ProjectedGB     float64 `json:"projectedGb"`
}

type CacheStats struct {
	MemoryUsedMB float64 `json:"memoryUsedMb"`
	DiskUsedMB   float64 `json:"diskUsedMb"`
	Entries      int64   `json:"entries"`
	HitCount     int64   `json:"hitCount"`
	MissCount    int64   `json:"missCount"`
	HitRatio     float64 `json:"hitRatio"`
	BytesSaved   int64   `json:"bytesSaved"`
}

type DomainStat struct {
	Domain      string  `json:"domain"`
	TotalBytes  int64   `json:"totalBytes"`
	Requests    int64   `json:"requests"`
	Route       string  `json:"route"`
	CacheHitPct float64 `json:"cacheHitPct"`
}

type BandwidthPoint struct {
	Time             string `json:"time"`
	ResidentialBytes int64  `json:"residentialBytes"`
	DatacenterBytes  int64  `json:"datacenterBytes"`
	DirectBytes      int64  `json:"directBytes"`
}

// OutputProxy represents a local port mapped to an upstream proxy.
type OutputProxy struct {
	ProxyID   int    `json:"proxyId"`
	LocalAddr string `json:"localAddr"` // e.g. "127.0.0.1:30000"
	LocalPort int    `json:"localPort"`
	Protocol  string `json:"protocol"`  // "http" or "socks5" (output protocol)
	Upstream  string `json:"upstream"`  // upstream address
	Type      string `json:"type"`      // upstream type
}
