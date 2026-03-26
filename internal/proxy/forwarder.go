package proxy

import (
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"
)

// ClassifyFunc classifies a domain and returns the route decision.
type ClassifyFunc func(req *http.Request) Route

// proxyForwarder is an HTTP handler that acts as a forward proxy,
// routing all traffic through a specific upstream proxy.
type proxyForwarder struct {
	upstream  UpstreamInfo
	proxyID   int
	meter     MeterCallback
	classify  ClassifyFunc
	transport *http.Transport
	directTransport *http.Transport
}

func newProxyForwarder(upstream UpstreamInfo, proxyID int, meter MeterCallback, classify ClassifyFunc) *proxyForwarder {
	proxyURL := &url.URL{
		Scheme: "http",
		Host:   upstream.Address,
	}
	if upstream.Username != "" {
		proxyURL.User = url.UserPassword(upstream.Username, upstream.Password)
	}
	return &proxyForwarder{
		upstream: upstream,
		proxyID:  proxyID,
		meter:    meter,
		classify: classify,
		transport: &http.Transport{
			Proxy:               http.ProxyURL(proxyURL),
			MaxIdleConns:        50,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
		},
		directTransport: &http.Transport{
			MaxIdleConns:        50,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}
}

func (f *proxyForwarder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		f.handleConnect(w, r)
	} else {
		f.handleHTTP(w, r)
	}
}

// checkRoute classifies and returns the route. Returns "" to proceed normally.
func (f *proxyForwarder) checkRoute(r *http.Request) Route {
	if f.classify == nil {
		return ""
	}
	return f.classify(r)
}

// handleHTTP forwards regular HTTP requests via upstream proxy.
func (f *proxyForwarder) handleHTTP(w http.ResponseWriter, r *http.Request) {
	domain := r.Host
	reqBytes := r.ContentLength
	if reqBytes < 0 {
		reqBytes = 0
	}

	route := f.checkRoute(r)

	// Block: return 204 No Content
	if route == RouteBlock {
		f.recordMeter(domain, 0, 0, "block")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Bypass: PAC file should handle this domain (client connects directly).
	// Return 502 so client knows this proxy won't serve it.
	if route == RouteBypass {
		f.recordMeter(domain, 0, 0, "bypass")
		http.Error(w, "bypass: use PAC file for direct connection", http.StatusBadGateway)
		return
	}

	r.RequestURI = ""

	// Bypass/Direct: use direct transport (no upstream proxy)
	transport := f.transport
	if route == RouteDirect {
		transport = f.directTransport
	}

	resp, err := transport.RoundTrip(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)

	respBytes, _ := io.Copy(w, resp.Body)

	f.recordMeter(domain, reqBytes, respBytes, route)
}

// handleConnect handles HTTPS tunneling (CONNECT method) via upstream proxy.
func (f *proxyForwarder) handleConnect(w http.ResponseWriter, r *http.Request) {
	domain := r.Host

	route := f.checkRoute(r)

	// Block: reject connection
	if route == RouteBlock {
		f.recordMeter(domain, 0, 0, "block")
		http.Error(w, "blocked", http.StatusForbidden)
		return
	}

	// Bypass: PAC file should handle this (client connects directly)
	if route == RouteBypass {
		f.recordMeter(domain, 0, 0, "bypass")
		http.Error(w, "bypass: use PAC file for direct connection", http.StatusBadGateway)
		return
	}

	// Connect to target: direct or via upstream proxy
	upConn, err := f.dialUpstream(route, r.Host)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		upConn.Close()
		http.Error(w, "hijack not supported", http.StatusInternalServerError)
		return
	}
	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		upConn.Close()
		return
	}

	clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))

	// Relay with byte counting — close both conns from outside goroutine to prevent leak
	var upBytes, downBytes atomic.Int64
	done := make(chan struct{})
	go func() {
		n, _ := io.Copy(upConn, clientConn)
		upBytes.Store(n)
		close(done)
	}()
	n, _ := io.Copy(clientConn, upConn)
	downBytes.Store(n)
	upConn.Close()
	clientConn.Close()
	<-done

	f.recordMeter(domain, upBytes.Load(), downBytes.Load(), route)
}

// dialUpstream connects to the target directly or via upstream proxy with handshake.
func (f *proxyForwarder) dialUpstream(route Route, targetHost string) (net.Conn, error) {
	if route == RouteDirect {
		return net.DialTimeout("tcp", targetHost, 10*time.Second)
	}

	conn, err := net.DialTimeout("tcp", f.upstream.Address, 10*time.Second)
	if err != nil {
		return nil, err
	}

	var handshakeErr error
	if f.upstream.Type == "socks5" {
		handshakeErr = socks5Handshake(conn, targetHost, f.upstream.Username, f.upstream.Password)
	} else {
		handshakeErr = httpConnectHandshake(conn, targetHost, f.upstream.Username, f.upstream.Password)
	}
	if handshakeErr != nil {
		conn.Close()
		return nil, handshakeErr
	}
	return conn, nil
}

// recordMeter records bandwidth data to the meter callback, defaulting empty routes to "residential".
func (f *proxyForwarder) recordMeter(domain string, reqBytes, respBytes int64, route Route) {
	if f.meter == nil {
		return
	}
	routeStr := string(route)
	if routeStr == "" {
		routeStr = "residential"
	}
	f.meter(domain, reqBytes, respBytes, f.proxyID, routeStr)
}

// httpConnectHandshake sends HTTP CONNECT to upstream proxy.
func httpConnectHandshake(conn net.Conn, targetHost, username, password string) error {
	connectReq := fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\n", targetHost, targetHost)
	if username != "" {
		creds := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
		connectReq += fmt.Sprintf("Proxy-Authorization: Basic %s\r\n", creds)
	}
	connectReq += "\r\n"

	if _, err := conn.Write([]byte(connectReq)); err != nil {
		return err
	}

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil || n < 12 || string(buf[9:12]) != "200" {
		return fmt.Errorf("upstream CONNECT rejected")
	}
	return nil
}

// socks5Handshake performs SOCKS5 handshake to connect to target host.
func socks5Handshake(conn net.Conn, targetHost, username, password string) error {
	host, port, err := net.SplitHostPort(targetHost)
	if err != nil {
		host = targetHost
		port = "443"
	}
	portNum := 0
	fmt.Sscanf(port, "%d", &portNum)

	// Auth method selection
	if username != "" {
		conn.Write([]byte{0x05, 0x01, 0x02})
	} else {
		conn.Write([]byte{0x05, 0x01, 0x00})
	}

	buf := make([]byte, 2)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return fmt.Errorf("auth response: %w", err)
	}
	if buf[0] != 0x05 {
		return fmt.Errorf("not socks5")
	}

	// Username/password auth sub-negotiation
	if buf[1] == 0x02 {
		user := []byte(username)
		pass := []byte(password)
		authReq := []byte{0x01, byte(len(user))}
		authReq = append(authReq, user...)
		authReq = append(authReq, byte(len(pass)))
		authReq = append(authReq, pass...)
		conn.Write(authReq)

		authResp := make([]byte, 2)
		if _, err := io.ReadFull(conn, authResp); err != nil {
			return fmt.Errorf("auth: %w", err)
		}
		if authResp[1] != 0x00 {
			return fmt.Errorf("auth failed")
		}
	} else if buf[1] == 0xFF {
		return fmt.Errorf("no acceptable auth method")
	}

	// CONNECT request
	req := []byte{0x05, 0x01, 0x00, 0x03, byte(len(host))}
	req = append(req, []byte(host)...)
	req = append(req, byte(portNum>>8), byte(portNum&0xFF))
	conn.Write(req)

	// Read response
	resp := make([]byte, 4)
	if _, err := io.ReadFull(conn, resp); err != nil {
		return fmt.Errorf("connect response: %w", err)
	}
	if resp[1] != 0x00 {
		return fmt.Errorf("connect failed: code %d", resp[1])
	}

	// Skip remaining address bytes
	switch resp[3] {
	case 0x01: // IPv4
		io.ReadFull(conn, make([]byte, 4+2))
	case 0x03: // Domain
		lenBuf := make([]byte, 1)
		io.ReadFull(conn, lenBuf)
		io.ReadFull(conn, make([]byte, int(lenBuf[0])+2))
	case 0x04: // IPv6
		io.ReadFull(conn, make([]byte, 16+2))
	}

	return nil
}
