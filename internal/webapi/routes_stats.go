package webapi

import (
	"net/http"
	"strconv"
)

func registerStatsRoutes(mux *http.ServeMux, app AppBackend) {
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
		proxyIDStr := r.URL.Query().Get("proxyId")
		proxyID := 0
		if proxyIDStr != "" {
			proxyID, _ = strconv.Atoi(proxyIDStr)
		}
		if proxyID > 0 {
			writeJSON(w, app.GetDomainStatsByPortJSON(period, proxyID))
		} else {
			writeJSON(w, app.GetDomainStatsJSON(period))
		}
	}))

	mux.HandleFunc("/api/stats/domains/clear", cors(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", 405)
			return
		}
		if err := app.ClearDomainStats(); err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, map[string]bool{"ok": true})
	}))
}
