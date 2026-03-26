package webapi

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func registerRulesRoutes(mux *http.ServeMux, app AppBackend) {
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

	mux.HandleFunc("/api/rules/clear", cors(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", 405)
			return
		}
		if err := app.ClearAllRules(); err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, map[string]bool{"ok": true})
	}))

	mux.HandleFunc("/api/rules/bulk", cors(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", 405)
			return
		}
		var body struct {
			Patterns []string `json:"patterns"`
			Action   string   `json:"action"`
			Priority int      `json:"priority"`
		}
		if err := readJSON(r, &body); err != nil {
			writeError(w, err)
			return
		}
		count, err := app.AddBulkRules(body.Patterns, body.Action, body.Priority)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, map[string]int{"count": count})
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
}
