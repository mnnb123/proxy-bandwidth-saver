package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"proxy-bandwidth-saver/internal/config"
	"proxy-bandwidth-saver/internal/webapi"
)

//go:embed all:frontend
var frontendFS embed.FS

func main() {
	webPort := flag.Int("web-port", 8080, "Web admin panel port")
	dataDir := flag.String("data-dir", "", "Data directory (default: ~/.proxy-bandwidth-saver)")
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Proxy Bandwidth Saver - Headless Server")

	cfg := config.DefaultConfig()
	if *dataDir != "" {
		cfg.DataDir = *dataDir
		cfg.DBPath = *dataDir + "/proxy-bw-saver.db"
		cfg.CacheDir = *dataDir + "/cache"
		cfg.CertDir = *dataDir + "/certs"
	}

	// SSE event broker
	events := webapi.NewEventBroker()

	// Create and start headless app
	app := webapi.NewHeadlessApp(cfg, events)
	if err := app.Start(); err != nil {
		log.Fatalf("Failed to start: %v", err)
	}

	// API router
	apiMux := webapi.Router(app)

	// Web panel auth middleware
	webAuth := webapi.NewWebAuth(app.GetWebCredentials)

	// Main HTTP server mux
	mainMux := http.NewServeMux()

	// PAC file endpoint (no auth - browsers need to fetch this before proxy config)
	mainMux.HandleFunc("/proxy.pac", func(w http.ResponseWriter, r *http.Request) {
		proxyAddr := r.URL.Query().Get("proxy")
		if proxyAddr == "" {
			// Default: use this server's address with main HTTP proxy port
			host := r.Host
			if idx := strings.Index(host, ":"); idx >= 0 {
				host = host[:idx]
			}
			proxyAddr = fmt.Sprintf("%s:%d", host, cfg.HTTPPort)
		}
		w.Header().Set("Content-Type", "application/x-ns-proxy-autoconfig")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write([]byte(app.GeneratePAC(proxyAddr)))
	})

	// SSE events endpoint
	mainMux.Handle("/api/events", events)

	// API routes
	mainMux.Handle("/api/", apiMux)

	// Serve embedded frontend (SPA)
	frontendDist, err := fs.Sub(frontendFS, "frontend")
	if err != nil {
		log.Fatalf("Failed to load frontend: %v", err)
	}
	fileServer := http.FileServer(http.FS(frontendDist))
	mainMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Try serving static file first
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}
		if _, err := fs.Stat(frontendDist, path[1:]); err == nil {
			fileServer.ServeHTTP(w, r)
			return
		}
		// SPA fallback: serve index.html for all non-file routes
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})

	addr := fmt.Sprintf("0.0.0.0:%d", *webPort)
	log.Printf("Web admin panel: http://%s", addr)
	log.Printf("HTTP Proxy: 0.0.0.0:%d", cfg.HTTPPort)
	log.Printf("SOCKS5 Proxy: 0.0.0.0:%d", cfg.SOCKS5Port)

	server := &http.Server{
		Addr:    addr,
		Handler: webAuth.Wrap(mainMux),
	}

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Web server error: %v", err)
		}
	}()

	<-sigCh
	log.Println("Shutting down...")
	app.Shutdown()
	server.Close()
}
