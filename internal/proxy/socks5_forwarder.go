package proxy

import (
	"fmt"
	"io"
	"net"
	"time"
)

// socks5Listener accepts SOCKS5 connections and forwards them through an upstream proxy.
type socks5Listener struct {
	upstream UpstreamInfo
	auth     *ProxyAuth
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
	buf := make([]byte, 258)
	n, err := conn.Read(buf)
	if err != nil || n < 2 || buf[0] != 0x05 {
		return
	}

	needAuth := s.auth != nil && s.auth.AuthEnabled()
	if needAuth {
		// Request username/password auth
		conn.Write([]byte{0x05, 0x02})
		if err := s.doAuth(conn); err != nil {
			return
		}
	} else {
		// No auth
		conn.Write([]byte{0x05, 0x00})
	}

	// 2. Read CONNECT request
	n, err = conn.Read(buf)
	if err != nil || n < 7 || buf[0] != 0x05 || buf[1] != 0x01 {
		conn.Write([]byte{0x05, 0x07, 0x00, 0x01, 0, 0, 0, 0, 0, 0}) // command not supported
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

	// 3. Connect to upstream proxy
	upConn, err := net.DialTimeout("tcp", s.upstream.Address, 10*time.Second)
	if err != nil {
		conn.Write([]byte{0x05, 0x05, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return
	}

	// 4. Handshake with upstream
	var hsErr error
	if s.upstream.Type == "socks5" {
		hsErr = socks5Handshake(upConn, target, s.upstream.Username, s.upstream.Password)
	} else {
		hsErr = httpConnectHandshake(upConn, target, s.upstream.Username, s.upstream.Password)
	}
	if hsErr != nil {
		upConn.Close()
		conn.Write([]byte{0x05, 0x05, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return
	}

	// 5. Success response
	conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0})

	// 6. Relay
	go func() { io.Copy(upConn, conn); upConn.Close() }()
	io.Copy(conn, upConn)
}

func (s *socks5Listener) doAuth(conn net.Conn) error {
	buf := make([]byte, 515)
	n, err := conn.Read(buf)
	if err != nil || n < 3 || buf[0] != 0x01 {
		conn.Write([]byte{0x01, 0x01}) // auth failed
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

	conn.Write([]byte{0x01, 0x00}) // success
	return nil
}

