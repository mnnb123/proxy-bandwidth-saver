package webapi

import (
	"fmt"
	"log"

	"proxy-bandwidth-saver/internal/config"
	"proxy-bandwidth-saver/internal/database"
	"proxy-bandwidth-saver/internal/proxy"
)

func (a *HeadlessApp) GetRealtimeStatsJSON() interface{} {
	if a.meter == nil {
		return database.RealtimeStats{}
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
	return database.RealtimeStats{
		BytesPerSecond:    bps,
		ResidentialBPS:    resBPS,
		TotalToday:        totalToday,
		ResidentialToday:  residentToday,
		CostToday:         config.BytesCost(residentToday, a.cfg.CostPerGB),
		CacheHitRatio:     hitRatio,
		ActiveConnections: connCount,
	}
}

func (a *HeadlessApp) GetCostSummaryJSON() interface{} {
	if a.meter == nil {
		return database.CostSummary{}
	}
	_, residentToday, _, _, _, _ := a.meter.GetRealtimeStats()
	savedCache, savedDirect, savedDC, savedComp := a.meter.GetSavings()
	totalSaved := savedCache + savedDirect + savedDC + savedComp
	return database.CostSummary{
		CostToday:  config.BytesCost(residentToday, a.cfg.CostPerGB),
		SavedBytes: totalSaved,
		SavedCost:  config.BytesCost(totalSaved, a.cfg.CostPerGB),
	}
}

func (a *HeadlessApp) GetBudgetStatusJSON() interface{} {
	status := database.BudgetStatus{
		MonthlyBudgetGB: a.cfg.MonthlyBudgetGB,
		CostPerGB:       a.cfg.CostPerGB,
	}
	if a.meter != nil {
		_, residentToday, _, _, _, _ := a.meter.GetRealtimeStats()
		usedGB := config.BytesToGB(residentToday)
		status.UsedGB = usedGB
		if a.cfg.MonthlyBudgetGB > 0 {
			status.UsedPercent = usedGB / a.cfg.MonthlyBudgetGB * 100
		}
		status.RemainingGB = a.cfg.MonthlyBudgetGB - usedGB
	}
	return status
}

func (a *HeadlessApp) GetCacheStatsJSON() interface{} {
	if a.cache == nil {
		return database.CacheStats{}
	}
	hits, misses, bytesSaved, hitRatio, _, diskMB := a.cache.GetStats()
	return database.CacheStats{
		HitCount:   hits,
		MissCount:  misses,
		BytesSaved: bytesSaved,
		HitRatio:   hitRatio,
		DiskUsedMB: diskMB,
	}
}

func (a *HeadlessApp) ClearCache() error {
	if a.cache == nil {
		return fmt.Errorf("cache not initialized")
	}
	a.cache.Clear()
	return nil
}

func (a *HeadlessApp) GetSettingsJSON() interface{} {
	if a.db == nil {
		return map[string]string{}
	}
	settings, err := a.db.GetAllSettings()
	if err != nil {
		log.Printf("Failed to get settings: %v", err)
		return map[string]string{}
	}
	return settings
}

// allowedSettingKeys is the allowlist of setting keys that can be updated via API.
// allowedSettingKeys is the allowlist of setting keys that can be updated via API.
var allowedSettingKeys = map[string]bool{
	"proxy_auth_enabled": true, "proxy_username": true, "proxy_password": true,
	"ip_whitelist_enabled": true, "ip_whitelist": true,
	"web_username": true, "web_password": true,
	"accept_encoding_enforce": true, "header_stripping": true,
	"html_minification": true, "image_recompression": true, "image_quality": true,
	"cache_enabled": true, "cache_max_size_mb": true, "cache_ttl_minutes": true,
	"budget_daily_gb": true, "budget_monthly_gb": true, "cost_per_gb": true,
	"auto_clear_minutes": true, "log_retention_days": true,
	"mitm_enabled": true,
}

func (a *HeadlessApp) UpdateSetting(key, value string) error {
	if a.db == nil {
		return fmt.Errorf("not initialized")
	}
	if !allowedSettingKeys[key] {
		return fmt.Errorf("setting key not allowed: %s", key)
	}
	if err := a.db.SetSetting(key, value); err != nil {
		return err
	}
	a.loadSettingsIntoConfig()
	if a.optCfg != nil {
		a.optCfg.AcceptEncodingEnforce = a.cfg.AcceptEncodingEnforce
		a.optCfg.HeaderStripping = a.cfg.HeaderStripping
		a.optCfg.HTMLMinification = a.cfg.HTMLMinification
		a.optCfg.ImageRecompression = a.cfg.ImageRecompression
		a.optCfg.ImageQuality = a.cfg.ImageQuality
	}
	a.configureProxyAuth()
	return nil
}

func (a *HeadlessApp) GetCACertPath() string {
	return proxy.GetCACertPath(a.cfg.CertDir)
}
