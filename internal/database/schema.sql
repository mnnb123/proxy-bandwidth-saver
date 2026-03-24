-- Proxy pool
CREATE TABLE IF NOT EXISTS proxies (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    address TEXT NOT NULL,
    username TEXT DEFAULT '',
    password TEXT DEFAULT '',
    type TEXT DEFAULT 'http',
    category TEXT DEFAULT 'residential',
    enabled INTEGER DEFAULT 1,
    weight INTEGER DEFAULT 1,
    total_bytes_up INTEGER DEFAULT 0,
    total_bytes_down INTEGER DEFAULT 0,
    total_requests INTEGER DEFAULT 0,
    fail_count INTEGER DEFAULT 0,
    avg_latency_ms INTEGER DEFAULT 0,
    last_check_at TEXT,
    last_error TEXT DEFAULT '',
    created_at TEXT DEFAULT (datetime('now'))
);

-- Traffic rules
CREATE TABLE IF NOT EXISTS rules (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    rule_type TEXT NOT NULL,
    pattern TEXT NOT NULL,
    action TEXT NOT NULL,
    priority INTEGER DEFAULT 0,
    enabled INTEGER DEFAULT 1,
    hit_count INTEGER DEFAULT 0,
    bytes_saved INTEGER DEFAULT 0,
    created_at TEXT DEFAULT (datetime('now'))
);

-- Bandwidth log (per-request through residential)
CREATE TABLE IF NOT EXISTS bandwidth_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp TEXT DEFAULT (datetime('now')),
    domain TEXT NOT NULL,
    url TEXT,
    method TEXT,
    proxy_id INTEGER REFERENCES proxies(id),
    request_bytes INTEGER NOT NULL DEFAULT 0,
    response_bytes INTEGER NOT NULL DEFAULT 0,
    total_bytes INTEGER NOT NULL DEFAULT 0,
    content_type TEXT DEFAULT '',
    status_code INTEGER DEFAULT 0,
    cache_hit INTEGER DEFAULT 0,
    route_decision TEXT,
    duration_ms INTEGER DEFAULT 0
);

-- Aggregated stats (pre-computed hourly)
CREATE TABLE IF NOT EXISTS bandwidth_stats (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    period TEXT NOT NULL,
    residential_bytes INTEGER DEFAULT 0,
    datacenter_bytes INTEGER DEFAULT 0,
    direct_bytes INTEGER DEFAULT 0,
    cache_saved_bytes INTEGER DEFAULT 0,
    total_requests INTEGER DEFAULT 0,
    cache_hit_count INTEGER DEFAULT 0,
    cache_miss_count INTEGER DEFAULT 0,
    cost_usd REAL DEFAULT 0
);

-- Settings key-value store
CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_bandwidth_log_timestamp ON bandwidth_log(timestamp);
CREATE INDEX IF NOT EXISTS idx_bandwidth_log_domain ON bandwidth_log(domain);
CREATE INDEX IF NOT EXISTS idx_bandwidth_stats_period ON bandwidth_stats(period);
CREATE INDEX IF NOT EXISTS idx_rules_type_pattern ON rules(rule_type, pattern);
