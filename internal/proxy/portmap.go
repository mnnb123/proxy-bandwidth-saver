package proxy

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"sort"
	"sync"
	"time"
)

// GetPublicAddr returns the bind address for display purposes.
// If bind is "0.0.0.0", it resolves to the machine's first non-loopback IP.
func GetPublicAddr(bind string) string {
	if bind != "0.0.0.0" && bind != "" {
		return bind
	}
	ifaces, err := net.Interfaces()
	if err != nil {
		return bind
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return bind
}

// MeterCallback is called for each proxied request with domain, bytes, and route info.
type MeterCallback func(domain string, reqBytes, respBytes int64, proxyID int, route string)

// PortMapper creates individual local proxy listeners (HTTP + SOCKS5),
// each forwarding all traffic through a specific upstream proxy.
type PortMapper struct {
	mu       sync.Mutex
	entries  map[int]*portMapEntry
	basePort int
	bindAddr string
	auth     *ProxyAuth
	meter    MeterCallback
	classify ClassifyFunc
	ctx      context.Context
	cancel   context.CancelFunc
}

type portMapEntry struct {
	proxyID  int
	port     int
	upstream UpstreamInfo

	listener net.Listener
	cancel   context.CancelFunc
}

// UpstreamInfo holds upstream proxy connection details.
type UpstreamInfo struct {
	Address  string
	Username string
	Password string
	Type     string // "http" or "socks5"
}

// MappedProxy represents a single local->upstream port mapping.
// Each port supports both HTTP and SOCKS5 via protocol auto-detection.
type MappedProxy struct {
	ProxyID    int    `json:"proxyId"`
	LocalAddr  string `json:"localAddr"`
	LocalPort  int    `json:"localPort"`
	Protocol   string `json:"protocol"` // "http+socks5"
	Upstream   string `json:"upstream"`
	Type       string `json:"type"`
}

func NewPortMapper(bindAddr string, basePort int, auth *ProxyAuth, meter MeterCallback, classify ClassifyFunc) *PortMapper {
	ctx, cancel := context.WithCancel(context.Background())
	return &PortMapper{
		entries:  make(map[int]*portMapEntry),
		basePort: basePort,
		bindAddr: bindAddr,
		auth:     auth,
		meter:    meter,
		classify: classify,
		ctx:      ctx,
		cancel:   cancel,
	}
}

// MapProxy creates a single local proxy listener for a given upstream proxy.
// The listener auto-detects HTTP vs SOCKS5 by peeking the first byte.
func (pm *PortMapper) MapProxy(proxyID int, up UpstreamInfo) (string, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if e, ok := pm.entries[proxyID]; ok {
		return fmt.Sprintf("%s:%d", pm.bindAddr, e.port), nil
	}

	port := pm.findAvailablePort()
	addr := fmt.Sprintf("%s:%d", pm.bindAddr, port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return "", fmt.Errorf("listen %s: %w", addr, err)
	}

	// Build HTTP handler
	forwarder := newProxyForwarder(up, proxyID, pm.meter, pm.classify)
	var handler http.Handler = forwarder
	if pm.auth != nil {
		handler = pm.auth.WrapHandler(forwarder)
	}
	httpSrv := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// Build SOCKS5 handler
	s5 := &socks5Listener{upstream: up, auth: pm.auth, proxyID: proxyID, meter: pm.meter, classify: pm.classify}

	ctx, cancel := context.WithCancel(pm.ctx)

	// Multiplexing listener: peek first byte to detect protocol
	httpLn := &muxListener{ch: make(chan net.Conn, 64), done: ctx.Done()}
	go func() {
		if err := httpSrv.Serve(httpLn); err != nil && err != http.ErrServerClosed {
			log.Printf("PortMapper: http serve error on %s: %v", addr, err)
		}
	}()

	go func() {
		defer ln.Close()
		for {
			conn, err := ln.Accept()
			if err != nil {
				select {
				case <-ctx.Done():
					return
				default:
					return
				}
			}
			go func(c net.Conn) {
				br := bufio.NewReaderSize(c, 1)
				first, err := br.Peek(1)
				if err != nil {
					c.Close()
					return
				}
				peeked := &peekedConn{Conn: c, reader: br}
				if first[0] == 0x05 {
					// SOCKS5
					s5.handleConn(peeked)
				} else {
					// HTTP — push to muxListener for http.Server
					select {
					case httpLn.ch <- peeked:
					case <-ctx.Done():
						c.Close()
					}
				}
			}(conn)
		}
	}()

	go func() {
		<-ctx.Done()
		httpSrv.Close()
		ln.Close()
	}()

	entry := &portMapEntry{
		proxyID:  proxyID,
		port:     port,
		upstream: up,
		listener: ln,
		cancel:   cancel,
	}
	pm.entries[proxyID] = entry

	log.Printf("PortMapper: %s (HTTP+SOCKS5) -> %s (%s)", addr, up.Address, up.Type)
	return addr, nil
}

// peekedConn wraps a net.Conn with a bufio.Reader that has peeked bytes.
type peekedConn struct {
	net.Conn
	reader *bufio.Reader
}

func (c *peekedConn) Read(b []byte) (int, error) {
	return c.reader.Read(b)
}

// muxListener is a net.Listener backed by a channel of connections.
// It allows http.Server.Serve() to receive connections dispatched by
// the multiplexing accept loop.
type muxListener struct {
	ch   chan net.Conn
	done <-chan struct{}
}

func (l *muxListener) Accept() (net.Conn, error) {
	select {
	case c, ok := <-l.ch:
		if !ok {
			return nil, net.ErrClosed
		}
		return c, nil
	case <-l.done:
		return nil, net.ErrClosed
	}
}

func (l *muxListener) Close() error {
	return nil // lifecycle managed by parent
}

func (l *muxListener) Addr() net.Addr {
	return &net.TCPAddr{}
}

// UnmapProxy stops and removes the listeners for a given proxy ID.
func (pm *PortMapper) UnmapProxy(proxyID int) {
	pm.mu.Lock()
	e, ok := pm.entries[proxyID]
	if ok {
		delete(pm.entries, proxyID)
	}
	pm.mu.Unlock()

	if ok {
		e.cancel()
		log.Printf("PortMapper: unmapped proxy %d (port:%d)", proxyID, e.port)
	}
}

// GetMappings returns all current port mappings sorted by local port.
// Each upstream proxy produces one entry supporting both HTTP and SOCKS5.
func (pm *PortMapper) GetMappings() []MappedProxy {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	displayAddr := GetPublicAddr(pm.bindAddr)

	result := make([]MappedProxy, 0, len(pm.entries))
	for _, e := range pm.entries {
		result = append(result, MappedProxy{
			ProxyID:   e.proxyID,
			LocalAddr: fmt.Sprintf("%s:%d", displayAddr, e.port),
			LocalPort: e.port,
			Protocol:  "http+socks5",
			Upstream:  e.upstream.Address,
			Type:      e.upstream.Type,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].LocalPort < result[j].LocalPort
	})
	return result
}

// StopAll shuts down all listeners.
func (pm *PortMapper) StopAll() {
	pm.cancel()
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for id, e := range pm.entries {
		e.cancel()
		delete(pm.entries, id)
	}
	log.Println("PortMapper: all stopped")
}

// findAvailablePort returns the next available port starting from basePort.
func (pm *PortMapper) findAvailablePort() int {
	used := make(map[int]bool)
	for _, e := range pm.entries {
		used[e.port] = true
	}
	for p := pm.basePort; ; p++ {
		if !used[p] {
			return p
		}
	}
}
