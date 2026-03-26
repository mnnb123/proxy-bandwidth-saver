package webapi

import "net/http"

func registerSettingsRoutes(mux *http.ServeMux, app AppBackend) {
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

	// CA cert
	mux.HandleFunc("/api/cert/path", cors(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]string{"path": app.GetCACertPath()})
	}))
}
