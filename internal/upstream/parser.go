package upstream

import (
	"fmt"
	"net/url"
	"strings"
)

// ParseProxyLine parses a proxy string in various formats.
// Supports: host:port, host:port:socks5, host:port:user:pass, host:port:user:pass:socks5
// Also URL formats: http://user:pass@host:port, socks5://user:pass@host:port
func ParseProxyLine(line string) (address, username, password, proxyType string, err error) {
	line = strings.TrimSpace(line)
	proxyType = "http"

	// Try URL format first (http://... or socks5://...)
	if u, e := url.Parse(line); e == nil && u.Host != "" && u.Scheme != "" {
		address = u.Host
		if u.User != nil {
			username = u.User.Username()
			password, _ = u.User.Password()
		}
		if u.Scheme == "socks5" {
			proxyType = "socks5"
		}
		return
	}

	// Try host:port[:user:pass[:type]]
	parts := splitColon(line)
	switch len(parts) {
	case 2: // host:port
		address = parts[0] + ":" + parts[1]
	case 3: // host:port:type
		address = parts[0] + ":" + parts[1]
		if parts[2] == "socks5" || parts[2] == "socks" {
			proxyType = "socks5"
		}
	case 4: // host:port:user:pass
		address = parts[0] + ":" + parts[1]
		username = parts[2]
		password = parts[3]
	case 5: // host:port:user:pass:type
		address = parts[0] + ":" + parts[1]
		username = parts[2]
		password = parts[3]
		if parts[4] == "socks5" || parts[4] == "socks" {
			proxyType = "socks5"
		}
	default:
		err = fmt.Errorf("invalid proxy format: %s", line)
	}
	return
}

func splitColon(s string) []string {
	var parts []string
	current := ""
	for _, c := range s {
		if c == ':' {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}
