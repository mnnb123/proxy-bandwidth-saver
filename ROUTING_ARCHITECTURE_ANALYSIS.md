# Proxy Traffic Routing Architecture Analysis

## Overview

The proxy-bandwidth-saver architecture has **two distinct request paths**, and only the main proxy path (8888/8889) applies domain-based classification rules. The output port forwarders (ports 30000+) bypass classification entirely.

## Path 1: Main Proxy (8888/8889) - Classification-Aware

Main proxy on ports 8888 (HTTP) and 8889 (SOCKS5) uses goproxy with integrated classification.

### Classification Pipeline (server.go lines 233-287)

1. IP Whitelist Check
2. Proxy-Authorization Check  
3. **CLASSIFY REQUEST** via s.pipeline.Classifier(req) (line 262)
4. Block Check - RouteBlock returns 204 No Content
5. Cache Check (checks before forwarding)
6. Select Transport Based on Route (line 279)

Routes: RouteDirect, RouteDatacenter, RouteResidential, RouteBlock

### Key Points

- Classification happens BEFORE forwarding (line 262)
- Domain rules are applied (RouteBlock, RouteDirect, RouteDatacenter, RouteResidential)
- Cache integration: Responses cached per route
- Full bandwidth accounting by route

## Path 2: Output Port Forwarders (30000+) - Classification-Bypassed

PortMapper.MapProxy() creates two listeners per upstream:
- HTTP: port 30000+ (even)
- SOCKS5: port 30001+ (odd)

These forward DIRECTLY to upstream proxy with NO classification.

### HTTP Forwarder (forwarder.go)

handleHTTP (lines 54-81):
- Extracts domain from request.Host
- DIRECTLY forwards via http.Transport (line 62)
- NO classification, NO cache check, NO route decision
- Just meters bytes (line 79)

handleConnect (lines 84-133):
- Extracts domain but doesn't use it for classification
- Tunnels to upstream with byte counting
- NO route decision, just meters (line 131)

### SOCKS5 Forwarder (socks5_forwarder.go)

handleConn (lines 30-125):
1. SOCKS5 greeting and auth
2. Parse target address (line 72)
3. Connect upstream (line 89)
4. Relay with byte counting
- NO classification, NO route decision
- Just meters bytes (line 123)

## Classifier (internal/classifier/classifier.go)

Classify() method (lines 65-120) applies rules in order:

1. **Exact Domain Match** (O(1)) -> ExactDomains map
2. **Wildcard Domain Match** (suffix) -> WildcardDomains list
3. **Keyword Domain Match** (contains) -> KeywordDomains list
4. **URL Pattern Match** (regex) -> URLPatterns list
5. **Content-Type Heuristic** (Accept header) -> ContentTypes map
6. **Default:** RouteResidential

### Default Rules (defaults.go)

- CDN domains (priority 100) -> RouteDirect
- Analytics (priority 90) -> RouteDirect
- Fonts (priority 95) -> RouteDirect
- Social media (priority 80) -> RouteDirect
- Static resources (priority 60) -> RouteDirect
- API paths (priority 55) -> RouteResidential
- Content-Type rules -> RouteDirect or RouteResidential

## The Gap

### Current Output Port Behavior

Request: GET http://cdn.example.com/image.jpg to port 30000

1. Classify: SKIPPED
2. RouteBlock: SKIPPED
3. Cache: SKIPPED
4. Transport: Forced to upstream
5. Meter: domain only, no route

### Main Proxy Behavior (for comparison)

Same request to port 8888:

1. Classify: cdn.example.com matches *.cloudflare.com -> RouteDirect
2. RouteBlock: false
3. Cache: check/store
4. Transport: Select based on route
5. Meter: tracked as RouteDirect

## Consequences of the Gap

1. Bypass rules ignored: *.cdn.example.com rules never applied
2. Block rules ignored: Custom blocks never enforced
3. No caching: Even cacheable resources bypass cache
4. Wrong upstream: All traffic goes to single upstream, no routing
5. Incomplete metering: Stats don't show which route each request used

## Solution: Add Classification to Output Port Forwarders

### Changes Needed

1. **Pass Classifier to forwarders**
   - Add classifier *Classifier field to proxyForwarder
   - Add classifier *Classifier field to socks5Listener
   - PortMapper holds classifier reference

2. **Classify before forwarding**
   - Extract domain from request
   - Check RouteBlock -> return 204
   - Classify to determine route

3. **Route Decision Logic**
   - RouteDirect: Send to direct upstream
   - RouteDatacenter: Send to datacenter upstream
   - RouteResidential: Send to residential upstream
   - RouteBlock: Return 204 No Content

4. **Update metering**
   - Pass route info to meter callback
   - Track which route each request used

5. **Optional: Add caching**
   - Integrate cache layer

### Implementation Files

1. internal/proxy/portmap.go
   - Add classifier field to PortMapper
   - Pass classifier to MapProxy()

2. internal/proxy/forwarder.go
   - Add classifier field to proxyForwarder
   - Classify in handleHTTP() and handleConnect()
   - Check RouteBlock before forwarding

3. internal/proxy/socks5_forwarder.go
   - Add classifier field to socks5Listener
   - Classify in handleConn()
   - Check RouteBlock before forwarding

4. app.go / app_proxies.go
   - Get classifier instance
   - Pass to PortMapper when creating forwarders

## Data Flow After Fix

Output port request with classification:

1. Client -> proxyForwarder.handleHTTP()
2. Extract domain: cdn.example.com
3. Classify(domain) -> RouteDirect
4. RouteBlock check: false (proceed)
5. SELECT UPSTREAM BASED ON ROUTE
6. Meter(domain, bytes, route, proxyID)

## Summary

| Feature | Main Proxy | Output Ports | Status |
|---------|-----------|--------------|--------|
| Domain classification | Yes | No | **GAP** |
| Block rules | Yes | No | **GAP** |
| Bypass rules | Yes | No | **GAP** |
| Route selection | Yes | No | **GAP** |
| Cache integration | Yes | No | Optional |
| Metering | Full | Partial | Needs route |
| Auth checks | Yes | Yes | Good |
| Upstream selection | Dynamic | Fixed | **GAP** |

