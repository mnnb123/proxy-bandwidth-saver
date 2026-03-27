package webapi

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func registerProxiesRoutes(mux *http.ServeMux, app AppBackend) {
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

	mux.HandleFunc("/api/proxies/clear", cors(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "method not allowed", 405)
			return
		}
		if err := app.ClearAllProxies(); err != nil {
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
}
