package classifier

import (
	"net"
	"net/http"
	"path"
	"regexp"
	"strings"
	"sync/atomic"

	"proxy-bandwidth-saver/internal/proxy"
)

// CompiledRules is the pre-compiled, read-optimized rule set
type CompiledRules struct {
	// O(1) exact domain match
	ExactDomains map[string]proxy.Route

	// Wildcard suffix rules (sorted longest first for best match)
	WildcardDomains []WildcardRule

	// Keyword rules (domain contains keyword)
	KeywordDomains []KeywordRule

	// URL pattern rules (pre-compiled regex)
	URLPatterns []URLPatternRule

	// Content-type rules (exact match)
	ContentTypes map[string]proxy.Route

	// Static resource extensions -> direct
	StaticExtensions map[string]bool
}

type WildcardRule struct {
	Suffix   string
	Route    proxy.Route
	Priority int
}

type KeywordRule struct {
	Keyword  string
	Route    proxy.Route
	Priority int
}

type URLPatternRule struct {
	Regex    *regexp.Regexp
	Route    proxy.Route
	Priority int
}

// Classifier classifies requests into routes using compiled rules
type Classifier struct {
	rules atomic.Value // stores *CompiledRules
}

func NewClassifier() *Classifier {
	c := &Classifier{}
	c.rules.Store(defaultCompiledRules())
	return c
}

// Classify determines the route for a request - designed for <1µs on exact match
func (c *Classifier) Classify(req *http.Request) proxy.Route {
	rules := c.rules.Load().(*CompiledRules)

	domain := extractDomain(req.Host)

	// 1. Exact domain match (O(1))
	if route, ok := rules.ExactDomains[domain]; ok {
		return route
	}

	// 2. Wildcard domain match (suffix scan)
	for _, wr := range rules.WildcardDomains {
		if strings.HasSuffix(domain, wr.Suffix) || domain == strings.TrimPrefix(wr.Suffix, ".") {
			return wr.Route
		}
	}

	// 2.5 Keyword domain match (contains)
	for _, kr := range rules.KeywordDomains {
		if strings.Contains(domain, kr.Keyword) {
			return kr.Route
		}
	}

	// 3. URL pattern match (only when URL path is visible - HTTP or MITM)
	if req.URL != nil && req.URL.Path != "" {
		urlPath := req.URL.Path

		// Static resource extension check (fast)
		ext := strings.ToLower(path.Ext(urlPath))
		if rules.StaticExtensions[ext] {
			return proxy.RouteDirect
		}

		// Regex URL patterns
		fullURL := req.URL.String()
		for _, up := range rules.URLPatterns {
			if up.Regex.MatchString(urlPath) || up.Regex.MatchString(fullURL) {
				return up.Route
			}
		}
	}

	// 4. Content-type heuristic from Accept header
	accept := req.Header.Get("Accept")
	if accept != "" {
		for ct, route := range rules.ContentTypes {
			if strings.Contains(accept, ct) {
				return route
			}
		}
	}

	// 5. Default: residential (conservative)
	return proxy.RouteResidential
}

// Reload atomically swaps the compiled rules (lock-free for readers)
func (c *Classifier) Reload(dbRules []Rule) {
	compiled := CompileRules(dbRules)
	c.rules.Store(compiled)
}

// TestClassify tests what route a given domain/URL/content-type would get
func (c *Classifier) TestClassify(domain, urlPath, contentType string) proxy.Route {
	req, _ := http.NewRequest("GET", "http://"+domain+urlPath, nil)
	if contentType != "" {
		req.Header.Set("Accept", contentType)
	}
	return c.Classify(req)
}

func extractDomain(host string) string {
	domain, _, err := net.SplitHostPort(host)
	if err != nil {
		return strings.ToLower(host)
	}
	return strings.ToLower(domain)
}
