package proxy

import (
	"net/http"
	"net/url"
	"sync"
	"time"
)

// TransportManager manages connection pools for different routes
type TransportManager struct {
	direct      *http.Transport
	datacenter  *http.Transport
	residential *http.Transport

	mu            sync.RWMutex
	dcProxyURL    *url.URL
	resProxyURL   *url.URL
}

func NewTransportManager() *TransportManager {
	tm := &TransportManager{
		direct: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
			MaxConnsPerHost:     20,
		},
	}
	tm.datacenter = tm.newTransport(nil)
	tm.residential = tm.newTransport(nil)
	return tm
}

func (tm *TransportManager) newTransport(proxyURL *url.URL) *http.Transport {
	t := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		MaxConnsPerHost:     10,
	}
	if proxyURL != nil {
		t.Proxy = http.ProxyURL(proxyURL)
	}
	return t
}

// SetDatacenterProxy sets the datacenter proxy URL
func (tm *TransportManager) SetDatacenterProxy(proxyURL *url.URL) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.dcProxyURL = proxyURL
	tm.datacenter.CloseIdleConnections()
	tm.datacenter = tm.newTransport(proxyURL)
}

// SetResidentialProxy sets the residential proxy URL
func (tm *TransportManager) SetResidentialProxy(proxyURL *url.URL) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.resProxyURL = proxyURL
	tm.residential.CloseIdleConnections()
	tm.residential = tm.newTransport(proxyURL)
}

// GetTransport returns the appropriate transport for the given route
func (tm *TransportManager) GetTransport(route Route) *http.Transport {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	switch route {
	case RouteDatacenter:
		return tm.datacenter
	case RouteResidential:
		return tm.residential
	default:
		return tm.direct
	}
}

// CloseAll closes all idle connections
func (tm *TransportManager) CloseAll() {
	tm.direct.CloseIdleConnections()
	tm.mu.RLock()
	tm.datacenter.CloseIdleConnections()
	tm.residential.CloseIdleConnections()
	tm.mu.RUnlock()
}
