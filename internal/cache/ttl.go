package cache

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultTTLStatic = 12 * time.Hour
	DefaultTTLImage  = 24 * time.Hour
	DefaultTTLHTML   = 5 * time.Minute
	DefaultTTLAPI    = 0 // no cache
	MaxTTL           = 24 * time.Hour
)

// ParseTTL determines cache TTL from response headers
func ParseTTL(resp *http.Response) time.Duration {
	cc := resp.Header.Get("Cache-Control")

	// Check s-maxage first (CDN directive)
	if sma := extractDirective(cc, "s-maxage"); sma > 0 {
		return clampTTL(time.Duration(sma) * time.Second)
	}

	// Check max-age
	if ma := extractDirective(cc, "max-age"); ma > 0 {
		return clampTTL(time.Duration(ma) * time.Second)
	}

	// Check Expires header
	if exp := resp.Header.Get("Expires"); exp != "" {
		if t, err := http.ParseTime(exp); err == nil {
			ttl := time.Until(t)
			if ttl > 0 {
				return clampTTL(ttl)
			}
		}
	}

	// Default based on content-type
	ct := resp.Header.Get("Content-Type")
	return defaultTTLForContentType(ct)
}

func extractDirective(cc, directive string) int {
	parts := strings.Split(cc, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if strings.HasPrefix(p, directive+"=") {
			val := strings.TrimPrefix(p, directive+"=")
			if n, err := strconv.Atoi(val); err == nil {
				return n
			}
		}
	}
	return 0
}

func defaultTTLForContentType(ct string) time.Duration {
	ct = strings.ToLower(ct)

	switch {
	case strings.HasPrefix(ct, "image/"):
		return DefaultTTLImage
	case strings.Contains(ct, "css"), strings.Contains(ct, "javascript"):
		return DefaultTTLStatic
	case strings.Contains(ct, "font"):
		return DefaultTTLStatic
	case strings.Contains(ct, "html"):
		return DefaultTTLHTML
	case strings.Contains(ct, "json"), strings.Contains(ct, "xml"):
		return DefaultTTLAPI
	default:
		return DefaultTTLHTML
	}
}

func clampTTL(ttl time.Duration) time.Duration {
	if ttl > MaxTTL {
		return MaxTTL
	}
	if ttl < 0 {
		return 0
	}
	return ttl
}
