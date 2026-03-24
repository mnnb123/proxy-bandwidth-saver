package config

import (
	"os"
	"path/filepath"
	"runtime"
)

type Config struct {
	// Proxy server
	HTTPPort  int `json:"httpPort"`
	SOCKS5Port int `json:"socks5Port"`
	BindAddress string `json:"bindAddress"`
	MaxConcurrent int `json:"maxConcurrent"`

	// Cache
	CacheMemoryMB int `json:"cacheMemoryMB"`
	CacheDiskMB   int `json:"cacheDiskMB"`

	// Budget
	MonthlyBudgetGB float64 `json:"monthlyBudgetGB"`
	CostPerGB       float64 `json:"costPerGB"`
	AutoPause       bool    `json:"autoPause"`
	AlertThreshold80 bool   `json:"alertThreshold80"`
	AlertThreshold95 bool   `json:"alertThreshold95"`

	// MITM
	MITMEnabled bool `json:"mitmEnabled"`

	// Optimization
	AcceptEncodingEnforce bool `json:"acceptEncodingEnforce"`
	HeaderStripping       bool `json:"headerStripping"`
	HTMLMinification      bool `json:"htmlMinification"`
	ImageRecompression    bool `json:"imageRecompression"`
	ImageQuality          int  `json:"imageQuality"`

	// Logging
	LogRetentionDays int    `json:"logRetentionDays"`
	LogLevel         string `json:"logLevel"`

	// Proxy Authentication
	ProxyAuthEnabled bool   `json:"proxyAuthEnabled"`
	ProxyUsername    string `json:"proxyUsername"`
	ProxyPassword    string `json:"proxyPassword"`
	IPWhitelistEnabled bool `json:"ipWhitelistEnabled"`
	IPWhitelist      string `json:"ipWhitelist"` // comma-separated IPs/CIDRs

	// Advanced
	RequestTimeoutSec int  `json:"requestTimeoutSec"`
	DNSCacheTTLSec    int  `json:"dnsCacheTTLSec"`
	SingleflightDedup bool `json:"singleflightDedup"`

	// Paths
	DataDir string `json:"-"`
	DBPath  string `json:"-"`
	CacheDir string `json:"-"`
	CertDir  string `json:"-"`
}

func DefaultConfig() *Config {
	dataDir := defaultDataDir()
	return &Config{
		HTTPPort:       8888,
		SOCKS5Port:     8889,
		BindAddress:    "127.0.0.1",
		MaxConcurrent:  500,

		CacheMemoryMB: 512,
		CacheDiskMB:   2048,

		MonthlyBudgetGB:  0,
		CostPerGB:        0,
		AutoPause:        false,
		AlertThreshold80: true,
		AlertThreshold95: true,

		MITMEnabled: false,

		ProxyAuthEnabled:   false,
		ProxyUsername:      "",
		ProxyPassword:      "",
		IPWhitelistEnabled: false,
		IPWhitelist:        "",

		AcceptEncodingEnforce: true,
		HeaderStripping:       true,
		HTMLMinification:      false,
		ImageRecompression:    false,
		ImageQuality:          85,

		LogRetentionDays: 7,
		LogLevel:         "normal",

		RequestTimeoutSec: 30,
		DNSCacheTTLSec:    300,
		SingleflightDedup: true,

		DataDir:  dataDir,
		DBPath:   filepath.Join(dataDir, "proxy-bandwidth-saver.db"),
		CacheDir: filepath.Join(dataDir, "cache"),
		CertDir:  filepath.Join(dataDir, "certs"),
	}
}

func defaultDataDir() string {
	if dir := os.Getenv("PBS_DATA_DIR"); dir != "" {
		return dir
	}
	switch runtime.GOOS {
	case "linux":
		// XDG or /var/lib for system service
		if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
			return filepath.Join(xdg, "proxy-bandwidth-saver")
		}
		home, _ := os.UserHomeDir()
		if home != "" {
			return filepath.Join(home, ".local", "share", "proxy-bandwidth-saver")
		}
		return "/var/lib/proxy-bandwidth-saver"
	default: // windows
		appData := os.Getenv("APPDATA")
		if appData == "" {
			home, _ := os.UserHomeDir()
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		return filepath.Join(appData, "ProxyBandwidthSaver")
	}
}

// BytesToGB converts bytes to gigabytes.
func BytesToGB(bytes int64) float64 {
	return float64(bytes) / (1024 * 1024 * 1024)
}

// BytesCost calculates cost for a given number of bytes at a per-GB rate.
func BytesCost(bytes int64, costPerGB float64) float64 {
	return BytesToGB(bytes) * costPerGB
}

func EnsureDataDirs(cfg *Config) error {
	dirs := []string{cfg.DataDir, cfg.CacheDir, cfg.CertDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return err
		}
	}
	return nil
}
