package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"proxy-bandwidth-saver/internal/cache"
	"proxy-bandwidth-saver/internal/classifier"
	"proxy-bandwidth-saver/internal/config"
	"proxy-bandwidth-saver/internal/database"
	"proxy-bandwidth-saver/internal/meter"
	"proxy-bandwidth-saver/internal/optimizer"
	"proxy-bandwidth-saver/internal/proxy"
	"proxy-bandwidth-saver/internal/upstream"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx        context.Context
	cfg        *config.Config
	db         *database.DB
	server     *proxy.ProxyServer
	classifier *classifier.Classifier
	cache      *cache.CacheLayer
	meter      *meter.Meter
	upstream   *upstream.Manager
	optCfg     *optimizer.Config
	portMapper *proxy.PortMapper
	proxyAuth  *proxy.ProxyAuth
}

func NewApp() *App {
	return &App{
		cfg: config.DefaultConfig(),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	if err := config.EnsureDataDirs(a.cfg); err != nil {
		log.Printf("Failed to create data dirs: %v", err)
	}

	db, err := database.NewDB(a.cfg.DBPath)
	if err != nil {
		log.Printf("Failed to init database: %v", err)
		wailsRuntime.MessageDialog(ctx, wailsRuntime.MessageDialogOptions{
			Type:    wailsRuntime.ErrorDialog,
			Title:   "Database Error",
			Message: fmt.Sprintf("Không thể khởi tạo database: %v", err),
		})
		return
	}
	a.db = db

	if err := a.db.SeedDefaults(); err != nil {
		log.Printf("Failed to seed defaults: %v", err)
	}
	if err := classifier.SeedDefaultRules(a.db.Writer); err != nil {
		log.Printf("Failed to seed default rules: %v", err)
	}

	a.loadSettingsIntoConfig()

	a.classifier = classifier.NewClassifier()
	a.reloadClassifier()

	cacheLayer, err := cache.NewCacheLayer(a.cfg.CacheDir, a.cfg.CacheMemoryMB, a.cfg.CacheDiskMB)
	if err != nil {
		log.Printf("Failed to init cache: %v", err)
	} else {
		a.cache = cacheLayer
	}

	a.meter = meter.NewMeter(a.db.Writer)
	a.meter.Start()

	a.upstream = upstream.NewManager(a.db.Reader)
	a.upstream.LoadProxies()
	a.upstream.StartHealthChecks(60 * time.Second)

	a.proxyAuth = proxy.NewProxyAuth()
	a.configureProxyAuth()
	basePort := a.db.GetSettingInt("base_port", 30000)
	a.portMapper = proxy.NewPortMapper("127.0.0.1", basePort, a.proxyAuth)

	a.optCfg = &optimizer.Config{
		AcceptEncodingEnforce: a.cfg.AcceptEncodingEnforce,
		HeaderStripping:       a.cfg.HeaderStripping,
		HTMLMinification:      a.cfg.HTMLMinification,
		ImageRecompression:    a.cfg.ImageRecompression,
		ImageQuality:          a.cfg.ImageQuality,
	}

	a.remapAllProxies()
	go a.emitRealtimeStats()

	log.Println("App started successfully")
}

func (a *App) shutdown(ctx context.Context) {
	if a.portMapper != nil {
		a.portMapper.StopAll()
	}
	if a.server != nil && a.server.IsRunning() {
		a.server.Stop()
	}
	if a.meter != nil {
		a.meter.Stop()
	}
	if a.upstream != nil {
		a.upstream.Stop()
	}
	if a.cache != nil {
		a.cache.Close()
	}
	if a.db != nil {
		a.db.Close()
	}
	log.Println("App shutdown complete")
}

// === Proxy Control ===

func (a *App) StartProxy() error {
	if a.server != nil && a.server.IsRunning() {
		return fmt.Errorf("proxy already running")
	}

	a.configureProxyAuth()

	a.server = proxy.NewProxyServer(proxy.ServerConfig{
		HTTPPort:    a.cfg.HTTPPort,
		SOCKS5Port:  a.cfg.SOCKS5Port,
		BindAddress: a.cfg.BindAddress,
		MITMEnabled: a.cfg.MITMEnabled,
		CertDir:     a.cfg.CertDir,
	})
	// Share auth instance with server
	a.server.GetAuth().Configure(
		a.cfg.ProxyAuthEnabled, a.cfg.ProxyUsername, a.cfg.ProxyPassword,
		a.cfg.IPWhitelistEnabled, a.cfg.IPWhitelist,
	)

	pipeline := proxy.NewDefaultPipeline()
	if a.classifier != nil {
		pipeline.Classifier = a.classifier.Classify
	}
	if a.cache != nil {
		pipeline.CacheCheck = a.cache.CheckCache
		pipeline.CacheStore = a.cache.StoreCache
	}
	if a.meter != nil {
		pipeline.Meter = func(ctx *proxy.RequestCtx) {
			respBytes := ctx.RespBytes
			if ctx.TotalBytes() > ctx.ReqBytes {
				respBytes = ctx.TotalBytes() - ctx.ReqBytes
			}
			a.meter.Record(meter.RequestLog{
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
	a.server.SetPipeline(pipeline)

	if err := a.server.Start(); err != nil {
		return err
	}
	wailsRuntime.EventsEmit(a.ctx, "proxy:started")
	return nil
}

func (a *App) StopProxy() error {
	if a.server == nil || !a.server.IsRunning() {
		return fmt.Errorf("proxy not running")
	}
	if err := a.server.Stop(); err != nil {
		return err
	}
	wailsRuntime.EventsEmit(a.ctx, "proxy:stopped")
	return nil
}

func (a *App) GetProxyStatus() database.ProxyStatus {
	status := database.ProxyStatus{
		HTTPPort:   a.cfg.HTTPPort,
		SOCKS5Port: a.cfg.SOCKS5Port,
	}
	if a.server != nil {
		status.Running = a.server.IsRunning()
		status.Uptime = a.server.GetUptime()
		status.Connections = a.server.GetConnectionCount()
	}
	return status
}

// === Internal helpers ===

func (a *App) emitRealtimeStats() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if a.ctx == nil || a.meter == nil {
			return
		}
		totalToday, residentToday, bps, resBPS, hits, misses := a.meter.GetRealtimeStats()
		hitRatio := float64(0)
		if total := hits + misses; total > 0 {
			hitRatio = float64(hits) / float64(total)
		}
		connCount := 0
		if a.server != nil {
			connCount = a.server.GetConnectionCount()
		}
		wailsRuntime.EventsEmit(a.ctx, "bandwidth:update", map[string]interface{}{
			"bytesPerSecond":    bps,
			"residentialBps":    resBPS,
			"totalToday":        totalToday,
			"residentialToday":  residentToday,
			"cacheHitRatio":     hitRatio,
			"activeConnections": connCount,
		})
	}
}

func (a *App) reloadClassifier() {
	if a.db == nil || a.classifier == nil {
		return
	}
	rules, err := classifier.LoadRulesFromDB(a.db.Reader)
	if err != nil {
		log.Printf("Failed to load rules: %v", err)
		return
	}
	a.classifier.Reload(rules)
	log.Printf("Classifier reloaded with %d rules", len(rules))
}

func (a *App) loadSettingsIntoConfig() {
	if a.db == nil {
		return
	}
	a.cfg.HTTPPort = a.db.GetSettingInt("http_port", 8888)
	a.cfg.SOCKS5Port = a.db.GetSettingInt("socks5_port", 8889)
	a.cfg.MaxConcurrent = a.db.GetSettingInt("max_concurrent", 500)
	a.cfg.CacheMemoryMB = a.db.GetSettingInt("cache_memory_mb", 512)
	a.cfg.CacheDiskMB = a.db.GetSettingInt("cache_disk_mb", 2048)
	a.cfg.MITMEnabled = a.db.GetSettingBool("mitm_enabled", false)
	a.cfg.AcceptEncodingEnforce = a.db.GetSettingBool("accept_encoding_enforce", true)
	a.cfg.HeaderStripping = a.db.GetSettingBool("header_stripping", true)
	a.cfg.HTMLMinification = a.db.GetSettingBool("html_minification", false)
	a.cfg.ImageRecompression = a.db.GetSettingBool("image_recompression", false)
	a.cfg.ImageQuality = a.db.GetSettingInt("image_quality", 85)
	a.cfg.LogRetentionDays = a.db.GetSettingInt("log_retention_days", 7)
	a.cfg.SingleflightDedup = a.db.GetSettingBool("singleflight_dedup", true)

	if addr, err := a.db.GetSetting("bind_address"); err == nil && addr != "" {
		a.cfg.BindAddress = addr
	}
	if level, err := a.db.GetSetting("log_level"); err == nil && level != "" {
		a.cfg.LogLevel = level
	}

	// Auth settings
	a.cfg.ProxyAuthEnabled = a.db.GetSettingBool("proxy_auth_enabled", false)
	if u, err := a.db.GetSetting("proxy_username"); err == nil && u != "" {
		a.cfg.ProxyUsername = u
	}
	if p, err := a.db.GetSetting("proxy_password"); err == nil && p != "" {
		a.cfg.ProxyPassword = p
	}
	a.cfg.IPWhitelistEnabled = a.db.GetSettingBool("ip_whitelist_enabled", false)
	if wl, err := a.db.GetSetting("ip_whitelist"); err == nil {
		a.cfg.IPWhitelist = wl
	}
}

func (a *App) configureProxyAuth() {
	if a.proxyAuth == nil {
		return
	}
	a.proxyAuth.Configure(
		a.cfg.ProxyAuthEnabled, a.cfg.ProxyUsername, a.cfg.ProxyPassword,
		a.cfg.IPWhitelistEnabled, a.cfg.IPWhitelist,
	)
}
