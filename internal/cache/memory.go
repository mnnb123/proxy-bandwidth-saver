package cache

import (
	"github.com/dgraph-io/ristretto/v2"
)

type MemoryCache struct {
	cache *ristretto.Cache[string, *CacheEntry]
}

func NewMemoryCache(maxSizeMB int) (*MemoryCache, error) {
	maxCost := int64(maxSizeMB) * 1024 * 1024

	cache, err := ristretto.NewCache(&ristretto.Config[string, *CacheEntry]{
		NumCounters: maxCost / 100, // ~10x expected items (avg 100 bytes)
		MaxCost:     maxCost,
		BufferItems: 64,
	})
	if err != nil {
		return nil, err
	}

	return &MemoryCache{cache: cache}, nil
}

func (m *MemoryCache) Get(key string) (*CacheEntry, bool) {
	entry, found := m.cache.Get(key)
	if !found || entry == nil {
		return nil, false
	}
	if entry.IsExpired() {
		m.cache.Del(key)
		return nil, false
	}
	return entry, true
}

func (m *MemoryCache) Set(key string, entry *CacheEntry) {
	cost := int64(len(entry.Body))
	if cost == 0 {
		cost = 1
	}
	m.cache.SetWithTTL(key, entry, cost, entry.TTL)
}

func (m *MemoryCache) Delete(key string) {
	m.cache.Del(key)
}

func (m *MemoryCache) Clear() {
	m.cache.Clear()
}

func (m *MemoryCache) Close() {
	m.cache.Close()
}
