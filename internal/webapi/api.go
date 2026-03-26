package webapi

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
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

	// Cache
	GetCacheStatsJSON() interface{}
	ClearCache() error

	// Settings
	GetSettingsJSON() interface{}
	UpdateSetting(key, value string) error

	// Certs
	GetCACertPath() string
}

// Router creates the HTTP API handler.
func Router(app AppBackend) http.Handler {
	mux := http.NewServeMux()

	// Proxy control
	mux.HandleFunc("/api/proxy/start", cors(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", 405)
			return
		}
		if err := app.StartProxy(); err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, map[string]bool{"ok": true})
	}))

	mux.HandleFunc("/api/proxy/stop", cors(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", 405)
			return
		}
		if err := app.StopProxy(); err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, map[string]bool{"ok": true})
	}))

	mux.HandleFunc("/api/proxy/status", cors(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, app.GetProxyStatusJSON())
	}))

	// Rules
	mux.HandleFunc("/api/rules", cors(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, app.GetRulesJSON())
		case http.MethodPost:
			var body struct {
				RuleType string `json:"ruleType"`
				Pattern  string `json:"pattern"`
				Action   string `json:"action"`
				Priority int    `json:"priority"`
			}
			if err := readJSON(r, &body); err != nil {
				writeError(w, err)
				return
			}
			if err := app.AddRule(body.RuleType, body.Pattern, body.Action, body.Priority); err != nil {
				writeError(w, err)
				return
			}
			writeJSON(w, map[string]bool{"ok": true})
		default:
			http.Error(w, "method not allowed", 405)
		}
	}))

	mux.HandleFunc("/api/rules/", cors(func(w http.ResponseWriter, r *http.Request) {
		idStr := strings.TrimPrefix(r.URL.Path, "/api/rules/")
		parts := strings.Split(idStr, "/")
		id, err := strconv.Atoi(parts[0])
		if err != nil {
			writeError(w, fmt.Errorf("invalid id"))
			return
		}

		switch r.Method {
		case http.MethodPut:
			var body struct {
				RuleType string `json:"ruleType"`
				Pattern  string `json:"pattern"`
				Action   string `json:"action"`
				Priority int    `json:"priority"`
				Enabled  bool   `json:"enabled"`
			}
			if err := readJSON(r, &body); err != nil {
				writeError(w, err)
				return
			}
			if err := app.UpdateRuleById(id, body.RuleType, body.Pattern, body.Action, body.Priority, body.Enabled); err != nil {
				writeError(w, err)
				return
			}
			writeJSON(w, map[string]bool{"ok": true})
		case http.MethodDelete:
			if err := app.DeleteRule(id); err != nil {
				writeError(w, err)
				return
			}
			writeJSON(w, map[string]bool{"ok": true})
		default:
			http.Error(w, "method not allowed", 405)
		}
	}))

	mux.HandleFunc("/api/rules/toggle", cors(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", 405)
			return
		}
		var body struct {
			ID      int  `json:"id"`
			Enabled bool `json:"enabled"`
		}
		if err := readJSON(r, &body); err != nil {
			writeError(w, err)
			return
		}
		if err := app.ToggleRule(body.ID, body.Enabled); err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, map[string]bool{"ok": true})
	}))

	mux.HandleFunc("/api/rules/test", cors(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", 405)
			return
		}
		var body struct {
			Domain      string `json:"domain"`
			URL         string `json:"url"`
			ContentType string `json:"contentType"`
		}
		if err := readJSON(r, &body); err != nil {
			writeError(w, err)
			return
		}
		result := app.TestRule(body.Domain, body.URL, body.ContentType)
		writeJSON(w, map[string]string{"result": result})
	}))

	mux.HandleFunc("/api/rules/import", cors(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", 405)
			return
		}
		var body struct {
			Data string `json:"data"`
		}
		if err := readJSON(r, &body); err != nil {
			writeError(w, err)
			return
		}
		count := app.ImportRules(body.Data)
		writeJSON(w, map[string]int{"count": count})
	}))

	mux.HandleFunc("/api/rules/export", cors(func(w http.ResponseWriter, r *http.Request) {
		data := app.ExportRules()
		writeJSON(w, map[string]string{"data": data})
	}))

	// Proxies
	mux.HandleFunc("/api/proxies", cors(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, app.GetProxiesJSON())
		case http.MethodPost:
			var body struct {
				Address  string `json:"address"`
				Username string `json:"username"`
				Password string `json:"password"`
				Type     string `json:"type"`
				Category string `json:"category"`
			}
			if err := readJSON(r, &body); err != nil {
				writeError(w, err)
				return
			}
			if err := app.AddProxy(body.Address, body.Username, body.Password, body.Type, body.Category); err != nil {
				writeError(w, err)
				return
			}
			writeJSON(w, map[string]bool{"ok": true})
		default:
			http.Error(w, "method not allowed", 405)
		}
	}))

	mux.HandleFunc("/api/proxies/", cors(func(w http.ResponseWriter, r *http.Request) {
		idStr := strings.TrimPrefix(r.URL.Path, "/api/proxies/")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			writeError(w, fmt.Errorf("invalid id"))
			return
		}
		if r.Method != http.MethodDelete {
			http.Error(w, "method not allowed", 405)
			return
		}
		if err := app.DeleteProxy(id); err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, map[string]bool{"ok": true})
	}))

	mux.HandleFunc("/api/proxies/import", cors(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", 405)
			return
		}
		var body struct {
			Data string `json:"data"`
		}
		if err := readJSON(r, &body); err != nil {
			writeError(w, err)
			return
		}
		count := app.ImportProxies(body.Data)
		writeJSON(w, map[string]int{"count": count})
	}))

	mux.HandleFunc("/api/proxies/output", cors(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, app.GetOutputProxiesJSON())
	}))

	// Stats
	mux.HandleFunc("/api/stats/realtime", cors(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, app.GetRealtimeStatsJSON())
	}))

	mux.HandleFunc("/api/stats/cost", cors(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, app.GetCostSummaryJSON())
	}))

	mux.HandleFunc("/api/stats/budget", cors(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, app.GetBudgetStatusJSON())
	}))

	mux.HandleFunc("/api/stats/domains", cors(func(w http.ResponseWriter, r *http.Request) {
		period := r.URL.Query().Get("period")
		if period == "" {
			period = "24h"
		}
		writeJSON(w, app.GetDomainStatsJSON(period))
	}))

	// Cache
	mux.HandleFunc("/api/cache/stats", cors(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, app.GetCacheStatsJSON())
	}))

	mux.HandleFunc("/api/cache/clear", cors(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", 405)
			return
		}
		if err := app.ClearCache(); err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, map[string]bool{"ok": true})
	}))

	// Settings
	mux.HandleFunc("/api/settings", cors(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, app.GetSettingsJSON())
		case http.MethodPut:
			var body struct {
				Key   string `json:"key"`
				Value string `json:"value"`
			}
			if err := readJSON(r, &body); err != nil {
				writeError(w, err)
				return
			}
			if err := app.UpdateSetting(body.Key, body.Value); err != nil {
				writeError(w, err)
				return
			}
			writeJSON(w, map[string]bool{"ok": true})
		default:
			http.Error(w, "method not allowed", 405)
		}
	}))

	// CA cert
	mux.HandleFunc("/api/cert/path", cors(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]string{"path": app.GetCACertPath()})
	}))

	return mux
}

func cors(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
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
