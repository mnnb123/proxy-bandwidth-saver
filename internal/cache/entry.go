package cache

import (
	"bytes"
	"encoding/gob"
	"net/http"
	"time"
)

type CacheEntry struct {
	StatusCode int         `json:"statusCode"`
	Headers    http.Header `json:"headers"`
	Body       []byte      `json:"body"`
	TTL        time.Duration `json:"ttl"`
	StoredAt   time.Time   `json:"storedAt"`
	Size       int64       `json:"size"`
}

func (e *CacheEntry) IsExpired() bool {
	if e.TTL <= 0 {
		return true
	}
	return time.Since(e.StoredAt) > e.TTL
}

func (e *CacheEntry) ToResponse(req *http.Request) *http.Response {
	return &http.Response{
		StatusCode: e.StatusCode,
		Header:     e.Headers.Clone(),
		Body:       io_NopCloser(bytes.NewReader(e.Body)),
		Request:    req,
	}
}

func (e *CacheEntry) Encode() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(e); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func DecodeEntry(data []byte) (*CacheEntry, error) {
	var entry CacheEntry
	dec := gob.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(&entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

// io_NopCloser wraps a reader with a no-op closer
func io_NopCloser(r *bytes.Reader) *nopCloserReader {
	return &nopCloserReader{r}
}

type nopCloserReader struct {
	*bytes.Reader
}

func (n *nopCloserReader) Close() error { return nil }
