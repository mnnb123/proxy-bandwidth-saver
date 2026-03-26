package webapi

import "net/http"

func registerProxyRoutes(mux *http.ServeMux, app AppBackend) {
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
}
