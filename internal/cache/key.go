package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
)

// GenerateKey creates a cache key from the request
// Key = SHA256(METHOD + URL + Host)
func GenerateKey(req *http.Request) string {
	var b strings.Builder
	b.WriteString(req.Method)
	b.WriteString("|")
	b.WriteString(req.URL.String())
	b.WriteString("|")
	b.WriteString(req.Host)

	hash := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(hash[:])
}

// ShouldBypass returns true if the request should skip cache
func ShouldBypass(req *http.Request) bool {
	// Never cache non-GET/HEAD
	if req.Method != http.MethodGet && req.Method != http.MethodHead {
		return true
	}

	// Skip if Authorization header present
	if req.Header.Get("Authorization") != "" {
		return true
	}

	// Skip if Cache-Control: no-store
	cc := req.Header.Get("Cache-Control")
	if strings.Contains(cc, "no-store") {
		return true
	}

	return false
}

// ShouldStoreResponse returns true if the response should be cached
func ShouldStoreResponse(resp *http.Response) bool {
	if resp == nil || resp.StatusCode >= 400 {
		return false
	}

	cc := resp.Header.Get("Cache-Control")
	if strings.Contains(cc, "no-store") || strings.Contains(cc, "private") {
		return false
	}

	return true
}
