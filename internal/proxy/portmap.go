package proxy

import (
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

// MeterCallback is called for each proxied request with domain and bytes transferred.
type MeterCallback func(domain string, reqBytes, respBytes int64, proxyID int)

// PortMapper creates individual local proxy listeners (HTTP + SOCKS5),
// each forwarding all traffic through a specific upstream proxy.
type PortMapper struct {
	mu       sync.Mutex
	entries  map[int]*portMapEntry
	basePort int
	bindAddr string
	auth     *ProxyAuth
	meter    MeterCallback
	ctx      context.Context
	cancel   context.CancelFunc
}

type portMapEntry struct {
	proxyID   int
	httpPort  int
	socks5Port int
	upstream  UpstreamInfo

	httpServer   *http.Server
	httpListener net.Listener

	socks5Listener net.Listener
	socks5Cancel   context.CancelFunc
}

// UpstreamInfo holds upstream proxy connection details.
type UpstreamInfo struct {
	Address  string
	Username string
	Password string
	Type     string // "http" or "socks5"
}

// MappedProxy represents a single local->upstream port mapping.
type MappedProxy struct {
	ProxyID    int    `json:"proxyId"`
	LocalAddr  string `json:"localAddr"`
	LocalPort  int    `json:"localPort"`
	Protocol   string `json:"protocol"` // "http" or "socks5"
	Upstream   string `json:"upstream"`
	Type       string `json:"type"`
}

func NewPortMapper(bindAddr string, basePort int, auth *ProxyAuth, meter MeterCallback) *PortMapper {
	ctx, cancel := context.WithCancel(context.Background())
	return &PortMapper{
		entries:  make(map[int]*portMapEntry),
		basePort: basePort,
		bindAddr: bindAddr,
		auth:     auth,
		meter:    meter,
		ctx:      ctx,
		cancel:   cancel,
	}
}

// MapProxy creates HTTP + SOCKS5 local proxy listeners for a given upstream proxy.
// Uses two consecutive ports: even=HTTP, odd=SOCKS5.
func (pm *PortMapper) MapProxy(proxyID int, up UpstreamInfo) (string, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if e, ok := pm.entries[proxyID]; ok {
		return fmt.Sprintf("%s:%d", pm.bindAddr, e.httpPort), nil
	}

	httpPort, socks5Port := pm.findAvailablePortPair()

	// --- HTTP listener ---
	httpAddr := fmt.Sprintf("%s:%d", pm.bindAddr, httpPort)
	httpLn, err := net.Listen("tcp", httpAddr)
	if err != nil {
		return "", fmt.Errorf("listen http %s: %w", httpAddr, err)
	}
	forwarder := newProxyForwarder(up, proxyID, pm.meter)
	var handler http.Handler = forwarder
	if pm.auth != nil {
		handler = pm.auth.WrapHandler(forwarder)
	}
	httpSrv := &http.Server{Handler: handler}
	go func() {
		if err := httpSrv.Serve(httpLn); err != nil && err != http.ErrServerClosed {
			log.Printf("PortMapper: http serve error on %s: %v", httpAddr, err)
		}
	}()

	// --- SOCKS5 listener ---
	socks5Addr := fmt.Sprintf("%s:%d", pm.bindAddr, socks5Port)
	socks5Ln, err := net.Listen("tcp", socks5Addr)
	if err != nil {
		httpSrv.Close()
		httpLn.Close()
		return "", fmt.Errorf("listen socks5 %s: %w", socks5Addr, err)
	}
	s5 := &socks5Listener{upstream: up, auth: pm.auth, proxyID: proxyID, meter: pm.meter}
	s5ctx, s5cancel := context.WithCancel(pm.ctx)
	go func() {
		go s5.Serve(socks5Ln)
		<-s5ctx.Done()
		socks5Ln.Close()
	}()

	entry := &portMapEntry{
		proxyID:        proxyID,
		httpPort:       httpPort,
		socks5Port:     socks5Port,
		upstream:       up,
		httpServer:     httpSrv,
		httpListener:   httpLn,
		socks5Listener: socks5Ln,
		socks5Cancel:   s5cancel,
	}
	pm.entries[proxyID] = entry

	log.Printf("PortMapper: HTTP %s + SOCKS5 %s -> %s (%s)", httpAddr, socks5Addr, up.Address, up.Type)
	return httpAddr, nil
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
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		e.httpServer.Shutdown(ctx)
		e.socks5Cancel()
		log.Printf("PortMapper: unmapped proxy %d (http:%d socks5:%d)", proxyID, e.httpPort, e.socks5Port)
	}
}

// GetMappings returns all current port mappings sorted by local port.
// Each upstream proxy produces two entries: one HTTP, one SOCKS5.
func (pm *PortMapper) GetMappings() []MappedProxy {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	displayAddr := GetPublicAddr(pm.bindAddr)

	result := make([]MappedProxy, 0, len(pm.entries)*2)
	for _, e := range pm.entries {
		result = append(result, MappedProxy{
			ProxyID:   e.proxyID,
			LocalAddr: fmt.Sprintf("%s:%d", displayAddr, e.httpPort),
			LocalPort: e.httpPort,
			Protocol:  "http",
			Upstream:  e.upstream.Address,
			Type:      e.upstream.Type,
		})
		result = append(result, MappedProxy{
			ProxyID:   e.proxyID,
			LocalAddr: fmt.Sprintf("%s:%d", displayAddr, e.socks5Port),
			LocalPort: e.socks5Port,
			Protocol:  "socks5",
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
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		e.httpServer.Shutdown(ctx)
		e.socks5Cancel()
		cancel()
		delete(pm.entries, id)
	}
	log.Println("PortMapper: all stopped")
}

// findAvailablePortPair returns two consecutive available ports.
func (pm *PortMapper) findAvailablePortPair() (httpPort, socks5Port int) {
	used := make(map[int]bool)
	for _, e := range pm.entries {
		used[e.httpPort] = true
		used[e.socks5Port] = true
	}
	for p := pm.basePort; ; p += 2 {
		if !used[p] && !used[p+1] {
			return p, p + 1
		}
	}
}
