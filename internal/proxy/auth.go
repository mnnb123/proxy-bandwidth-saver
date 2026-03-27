package proxy

import (
	"encoding/base64"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
)

// ProxyAuth handles authentication and IP whitelisting for proxy connections.
type ProxyAuth struct {
	mu sync.RWMutex

	// Username/password auth
	authEnabled bool
	username    string
	password    string

	// IP whitelist
	whitelistEnabled bool
	allowedIPs       map[string]bool
	allowedNets      []*net.IPNet
}

// NewProxyAuth creates a new ProxyAuth instance.
func NewProxyAuth() *ProxyAuth {
	return &ProxyAuth{
		allowedIPs: make(map[string]bool),
	}
}

// Configure updates auth settings. Thread-safe.
func (a *ProxyAuth) Configure(authEnabled bool, username, password string, whitelistEnabled bool, whitelist string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.authEnabled = authEnabled
	a.username = username
	a.password = password
	a.whitelistEnabled = whitelistEnabled

	// Parse whitelist
	a.allowedIPs = make(map[string]bool)
	a.allowedNets = nil

	if whitelist != "" {
		// Support both newline-separated and comma-separated entries
		normalized := strings.NewReplacer("\r\n", "\n", "\r", "\n", ",", "\n").Replace(whitelist)
		for _, entry := range strings.Split(normalized, "\n") {
			entry = strings.TrimSpace(entry)
			if entry == "" {
				continue
			}
			if strings.Contains(entry, "/") {
				_, cidr, err := net.ParseCIDR(entry)
				if err != nil {
					log.Printf("ProxyAuth: invalid CIDR %q: %v", entry, err)
					continue
				}
				a.allowedNets = append(a.allowedNets, cidr)
			} else {
				ip := net.ParseIP(entry)
				if ip == nil {
					log.Printf("ProxyAuth: invalid IP %q", entry)
					continue
				}
				a.allowedIPs[ip.String()] = true
			}
		}
	}

	// Always allow loopback
	a.allowedIPs["127.0.0.1"] = true
	a.allowedIPs["::1"] = true
}

// IsEnabled returns true if any auth mechanism is active.
func (a *ProxyAuth) IsEnabled() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.authEnabled || a.whitelistEnabled
}

// AuthEnabled returns true if username/password auth is required.
func (a *ProxyAuth) AuthEnabled() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.authEnabled
}

// CheckIP returns true if the IP is allowed (or whitelist is disabled).
func (a *ProxyAuth) CheckIP(remoteAddr string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if !a.whitelistEnabled {
		return true
	}

	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}

	// Check exact match
	if a.allowedIPs[ip.String()] {
		return true
	}

	// Check CIDR ranges
	for _, cidr := range a.allowedNets {
		if cidr.Contains(ip) {
			return true
		}
	}

	return false
}

// CheckCredentials returns true if username/password match (or auth is disabled).
func (a *ProxyAuth) CheckCredentials(username, password string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if !a.authEnabled {
		return true
	}
	return username == a.username && password == a.password
}

// CheckHTTPAuth validates Proxy-Authorization header.
// Returns true if auth passes. If false, writes 407 response.
func (a *ProxyAuth) CheckHTTPAuth(w http.ResponseWriter, r *http.Request) bool {
	// Check IP whitelist first
	if !a.CheckIP(r.RemoteAddr) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return false
	}

	a.mu.RLock()
	authRequired := a.authEnabled
	a.mu.RUnlock()

	if !authRequired {
		return true
	}

	// Parse Proxy-Authorization header
	authHeader := r.Header.Get("Proxy-Authorization")
	if authHeader == "" {
		w.Header().Set("Proxy-Authenticate", "Basic realm=\"Proxy\"")
		http.Error(w, "Proxy Authentication Required", http.StatusProxyAuthRequired)
		return false
	}

	username, password, ok := parseProxyAuth(authHeader)
	if !ok || !a.CheckCredentials(username, password) {
		w.Header().Set("Proxy-Authenticate", "Basic realm=\"Proxy\"")
		http.Error(w, "Proxy Authentication Required", http.StatusProxyAuthRequired)
		return false
	}

	// Remove auth header before forwarding
	r.Header.Del("Proxy-Authorization")
	return true
}

// parseProxyAuth parses "Basic base64(user:pass)" header value.
func parseProxyAuth(header string) (username, password string, ok bool) {
	const prefix = "Basic "
	if !strings.HasPrefix(header, prefix) {
		return "", "", false
	}

	decoded, err := base64.StdEncoding.DecodeString(header[len(prefix):])
	if err != nil {
		return "", "", false
	}

	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return "", "", false
	}

	return parts[0], parts[1], true
}

// WrapHandler wraps an HTTP handler with auth checks (for PortMapper output ports).
func (a *ProxyAuth) WrapHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !a.CheckHTTPAuth(w, r) {
			return
		}
		next.ServeHTTP(w, r)
	})
}
