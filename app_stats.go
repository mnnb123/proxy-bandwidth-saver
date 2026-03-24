package main

import (
	"fmt"
	"log"

	"proxy-bandwidth-saver/internal/config"
	"proxy-bandwidth-saver/internal/database"
	"proxy-bandwidth-saver/internal/proxy"
)

// === Bandwidth & Cost ===

func (a *App) GetRealtimeStats() database.RealtimeStats {
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

func (a *App) GetCostSummary() database.CostSummary {
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

func (a *App) GetBudgetStatus() database.BudgetStatus {
	if a.meter == nil {
		return database.BudgetStatus{
			MonthlyBudgetGB: a.cfg.MonthlyBudgetGB,
			CostPerGB:       a.cfg.CostPerGB,
		}
	}
	_, residentToday, _, _, _, _ := a.meter.GetRealtimeStats()
	usedGB := config.BytesToGB(residentToday)
	usedPct := float64(0)
	if a.cfg.MonthlyBudgetGB > 0 {
		usedPct = usedGB / a.cfg.MonthlyBudgetGB * 100
	}
	return database.BudgetStatus{
		MonthlyBudgetGB: a.cfg.MonthlyBudgetGB,
		UsedGB:          usedGB,
		UsedPercent:     usedPct,
		RemainingGB:     a.cfg.MonthlyBudgetGB - usedGB,
		CostPerGB:       a.cfg.CostPerGB,
	}
}

// === Cache ===

func (a *App) GetCacheStats() database.CacheStats {
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

func (a *App) ClearCache() error {
	if a.cache == nil {
		return fmt.Errorf("cache not initialized")
	}
	a.cache.Clear()
	return nil
}

// === Settings ===

func (a *App) GetSettings() map[string]string {
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

func (a *App) UpdateSetting(key, value string) error {
	if a.db == nil {
		return fmt.Errorf("database not initialized")
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

func (a *App) GetCACertPath() string {
	return proxy.GetCACertPath(a.cfg.CertDir)
}
