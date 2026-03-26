package webapi

import (
	"database/sql"
	"log"
	"time"

	"proxy-bandwidth-saver/internal/cache"
	"proxy-bandwidth-saver/internal/classifier"
	"proxy-bandwidth-saver/internal/config"
	"proxy-bandwidth-saver/internal/database"
	"proxy-bandwidth-saver/internal/meter"
	"proxy-bandwidth-saver/internal/proxy"
)

// LoadSettingsIntoConfig loads DB settings into the config struct.
// This is the single source of truth for settings-to-config mapping.
func LoadSettingsIntoConfig(db *database.DB, cfg *config.Config) {
	if db == nil {
		return
	}
	cfg.HTTPPort = db.GetSettingInt("http_port", 8888)
	cfg.SOCKS5Port = db.GetSettingInt("socks5_port", 8889)
	cfg.MaxConcurrent = db.GetSettingInt("max_concurrent", 500)
	cfg.CacheMemoryMB = db.GetSettingInt("cache_memory_mb", 512)
	cfg.CacheDiskMB = db.GetSettingInt("cache_disk_mb", 2048)
	cfg.MITMEnabled = db.GetSettingBool("mitm_enabled", false)
	cfg.AcceptEncodingEnforce = db.GetSettingBool("accept_encoding_enforce", true)
	cfg.HeaderStripping = db.GetSettingBool("header_stripping", true)
	cfg.HTMLMinification = db.GetSettingBool("html_minification", false)
	cfg.ImageRecompression = db.GetSettingBool("image_recompression", false)
	cfg.ImageQuality = db.GetSettingInt("image_quality", 85)
	cfg.LogRetentionDays = db.GetSettingInt("log_retention_days", 7)

	if addr, err := db.GetSetting("bind_address"); err == nil && addr != "" {
		cfg.BindAddress = addr
	}

	// Auth settings
	cfg.ProxyAuthEnabled = db.GetSettingBool("proxy_auth_enabled", false)
	if u, err := db.GetSetting("proxy_username"); err == nil && u != "" {
		cfg.ProxyUsername = u
	}
	if p, err := db.GetSetting("proxy_password"); err == nil && p != "" {
		cfg.ProxyPassword = p
	}
	cfg.IPWhitelistEnabled = db.GetSettingBool("ip_whitelist_enabled", false)
	if wl, err := db.GetSetting("ip_whitelist"); err == nil {
		cfg.IPWhitelist = wl
	}
}

// ConfigureProxyAuth configures proxy auth from config.
func ConfigureProxyAuth(auth *proxy.ProxyAuth, cfg *config.Config) {
	if auth == nil {
		return
	}
	auth.Configure(
		cfg.ProxyAuthEnabled, cfg.ProxyUsername, cfg.ProxyPassword,
		cfg.IPWhitelistEnabled, cfg.IPWhitelist,
	)
}

// ReloadClassifier reloads rules from DB into classifier.
func ReloadClassifier(reader *sql.DB, cls *classifier.Classifier) {
	if reader == nil || cls == nil {
		return
	}
	rules, err := classifier.LoadRulesFromDB(reader)
	if err != nil {
		log.Printf("Failed to load rules: %v", err)
		return
	}
	cls.Reload(rules)
	log.Printf("Classifier reloaded with %d rules", len(rules))
}

// BuildPipeline creates a proxy pipeline from components.
// Any component may be nil, in which case the corresponding pipeline hook
// retains its default no-op behaviour.
func BuildPipeline(cls *classifier.Classifier, c *cache.CacheLayer, m *meter.Meter) *proxy.Pipeline {
	pipeline := proxy.NewDefaultPipeline()
	if cls != nil {
		pipeline.Classifier = cls.Classify
	}
	if c != nil {
		pipeline.CacheCheck = c.CheckCache
		pipeline.CacheStore = c.StoreCache
	}
	if m != nil {
		pipeline.Meter = func(ctx *proxy.RequestCtx) {
			respBytes := ctx.RespBytes
			if ctx.TotalBytes() > ctx.ReqBytes {
				respBytes = ctx.TotalBytes() - ctx.ReqBytes
			}
			m.Record(meter.RequestLog{
				Timestamp:     ctx.StartTime,
				Domain:        ctx.Domain,
				Method:        ctx.Request.Method,
				URL:           ctx.Request.URL.String(),
				Route:         string(ctx.Route),
				RequestBytes:  ctx.ReqBytes,
				ResponseBytes: respBytes,
				Cached:        ctx.Cached,
				ProxyID:       ctx.ProxyID,
				LatencyMs:     time.Since(ctx.StartTime).Milliseconds(),
			})
		}
	}
	return pipeline
}
