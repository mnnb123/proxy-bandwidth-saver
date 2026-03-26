package webapi

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

// AppBackend is the interface the API layer needs from the app.
type AppBackend interface {
	// Proxy control
	StartProxy() error
	StopProxy() error
	GetProxyStatusJSON() interface{}

	// Rules
	GetRulesJSON() interface{}
	AddRule(ruleType, pattern, action string, priority int) error
	AddBulkRules(patterns []string, action string, priority int) (int, error)
	ClearAllRules() error
	UpdateRuleById(id int, ruleType, pattern, action string, priority int, enabled bool) error
	DeleteRule(id int) error
	ToggleRule(id int, enabled bool) error
	TestRule(domain, urlPath, contentType string) string
	ImportRules(jsonStr string) int
	ExportRules() string

	// Proxies
	GetProxiesJSON() interface{}
	AddProxy(address, username, password, proxyType, category string) error
	DeleteProxy(id int) error
	ImportProxies(text string) int
	GetOutputProxiesJSON() interface{}

	// Stats
	GetRealtimeStatsJSON() interface{}
	GetCostSummaryJSON() interface{}
	GetBudgetStatusJSON() interface{}
	GetDomainStatsJSON(period string) interface{}
	GetDomainStatsByPortJSON(period string, proxyID int) interface{}
	ClearDomainStats() error

	// Cache
	GetCacheStatsJSON() interface{}
	ClearCache() error

	// Settings
	GetSettingsJSON() interface{}
	UpdateSetting(key, value string) error

	// Certs
	GetCACertPath() string

	// PAC
	GeneratePAC(proxyAddr string) string
}

// AppVersion is the application version string, set at build time or here.
var AppVersion = "1.1.0"

// Router creates the HTTP API handler.
func Router(app AppBackend) http.Handler {
	mux := http.NewServeMux()

	// Version
	mux.HandleFunc("/api/version", cors(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]string{"version": AppVersion})
	}))

	// Register route groups
	registerProxyRoutes(mux, app)
	registerRulesRoutes(mux, app)
	registerProxiesRoutes(mux, app)
	registerStatsRoutes(mux, app)
	registerSettingsRoutes(mux, app)

	return mux
}

func cors(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(204)
			return
		}
		h(w, r)
	}
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("JSON encode error: %v", err)
	}
}

func writeError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(400)
	json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}

func readJSON(r *http.Request, v interface{}) error {
	body, err := io.ReadAll(io.LimitReader(r.Body, 10<<20)) // 10MB limit
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}
