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

// PortMapper creates individual local HTTP proxy listeners, each forwarding
// all traffic through a specific upstream proxy.
type PortMapper struct {
	mu       sync.Mutex
	entries  map[int]*portMapEntry
	basePort int
	bindAddr string
	auth     *ProxyAuth
	ctx      context.Context
	cancel   context.CancelFunc
}

type portMapEntry struct {
	proxyID   int
	localPort int
	localAddr string
	upstream  UpstreamInfo
	server    *http.Server
	listener  net.Listener
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
	ProxyID   int    `json:"proxyId"`
	LocalAddr string `json:"localAddr"`
	LocalPort int    `json:"localPort"`
	Upstream  string `json:"upstream"`
	Type      string `json:"type"`
}

func NewPortMapper(bindAddr string, basePort int, auth *ProxyAuth) *PortMapper {
	ctx, cancel := context.WithCancel(context.Background())
	return &PortMapper{
		entries:  make(map[int]*portMapEntry),
		basePort: basePort,
		bindAddr: bindAddr,
		auth:     auth,
		ctx:      ctx,
		cancel:   cancel,
	}
}

// MapProxy creates a local HTTP proxy listener for a given upstream proxy.
func (pm *PortMapper) MapProxy(proxyID int, up UpstreamInfo) (string, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if e, ok := pm.entries[proxyID]; ok {
		return e.localAddr, nil
	}

	port := pm.findAvailablePort()
	addr := fmt.Sprintf("%s:%d", pm.bindAddr, port)

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return "", fmt.Errorf("listen %s: %w", addr, err)
	}

	forwarder := newProxyForwarder(up)
	var handler http.Handler = forwarder
	if pm.auth != nil {
		handler = pm.auth.WrapHandler(forwarder)
	}
	srv := &http.Server{Handler: handler}

	entry := &portMapEntry{
		proxyID:   proxyID,
		localPort: port,
		localAddr: addr,
		upstream:  up,
		server:    srv,
		listener:  ln,
	}
	pm.entries[proxyID] = entry

	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Printf("PortMapper: serve error on %s: %v", addr, err)
		}
	}()

	log.Printf("PortMapper: %s -> %s (%s)", addr, up.Address, up.Type)
	return addr, nil
}

// UnmapProxy stops and removes the listener for a given proxy ID.
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
		e.server.Shutdown(ctx)
		log.Printf("PortMapper: unmapped proxy %d (%s)", proxyID, e.localAddr)
	}
}

// GetMappings returns all current port mappings sorted by local port.
// Display addresses replace "0.0.0.0" with the machine's real IP.
func (pm *PortMapper) GetMappings() []MappedProxy {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	displayAddr := GetPublicAddr(pm.bindAddr)

	result := make([]MappedProxy, 0, len(pm.entries))
	for _, e := range pm.entries {
		addr := fmt.Sprintf("%s:%d", displayAddr, e.localPort)
		result = append(result, MappedProxy{
			ProxyID:   e.proxyID,
			LocalAddr: addr,
			LocalPort: e.localPort,
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
		e.server.Shutdown(ctx)
		cancel()
		delete(pm.entries, id)
	}
	log.Println("PortMapper: all stopped")
}

func (pm *PortMapper) findAvailablePort() int {
	used := make(map[int]bool)
	for _, e := range pm.entries {
		used[e.localPort] = true
	}
	for p := pm.basePort; ; p++ {
		if !used[p] {
			return p
		}
	}
}
