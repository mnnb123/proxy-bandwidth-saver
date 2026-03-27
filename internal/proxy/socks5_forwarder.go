package proxy

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// Buffer pools to avoid per-connection heap allocations.
var (
	// greetBufPool serves the 258-byte buffers used for the SOCKS5
	// greeting and CONNECT request parsing.
	greetBufPool = sync.Pool{
		New: func() any { return make([]byte, 258) },
	}
	// authBufPool serves the 515-byte buffers used for username/password
	// sub-negotiation (RFC 1929).
	authBufPool = sync.Pool{
		New: func() any { return make([]byte, 515) },
	}
)

// socks5Listener accepts SOCKS5 connections and forwards them through an upstream proxy.
type socks5Listener struct {
	upstream UpstreamInfo
	auth     *ProxyAuth
	proxyID  int
	meter    MeterCallback
	classify ClassifyFunc
}

// Serve accepts connections on the listener and handles each in a goroutine.
func (s *socks5Listener) Serve(ln net.Listener) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			return // listener closed
		}
		go s.handleConn(conn)
	}
}

func (s *socks5Listener) handleConn(conn net.Conn) {
	defer conn.Close()

	// 1. SOCKS5 greeting
	buf := greetBufPool.Get().([]byte)
	defer greetBufPool.Put(buf)
	n, err := conn.Read(buf)
	if err != nil || n < 2 || buf[0] != 0x05 {
		return
	}

	needAuth := s.auth != nil && s.auth.AuthEnabled()
	if needAuth {
		conn.Write([]byte{0x05, 0x02})
		if err := s.doAuth(conn); err != nil {
			return
		}
	} else {
		conn.Write([]byte{0x05, 0x00})
	}

	// 2. Read CONNECT request
	n, err = conn.Read(buf)
	if err != nil || n < 7 || buf[0] != 0x05 || buf[1] != 0x01 {
		conn.Write([]byte{0x05, 0x07, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return
	}

	// Parse target address
	var targetHost string
	var addrEnd int
	switch buf[3] {
	case 0x01: // IPv4
		if n < 10 {
			return
		}
		targetHost = fmt.Sprintf("%d.%d.%d.%d", buf[4], buf[5], buf[6], buf[7])
		addrEnd = 8
	case 0x03: // Domain
		domainLen := int(buf[4])
		if n < 5+domainLen+2 {
			return
		}
		targetHost = string(buf[5 : 5+domainLen])
		addrEnd = 5 + domainLen
	case 0x04: // IPv6
		if n < 22 {
			return
		}
		ip := net.IP(buf[4:20])
		targetHost = ip.String()
		addrEnd = 20
	default:
		conn.Write([]byte{0x05, 0x08, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return
	}
	port := int(buf[addrEnd])<<8 | int(buf[addrEnd+1])
	target := fmt.Sprintf("%s:%d", targetHost, port)

	// Classify the request
	var route Route
	if s.classify != nil {
		req, _ := http.NewRequest("CONNECT", "https://"+target, nil)
		req.Host = target
		route = s.classify(req)
	}

	// Block: reject
	if route == RouteBlock {
		s.recordMeter(targetHost, 0, 0, "block")
		conn.Write([]byte{0x05, 0x02, 0x00, 0x01, 0, 0, 0, 0, 0, 0}) // connection not allowed
		return
	}

	// 3. Connect: direct or via upstream
	upConn, dialErr := s.dialUpstream(route, target)
	if dialErr != nil {
		conn.Write([]byte{0x05, 0x05, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return
	}

	// 5. Success response
	conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0})

	// 6. Relay with byte counting — close both conns from outside goroutine to prevent leak
	var upBytes, downBytes atomic.Int64
	done := make(chan struct{})
	go func() {
		n, _ := io.Copy(upConn, conn)
		upBytes.Store(n)
		close(done)
	}()
	n2, _ := io.Copy(conn, upConn)
	downBytes.Store(n2)
	upConn.Close()
	<-done

	// Record to meter
	s.recordMeter(targetHost, upBytes.Load(), downBytes.Load(), route)
}

// recordMeter records bandwidth data to the meter callback, defaulting empty routes to "residential".
func (s *socks5Listener) recordMeter(domain string, reqBytes, respBytes int64, route Route) {
	if s.meter == nil {
		return
	}
	routeStr := string(route)
	if routeStr == "" {
		routeStr = "residential"
	}
	s.meter(domain, reqBytes, respBytes, s.proxyID, routeStr)
}

// dialUpstream connects to the target directly or via upstream proxy.
func (s *socks5Listener) dialUpstream(route Route, target string) (net.Conn, error) {
	if route == RouteDirect {
		return net.DialTimeout("tcp", target, 10*time.Second)
	}

	conn, err := net.DialTimeout("tcp", s.upstream.Address, 10*time.Second)
	if err != nil {
		return nil, err
	}

	var handshakeErr error
	if s.upstream.Type == "socks5" {
		handshakeErr = socks5Handshake(conn, target, s.upstream.Username, s.upstream.Password)
	} else {
		handshakeErr = httpConnectHandshake(conn, target, s.upstream.Username, s.upstream.Password)
	}
	if handshakeErr != nil {
		conn.Close()
		return nil, handshakeErr
	}
	return conn, nil
}

func (s *socks5Listener) doAuth(conn net.Conn) error {
	buf := authBufPool.Get().([]byte)
	defer authBufPool.Put(buf)
	n, err := conn.Read(buf)
	if err != nil || n < 3 || buf[0] != 0x01 {
		conn.Write([]byte{0x01, 0x01})
		return fmt.Errorf("bad auth")
	}

	ulen := int(buf[1])
	if n < 2+ulen+1 {
		conn.Write([]byte{0x01, 0x01})
		return fmt.Errorf("bad auth")
	}
	username := string(buf[2 : 2+ulen])
	plen := int(buf[2+ulen])
	if n < 3+ulen+plen {
		conn.Write([]byte{0x01, 0x01})
		return fmt.Errorf("bad auth")
	}
	password := string(buf[3+ulen : 3+ulen+plen])

	if !s.auth.CheckCredentials(username, password) {
		conn.Write([]byte{0x01, 0x01})
		return fmt.Errorf("auth failed")
	}

	conn.Write([]byte{0x01, 0x00})
	return nil
}
