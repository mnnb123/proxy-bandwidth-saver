package proxy

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/elazarl/goproxy"
)

type ServerState int32

const (
	StateStopped  ServerState = 0
	StateStarting ServerState = 1
	StateRunning  ServerState = 2
	StateStopping ServerState = 3
)

type ProxyServer struct {
	httpPort   int
	socks5Port int
	bindAddr   string

	state      atomic.Int32
	pipeline   *Pipeline
	transport  *TransportManager
	auth       *ProxyAuth

	httpProxy    *goproxy.ProxyHttpServer
	httpListener net.Listener
	socks5Listener net.Listener

	connCount  atomic.Int32
	startTime  time.Time

	mitmEnabled bool
	caCert      *tls.Certificate
	certDir     string

	mu         sync.Mutex
	cancelFunc context.CancelFunc
}

type ServerConfig struct {
	HTTPPort    int
	SOCKS5Port  int
	BindAddress string
	MITMEnabled bool
	CertDir     string
}

func NewProxyServer(cfg ServerConfig) *ProxyServer {
	return &ProxyServer{
		httpPort:    cfg.HTTPPort,
		socks5Port:  cfg.SOCKS5Port,
		bindAddr:    cfg.BindAddress,
		mitmEnabled: cfg.MITMEnabled,
		certDir:     cfg.CertDir,
		pipeline:    NewDefaultPipeline(),
		transport:   NewTransportManager(),
		auth:        NewProxyAuth(),
	}
}

// GetAuth returns the proxy auth instance for external configuration.
func (s *ProxyServer) GetAuth() *ProxyAuth {
	return s.auth
}

// SetPipeline replaces pipeline hooks
func (s *ProxyServer) SetPipeline(p *Pipeline) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pipeline = p
}

// GetTransportManager returns the transport manager for external configuration
func (s *ProxyServer) GetTransportManager() *TransportManager {
	return s.transport
}

// Start launches both HTTP and SOCKS5 proxy servers
func (s *ProxyServer) Start() error {
	if !s.state.CompareAndSwap(int32(StateStopped), int32(StateStarting)) {
		return fmt.Errorf("proxy already running or starting")
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.cancelFunc = cancel

	// Setup MITM if enabled
	if s.mitmEnabled && s.certDir != "" {
		cert, err := LoadOrCreateCA(s.certDir)
		if err != nil {
			s.state.Store(int32(StateStopped))
			return fmt.Errorf("load CA: %w", err)
		}
		s.caCert = cert
	}

	// Setup HTTP proxy
	if err := s.setupHTTPProxy(); err != nil {
		s.state.Store(int32(StateStopped))
		return fmt.Errorf("setup HTTP proxy: %w", err)
	}

	// Start HTTP listener
	httpAddr := fmt.Sprintf("%s:%d", s.bindAddr, s.httpPort)
	httpLn, err := net.Listen("tcp", httpAddr)
	if err != nil {
		s.state.Store(int32(StateStopped))
		return fmt.Errorf("listen HTTP %s: %w", httpAddr, err)
	}
	s.httpListener = httpLn

	// Start SOCKS5 listener
	socks5Addr := fmt.Sprintf("%s:%d", s.bindAddr, s.socks5Port)
	socks5Ln, err := net.Listen("tcp", socks5Addr)
	if err != nil {
		httpLn.Close()
		s.state.Store(int32(StateStopped))
		return fmt.Errorf("listen SOCKS5 %s: %w", socks5Addr, err)
	}
	s.socks5Listener = socks5Ln

	// Start serving
	s.startTime = time.Now()

	go func() {
		srv := &http.Server{
			Handler:           s.httpProxy,
			ReadHeaderTimeout: 10 * time.Second,
			IdleTimeout:       120 * time.Second,
		}
		if err := srv.Serve(httpLn); err != nil && ctx.Err() == nil {
			log.Printf("HTTP proxy error: %v", err)
		}
	}()

	go func() {
		socks5srv, err := NewSOCKS5Server(s.auth)
		if err != nil {
			log.Printf("SOCKS5 setup error: %v", err)
			return
		}
		if err := socks5srv.Serve(socks5Ln); err != nil && ctx.Err() == nil {
			log.Printf("SOCKS5 proxy error: %v", err)
		}
	}()

	s.state.Store(int32(StateRunning))
	log.Printf("Proxy started: HTTP=%s SOCKS5=%s MITM=%v", httpAddr, socks5Addr, s.mitmEnabled)
	return nil
}

// Stop gracefully shuts down the proxy
func (s *ProxyServer) Stop() error {
	if !s.state.CompareAndSwap(int32(StateRunning), int32(StateStopping)) {
		return fmt.Errorf("proxy not running")
	}

	log.Println("Stopping proxy...")

	// Signal shutdown
	if s.cancelFunc != nil {
		s.cancelFunc()
	}

	// Close listeners (stop accepting new connections)
	if s.httpListener != nil {
		s.httpListener.Close()
	}
	if s.socks5Listener != nil {
		s.socks5Listener.Close()
	}

	// Wait for drain (5s max)
	deadline := time.After(5 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-deadline:
			log.Println("Drain timeout, forcing shutdown")
			goto done
		case <-ticker.C:
			if s.connCount.Load() == 0 {
				goto done
			}
		}
	}

done:
	s.transport.CloseAll()
	s.state.Store(int32(StateStopped))
	log.Println("Proxy stopped")
	return nil
}

// IsRunning returns true if proxy is running
func (s *ProxyServer) IsRunning() bool {
	return ServerState(s.state.Load()) == StateRunning
}

// GetConnectionCount returns active connection count
func (s *ProxyServer) GetConnectionCount() int {
	return int(s.connCount.Load())
}

// GetUptime returns seconds since start
func (s *ProxyServer) GetUptime() int64 {
	if !s.IsRunning() {
		return 0
	}
	return int64(time.Since(s.startTime).Seconds())
}

func (s *ProxyServer) setupHTTPProxy() error {
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = false

	// MITM setup
	if s.mitmEnabled && s.caCert != nil {
		goproxy.GoproxyCa = *s.caCert
		proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	}

	// Request handler - pipeline integration
	proxy.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		s.connCount.Add(1)

		// IP whitelist check
		if !s.auth.CheckIP(req.RemoteAddr) {
			s.connCount.Add(-1)
			return req, goproxy.NewResponse(req, goproxy.ContentTypeText, http.StatusForbidden, "Forbidden")
		}

		// Proxy-Authorization check
		if s.auth.authEnabled {
			authHeader := req.Header.Get("Proxy-Authorization")
			username, password, ok := parseProxyAuth(authHeader)
			if !ok || !s.auth.CheckCredentials(username, password) {
				s.connCount.Add(-1)
				resp := goproxy.NewResponse(req, goproxy.ContentTypeText, http.StatusProxyAuthRequired, "Proxy Authentication Required")
				resp.Header.Set("Proxy-Authenticate", "Basic realm=\"Proxy\"")
				return req, resp
			}
			req.Header.Del("Proxy-Authorization")
		}

		rctx := &RequestCtx{
			Request:   req,
			StartTime: time.Now(),
			Domain:    extractDomain(req.Host),
		}

		// Classify
		rctx.Route = s.pipeline.Classifier(req)

		// Block check
		if rctx.Route == RouteBlock {
			s.pipeline.Meter(rctx)
			s.connCount.Add(-1)
			return req, goproxy.NewResponse(req, goproxy.ContentTypeText, http.StatusNoContent, "")
		}

		// Bypass: PAC file should handle this. If request still arrives,
		// treat as direct (server's own IP) as fallback.
		if rctx.Route == RouteBypass {
			rctx.Route = RouteDirect
		}

		// Cache check
		if cachedResp := s.pipeline.CacheCheck(rctx); cachedResp != nil {
			rctx.Cached = true
			s.pipeline.Meter(rctx)
			s.connCount.Add(-1)
			return req, cachedResp
		}

		// Set transport based on route
		transport := s.transport.GetTransport(rctx.Route)
		ctx.RoundTripper = goproxy.RoundTripperFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Response, error) {
			return transport.RoundTrip(req)
		})

		// Store context for response handler
		ctx.UserData = rctx
		return req, nil
	})

	// Response handler
	proxy.OnResponse().DoFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		defer s.connCount.Add(-1)

		rctx, ok := ctx.UserData.(*RequestCtx)
		if !ok || rctx == nil {
			return resp
		}

		if resp != nil {
			// Estimate request size (method + URL + headers)
			rctx.ReqBytes = estimateRequestSize(rctx.Request)

			// Wrap response body to count bytes
			if resp.Body != nil {
				counter := &countingReadCloser{ReadCloser: resp.Body}
				resp.Body = counter
				// Store counter ref so meter can read final count after body consumed
				rctx.respCounter = counter
			}

			// Estimate response header size
			rctx.RespBytes = estimateHeaderSize(resp.Header) + int64(resp.ContentLength)

			// Store in cache
			s.pipeline.CacheStore(rctx, resp)
		}

		// Meter
		s.pipeline.Meter(rctx)
		return resp
	})

	s.httpProxy = proxy
	return nil
}

func extractDomain(host string) string {
	domain, _, err := net.SplitHostPort(host)
	if err != nil {
		return host
	}
	return domain
}

// countingReadCloser wraps an io.ReadCloser to count bytes read (thread-safe)
type countingReadCloser struct {
	io.ReadCloser
	n atomic.Int64
}

func (c *countingReadCloser) Read(p []byte) (int, error) {
	n, err := c.ReadCloser.Read(p)
	c.n.Add(int64(n))
	return n, err
}

func (c *countingReadCloser) BytesRead() int64 {
	return c.n.Load()
}

// estimateRequestSize estimates the byte size of a request (method + URL + headers)
func estimateRequestSize(req *http.Request) int64 {
	if req == nil {
		return 0
	}
	// Request line: "GET /path HTTP/1.1\r\n"
	size := int64(len(req.Method) + 1 + len(req.URL.RequestURI()) + 11)
	// Host header
	size += int64(6 + len(req.Host) + 2) // "Host: xxx\r\n"
	// Other headers
	size += estimateHeaderSize(req.Header)
	// Empty line
	size += 2
	// Content-Length for body
	if req.ContentLength > 0 {
		size += req.ContentLength
	}
	return size
}

// estimateHeaderSize sums up header byte sizes
func estimateHeaderSize(h http.Header) int64 {
	var size int64
	for key, vals := range h {
		for _, v := range vals {
			size += int64(len(key) + 2 + len(v) + 2) // "Key: Value\r\n"
		}
	}
	return size
}
