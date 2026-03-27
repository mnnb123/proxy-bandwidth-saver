package webapi

import (
	"encoding/json"
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
)

// HeadlessApp is the backend app without Wails dependency.
type HeadlessApp struct {
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
	events     *EventBroker
	stopCh     chan struct{}
}

func NewHeadlessApp(cfg *config.Config, events *EventBroker) *HeadlessApp {
	return &HeadlessApp{
		cfg:    cfg,
		events: events,
		stopCh: make(chan struct{}),
	}
}

func (a *HeadlessApp) Start() error {
	if err := config.EnsureDataDirs(a.cfg); err != nil {
		return fmt.Errorf("data dirs: %w", err)
	}

	db, err := database.NewDB(a.cfg.DBPath)
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	a.db = db

	if err := a.db.SeedDefaults(); err != nil {
		log.Printf("Failed to seed defaults: %v", err)
	}
	if err := classifier.SeedDefaultRules(a.db.Writer); err != nil {
		log.Printf("Failed to seed default rules: %v", err)
	}

	LoadSettingsIntoConfig(a.db, a.cfg)

	a.classifier = classifier.NewClassifier()
	ReloadClassifier(a.db.Reader, a.classifier)

	cacheLayer, err := cache.NewCacheLayer(a.cfg.CacheDir, a.cfg.CacheMemoryMB, a.cfg.CacheDiskMB)
	if err != nil {
		log.Printf("Cache init failed: %v", err)
	} else {
		a.cache = cacheLayer
	}

	a.meter = meter.NewMeter(a.db.Writer)
	a.meter.Start()

	a.upstream = upstream.NewManager(a.db.Reader)
	a.upstream.LoadProxies()
	a.upstream.StartHealthChecks(60 * time.Second)

	a.proxyAuth = proxy.NewProxyAuth()
	ConfigureProxyAuth(a.proxyAuth, a.cfg)
	basePort := a.db.GetSettingInt("base_port", 30000)
	var classifyFunc proxy.ClassifyFunc
	if a.classifier != nil {
		classifyFunc = a.classifier.Classify
	}
	a.portMapper = proxy.NewPortMapper("0.0.0.0", basePort, a.proxyAuth, func(domain string, reqBytes, respBytes int64, proxyID int, route string) {
		if a.meter != nil {
			a.meter.Record(meter.RequestLog{
				Timestamp:     time.Now(),
				Domain:        domain,
				Route:         route,
				RequestBytes:  reqBytes,
				ResponseBytes: respBytes,
				ProxyID:       proxyID,
			})
		}
	}, classifyFunc)
	a.remapAllProxies()

	a.optCfg = &optimizer.Config{
		AcceptEncodingEnforce: a.cfg.AcceptEncodingEnforce,
		HeaderStripping:       a.cfg.HeaderStripping,
		HTMLMinification:      a.cfg.HTMLMinification,
		ImageRecompression:    a.cfg.ImageRecompression,
		ImageQuality:          a.cfg.ImageQuality,
	}

	// Main proxy (8888/8889) removed — only output ports (30000+) are used.

	go a.emitRealtimeStats()
	go a.autoClearLoop()

	log.Println("Headless app started")
	return nil
}

func (a *HeadlessApp) Shutdown() {
	close(a.stopCh)
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
	log.Println("Headless app shutdown")
}

// === Proxy Control ===

func (a *HeadlessApp) StartProxy() error {
	if a.server != nil && a.server.IsRunning() {
		return fmt.Errorf("proxy already running")
	}

	ConfigureProxyAuth(a.proxyAuth, a.cfg)

	a.server = proxy.NewProxyServer(proxy.ServerConfig{
		HTTPPort:    a.cfg.HTTPPort,
		SOCKS5Port:  a.cfg.SOCKS5Port,
		BindAddress: "0.0.0.0",
		MITMEnabled: a.cfg.MITMEnabled,
		CertDir:     a.cfg.CertDir,
	})
	a.server.GetAuth().Configure(
		a.cfg.ProxyAuthEnabled, a.cfg.ProxyUsername, a.cfg.ProxyPassword,
		a.cfg.IPWhitelistEnabled, a.cfg.IPWhitelist,
	)

	a.server.SetPipeline(BuildPipeline(a.classifier, a.cache, a.meter))
	return a.server.Start()
}

func (a *HeadlessApp) StopProxy() error {
	if a.server == nil || !a.server.IsRunning() {
		return fmt.Errorf("proxy not running")
	}
	return a.server.Stop()
}

func (a *HeadlessApp) GetProxyStatusJSON() interface{} {
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

func (a *HeadlessApp) emitRealtimeStats() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-a.stopCh:
			return
		case <-ticker.C:
			if a.meter == nil {
				continue
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
			a.events.Emit("bandwidth:update", map[string]interface{}{
				"bytesPerSecond":    bps,
				"residentialBps":    resBPS,
				"totalToday":        totalToday,
				"residentialToday":  residentToday,
				"cacheHitRatio":     hitRatio,
				"activeConnections": connCount,
			})
		}
	}
}

// reloadClassifier delegates to the shared ReloadClassifier in appcore.go.
func (a *HeadlessApp) reloadClassifier() {
	ReloadClassifier(a.db.Reader, a.classifier)
}

// loadSettingsIntoConfig delegates to the shared LoadSettingsIntoConfig in appcore.go.
func (a *HeadlessApp) loadSettingsIntoConfig() {
	LoadSettingsIntoConfig(a.db, a.cfg)
}

// configureProxyAuth delegates to the shared ConfigureProxyAuth in appcore.go.
func (a *HeadlessApp) configureProxyAuth() {
	ConfigureProxyAuth(a.proxyAuth, a.cfg)
}

func (a *HeadlessApp) remapAllProxies() {
	if a.portMapper == nil || a.db == nil {
		return
	}
	proxies := a.getProxies()
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

// GetWebCredentials returns the current web panel login credentials.
func (a *HeadlessApp) GetWebCredentials() (string, string) {
	if a.db == nil {
		return "", ""
	}
	user, _ := a.db.GetSetting("web_username")
	pass, _ := a.db.GetSetting("web_password")
	return user, pass
}

// autoClearLoop periodically clears old domain stats based on the setting.
func (a *HeadlessApp) autoClearLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-a.stopCh:
			return
		case <-ticker.C:
			if a.db == nil {
				continue
			}
			minutes := a.db.GetSettingInt("domain_report_clear_minutes", 0)
			if minutes > 0 {
				a.AutoClearDomainStats(minutes)
			}
		}
	}
}

// JSON helpers
func jsonMarshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func jsonUnmarshal(s string, v interface{}) error {
	return json.Unmarshal([]byte(s), v)
}
