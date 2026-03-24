package meter

import (
	"context"
	"database/sql"
	"log"
	"sync/atomic"
	"time"
)

// RequestLog represents bandwidth data for a single request
type RequestLog struct {
	Timestamp     time.Time
	Domain        string
	Method        string
	URL           string
	Route         string
	RequestBytes  int64
	ResponseBytes int64
	Cached        bool
	ProxyID       int
	LatencyMs     int64
	ContentType   string
	StatusCode    int
}

// Meter tracks bandwidth and costs
type Meter struct {
	logChan chan RequestLog

	// Atomic counters for realtime stats
	totalBytesToday       atomic.Int64
	residentialBytesToday atomic.Int64
	bytesPerSecond        atomic.Int64
	residentialBPS        atomic.Int64
	cacheHits             atomic.Int64
	cacheMisses           atomic.Int64

	// Savings
	bytesSavedCache       atomic.Int64
	bytesSavedDirect      atomic.Int64
	bytesSavedDatacenter  atomic.Int64
	bytesSavedCompression atomic.Int64

	db         *sql.DB
	cancelFunc context.CancelFunc
}

func NewMeter(db *sql.DB) *Meter {
	return &Meter{
		logChan: make(chan RequestLog, 10000),
		db:      db,
	}
}

// Start launches background goroutines for logging and aggregation
func (m *Meter) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	m.cancelFunc = cancel

	// Log writer goroutine
	go m.logWriter(ctx)

	// Speed calculator goroutine (1s interval)
	go m.speedCalculator(ctx)

	// Aggregation goroutine (5min interval)
	go m.aggregator(ctx)
}

// Stop shuts down the meter
func (m *Meter) Stop() {
	if m.cancelFunc != nil {
		m.cancelFunc()
	}
}

// Record logs a request's bandwidth data
func (m *Meter) Record(entry RequestLog) {
	totalBytes := entry.RequestBytes + entry.ResponseBytes

	// Update atomic counters
	m.totalBytesToday.Add(totalBytes)
	if entry.Route == "residential" {
		m.residentialBytesToday.Add(totalBytes)
	}
	if entry.Cached {
		m.cacheHits.Add(1)
		m.bytesSavedCache.Add(totalBytes)
	} else {
		m.cacheMisses.Add(1)
	}
	if entry.Route == "direct" {
		m.bytesSavedDirect.Add(totalBytes)
	} else if entry.Route == "datacenter" {
		m.bytesSavedDatacenter.Add(totalBytes)
	}

	// Send to log channel (non-blocking)
	select {
	case m.logChan <- entry:
	default:
		// Channel full, drop oldest
		log.Println("Warning: meter log channel full, dropping entry")
	}
}

// GetRealtimeStats returns current stats
func (m *Meter) GetRealtimeStats() (totalToday, residentialToday, bps, resBPS, cacheHits, cacheMisses int64) {
	return m.totalBytesToday.Load(),
		m.residentialBytesToday.Load(),
		m.bytesPerSecond.Load(),
		m.residentialBPS.Load(),
		m.cacheHits.Load(),
		m.cacheMisses.Load()
}

// GetSavings returns bytes saved by each source
func (m *Meter) GetSavings() (cache, direct, datacenter, compression int64) {
	return m.bytesSavedCache.Load(),
		m.bytesSavedDirect.Load(),
		m.bytesSavedDatacenter.Load(),
		m.bytesSavedCompression.Load()
}

func (m *Meter) logWriter(ctx context.Context) {
	batch := make([]RequestLog, 0, 100)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Flush remaining
			if len(batch) > 0 {
				m.writeBatch(batch)
			}
			return
		case entry := <-m.logChan:
			batch = append(batch, entry)
			if len(batch) >= 100 {
				m.writeBatch(batch)
				batch = batch[:0]
			}
		case <-ticker.C:
			if len(batch) > 0 {
				m.writeBatch(batch)
				batch = batch[:0]
			}
		}
	}
}

func (m *Meter) writeBatch(entries []RequestLog) {
	tx, err := m.db.Begin()
	if err != nil {
		log.Printf("Meter: begin tx error: %v", err)
		return
	}

	stmt, err := tx.Prepare(
		`INSERT INTO bandwidth_log (domain, url, method, route_decision, request_bytes, response_bytes, total_bytes, cache_hit, proxy_id, duration_ms, content_type, status_code)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
	)
	if err != nil {
		tx.Rollback()
		return
	}
	defer stmt.Close()

	for _, e := range entries {
		url := e.URL
		if len(url) > 500 {
			url = url[:500]
		}
		cacheHit := 0
		if e.Cached {
			cacheHit = 1
		}
		stmt.Exec(e.Domain, url, e.Method, e.Route,
			e.RequestBytes, e.ResponseBytes, e.RequestBytes+e.ResponseBytes,
			cacheHit, e.ProxyID, e.LatencyMs, e.ContentType, e.StatusCode)
	}

	tx.Commit()
}

func (m *Meter) speedCalculator(ctx context.Context) {
	var lastTotal, lastRes int64
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			currentTotal := m.totalBytesToday.Load()
			currentRes := m.residentialBytesToday.Load()
			m.bytesPerSecond.Store(currentTotal - lastTotal)
			m.residentialBPS.Store(currentRes - lastRes)
			lastTotal = currentTotal
			lastRes = currentRes
		}
	}
}

func (m *Meter) aggregator(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.runAggregation()
		}
	}
}

func (m *Meter) runAggregation() {
	// Aggregate bandwidth_log into bandwidth_stats (hourly buckets)
	_, err := m.db.Exec(`
		INSERT INTO bandwidth_stats (period, residential_bytes, datacenter_bytes, direct_bytes, cache_saved_bytes, total_requests, cache_hit_count, cache_miss_count)
		SELECT
			strftime('%Y-%m-%d %H:00', timestamp) as period,
			COALESCE(SUM(CASE WHEN route_decision = 'residential' THEN total_bytes ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN route_decision = 'datacenter' THEN total_bytes ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN route_decision = 'direct' THEN total_bytes ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN cache_hit = 1 THEN total_bytes ELSE 0 END), 0),
			COUNT(*),
			SUM(cache_hit),
			SUM(CASE WHEN cache_hit = 0 THEN 1 ELSE 0 END)
		FROM bandwidth_log
		WHERE timestamp > datetime('now', '-10 minutes')
		GROUP BY strftime('%Y-%m-%d %H:00', timestamp)
		ON CONFLICT(period) DO UPDATE SET
			residential_bytes = residential_bytes + excluded.residential_bytes,
			datacenter_bytes = datacenter_bytes + excluded.datacenter_bytes,
			direct_bytes = direct_bytes + excluded.direct_bytes,
			cache_saved_bytes = cache_saved_bytes + excluded.cache_saved_bytes,
			total_requests = total_requests + excluded.total_requests,
			cache_hit_count = cache_hit_count + excluded.cache_hit_count,
			cache_miss_count = cache_miss_count + excluded.cache_miss_count
	`)
	if err != nil {
		log.Printf("Aggregation error: %v", err)
	}

	// Clean old logs
	_, err = m.db.Exec("DELETE FROM bandwidth_log WHERE timestamp < datetime('now', '-7 days')")
	if err != nil {
		log.Printf("Log cleanup error: %v", err)
	}
}
