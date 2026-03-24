package cache

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"path/filepath"

	"proxy-bandwidth-saver/internal/proxy"

	"golang.org/x/sync/singleflight"
)

const (
	defaultMemoryThreshold = 1024 * 1024 // 1MB: smaller -> memory, larger -> disk
)

// CacheLayer provides two-tier caching with singleflight dedup
type CacheLayer struct {
	memory    *MemoryCache
	disk      *DiskCache
	group     singleflight.Group
	stats     *Stats
	threshold int64 // body size threshold for memory vs disk
}

// NewCacheLayer initializes the two-tier cache
func NewCacheLayer(cacheDir string, memoryMB, diskMB int) (*CacheLayer, error) {
	mem, err := NewMemoryCache(memoryMB)
	if err != nil {
		return nil, err
	}

	diskPath := filepath.Join(cacheDir, "responses.db")
	disk, err := NewDiskCache(diskPath, diskMB)
	if err != nil {
		mem.Close()
		return nil, err
	}

	return &CacheLayer{
		memory:    mem,
		disk:      disk,
		stats:     &Stats{},
		threshold: defaultMemoryThreshold,
	}, nil
}

// CheckCache is the pipeline hook for cache lookups
func (cl *CacheLayer) CheckCache(ctx *proxy.RequestCtx) *http.Response {
	req := ctx.Request
	if ShouldBypass(req) {
		return nil
	}

	key := GenerateKey(req)
	ctx.CacheKey = key

	// Check memory cache
	if entry, ok := cl.memory.Get(key); ok {
		cl.stats.RecordHit(int64(len(entry.Body)))
		resp := entry.ToResponse(req)
		resp.Header.Set("X-Cache", "HIT-MEM")
		return resp
	}

	// Check disk cache
	if entry, ok := cl.disk.Get(key); ok {
		cl.stats.RecordHit(int64(len(entry.Body)))
		// Promote to memory if small enough
		if int64(len(entry.Body)) < cl.threshold {
			cl.memory.Set(key, entry)
		}
		resp := entry.ToResponse(req)
		resp.Header.Set("X-Cache", "HIT-DISK")
		return resp
	}

	cl.stats.RecordMiss()
	return nil
}

// StoreCache is the pipeline hook for storing responses
func (cl *CacheLayer) StoreCache(ctx *proxy.RequestCtx, resp *http.Response) {
	if resp == nil || ctx.CacheKey == "" {
		return
	}
	if !ShouldStoreResponse(resp) {
		return
	}

	// Read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	// Replace body so it can be read again
	resp.Body = io_NopCloser2(body)

	ttl := ParseTTL(resp)
	if ttl <= 0 {
		return
	}

	entry := &CacheEntry{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header.Clone(),
		Body:       body,
		TTL:        ttl,
		StoredAt:   ctx.StartTime,
		Size:       int64(len(body)),
	}

	// Store in appropriate tier
	if int64(len(body)) < cl.threshold {
		cl.memory.Set(ctx.CacheKey, entry)
	} else {
		cl.disk.Set(ctx.CacheKey, entry)
	}

	cl.stats.Entries.Add(1)
}

// GetStats returns cache statistics
func (cl *CacheLayer) GetStats() (hits, misses, bytesSaved int64, hitRatio float64, memUsedMB, diskUsedMB float64) {
	hits = cl.stats.Hits.Load()
	misses = cl.stats.Misses.Load()
	bytesSaved = cl.stats.BytesSaved.Load()
	hitRatio = cl.stats.HitRatio()
	diskUsedMB = float64(cl.disk.UsedSize()) / (1024 * 1024)
	return
}

// Clear empties both cache tiers
func (cl *CacheLayer) Clear() {
	cl.memory.Clear()
	cl.disk.Clear()
	cl.stats.Entries.Store(0)
	log.Println("Cache cleared")
}

// Close shuts down cache resources
func (cl *CacheLayer) Close() {
	cl.memory.Close()
	cl.disk.Close()
}

// io_NopCloser2 creates an io.ReadCloser from a byte slice
func io_NopCloser2(data []byte) io.ReadCloser {
	return io.NopCloser(bytes.NewReader(data))
}
