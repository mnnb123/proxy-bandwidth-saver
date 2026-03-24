package cache

import (
	"log"
	"sync/atomic"

	bolt "go.etcd.io/bbolt"
)

var cacheBucket = []byte("cache")

type DiskCache struct {
	db       *bolt.DB
	maxSize  int64
	usedSize atomic.Int64
}

func NewDiskCache(path string, maxSizeMB int) (*DiskCache, error) {
	db, err := bolt.Open(path, 0600, &bolt.Options{
		NoSync: true, // async sync for performance
	})
	if err != nil {
		return nil, err
	}

	// Create bucket
	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(cacheBucket)
		return err
	}); err != nil {
		db.Close()
		return nil, err
	}

	dc := &DiskCache{
		db:      db,
		maxSize: int64(maxSizeMB) * 1024 * 1024,
	}

	// Calculate current size
	dc.recalcSize()
	return dc, nil
}

func (d *DiskCache) Get(key string) (*CacheEntry, bool) {
	var data []byte
	d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(cacheBucket)
		if b == nil {
			return nil
		}
		v := b.Get([]byte(key))
		if v != nil {
			data = make([]byte, len(v))
			copy(data, v)
		}
		return nil
	})

	if data == nil {
		return nil, false
	}

	entry, err := DecodeEntry(data)
	if err != nil {
		return nil, false
	}

	if entry.IsExpired() {
		d.Delete(key)
		return nil, false
	}

	return entry, true
}

func (d *DiskCache) Set(key string, entry *CacheEntry) {
	data, err := entry.Encode()
	if err != nil {
		return
	}

	// Check size limit
	if d.usedSize.Load()+int64(len(data)) > d.maxSize {
		d.evictOldest(int64(len(data)))
	}

	d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(cacheBucket)
		if b == nil {
			return nil
		}
		return b.Put([]byte(key), data)
	})

	d.usedSize.Add(int64(len(data)))
}

func (d *DiskCache) Delete(key string) {
	d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(cacheBucket)
		if b == nil {
			return nil
		}
		v := b.Get([]byte(key))
		if v != nil {
			d.usedSize.Add(-int64(len(v)))
		}
		return b.Delete([]byte(key))
	})
}

func (d *DiskCache) Clear() {
	d.db.Update(func(tx *bolt.Tx) error {
		tx.DeleteBucket(cacheBucket)
		_, err := tx.CreateBucket(cacheBucket)
		return err
	})
	d.usedSize.Store(0)
}

func (d *DiskCache) Close() error {
	return d.db.Close()
}

func (d *DiskCache) UsedSize() int64 {
	return d.usedSize.Load()
}

func (d *DiskCache) recalcSize() {
	var total int64
	d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(cacheBucket)
		if b == nil {
			return nil
		}
		b.ForEach(func(k, v []byte) error {
			total += int64(len(v))
			return nil
		})
		return nil
	})
	d.usedSize.Store(total)
}

func (d *DiskCache) evictOldest(needed int64) {
	var keysToDelete []string
	var freed int64

	d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(cacheBucket)
		if b == nil {
			return nil
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil && freed < needed; k, v = c.Next() {
			keysToDelete = append(keysToDelete, string(k))
			freed += int64(len(v))
		}
		return nil
	})

	for _, key := range keysToDelete {
		d.Delete(key)
	}

	if len(keysToDelete) > 0 {
		log.Printf("Cache evicted %d entries, freed %d bytes", len(keysToDelete), freed)
	}
}
