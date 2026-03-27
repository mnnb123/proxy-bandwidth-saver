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
	"syscall"
	"time"

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

	server := &http.Server{
		Addr:              addr,
		Handler:           webAuth.Wrap(mainMux),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
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
