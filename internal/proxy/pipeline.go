package proxy

import (
	"net/http"
	"time"
)

// Route represents how traffic should be routed
type Route string

const (
	RouteDirect      Route = "direct"
	RouteDatacenter  Route = "datacenter"
	RouteResidential Route = "residential"
	RouteBlock       Route = "block"
)

// RequestCtx carries per-request state through the pipeline
type RequestCtx struct {
	Request      *http.Request
	Route        Route
	CacheKey     string
	StartTime    time.Time
	ProxyID      int
	ReqBytes     int64
	RespBytes    int64
	Domain       string
	Cached       bool

	// internal: response body byte counter (set by server, read by meter)
	respCounter interface{ BytesRead() int64 }
}

// TotalBytes returns the total bytes for this request (req + resp)
// Call after response body has been fully consumed by client
func (ctx *RequestCtx) TotalBytes() int64 {
	resp := ctx.RespBytes
	if ctx.respCounter != nil {
		resp = ctx.respCounter.BytesRead()
	}
	return ctx.ReqBytes + resp
}

// ClassifierFunc classifies a request and returns the route
type ClassifierFunc func(req *http.Request) Route

// CacheCheckFunc checks cache for a response. Returns nil if miss.
type CacheCheckFunc func(ctx *RequestCtx) *http.Response

// CacheStoreFunc stores a response in cache
type CacheStoreFunc func(ctx *RequestCtx, resp *http.Response)

// MeterFunc records bandwidth for a completed request
type MeterFunc func(ctx *RequestCtx)

// Pipeline holds all middleware hooks
type Pipeline struct {
	Classifier ClassifierFunc
	CacheCheck CacheCheckFunc
	CacheStore CacheStoreFunc
	Meter      MeterFunc
}

// NewDefaultPipeline returns pipeline with no-op hooks
func NewDefaultPipeline() *Pipeline {
	return &Pipeline{
		Classifier: func(req *http.Request) Route {
			return RouteResidential // conservative default
		},
		CacheCheck: func(ctx *RequestCtx) *http.Response {
			return nil // no cache
		},
		CacheStore: func(ctx *RequestCtx, resp *http.Response) {
			// no-op
		},
		Meter: func(ctx *RequestCtx) {
			// no-op
		},
	}
}
