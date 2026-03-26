package classifier

import (
	"regexp"
	"sort"
	"strings"

	"proxy-bandwidth-saver/internal/proxy"
)

// Rule represents a traffic classification rule from the database
type Rule struct {
	ID       int    `json:"id"`
	RuleType string `json:"ruleType"` // domain | content_type | url_pattern
	Pattern  string `json:"pattern"`
	Action   string `json:"action"` // direct | datacenter | residential | block
	Priority int    `json:"priority"`
	Enabled  bool   `json:"enabled"`
}

func parseRoute(action string) proxy.Route {
	switch strings.ToLower(action) {
	case "direct", "bypass", "bypass_vps":
		return proxy.RouteDirect
	case "datacenter":
		return proxy.RouteDatacenter
	case "residential":
		return proxy.RouteResidential
	case "block":
		return proxy.RouteBlock
	default:
		return proxy.RouteResidential
	}
}

// CompileRules compiles database rules into an optimized lookup structure
func CompileRules(rules []Rule) *CompiledRules {
	compiled := &CompiledRules{
		ExactDomains:     make(map[string]proxy.Route),
		WildcardDomains:  make([]WildcardRule, 0),
		URLPatterns:      make([]URLPatternRule, 0),
		ContentTypes:     make(map[string]proxy.Route),
		StaticExtensions: defaultStaticExtensions(),
	}

	// Sort by priority descending
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Priority > rules[j].Priority
	})

	for _, r := range rules {
		if !r.Enabled {
			continue
		}
		route := parseRoute(r.Action)

		switch r.RuleType {
		case "domain":
			compileDomainRule(compiled, r.Pattern, route, r.Priority)
		case "content_type":
			compiled.ContentTypes[strings.ToLower(r.Pattern)] = route
		case "url_pattern":
			compileURLRule(compiled, r.Pattern, route, r.Priority)
		}
	}

	// Sort wildcard rules by suffix length desc (longest match first)
	sort.Slice(compiled.WildcardDomains, func(i, j int) bool {
		if len(compiled.WildcardDomains[i].Suffix) != len(compiled.WildcardDomains[j].Suffix) {
			return len(compiled.WildcardDomains[i].Suffix) > len(compiled.WildcardDomains[j].Suffix)
		}
		return compiled.WildcardDomains[i].Priority > compiled.WildcardDomains[j].Priority
	})

	return compiled
}

func compileDomainRule(c *CompiledRules, pattern string, route proxy.Route, priority int) {
	pattern = strings.ToLower(strings.TrimSpace(pattern))

	if strings.HasPrefix(pattern, "*.") {
		// Wildcard: *.example.com -> suffix .example.com
		suffix := pattern[1:] // ".example.com"
		c.WildcardDomains = append(c.WildcardDomains, WildcardRule{
			Suffix:   suffix,
			Route:    route,
			Priority: priority,
		})
	} else if strings.Contains(pattern, "*") {
		// Complex wildcard -> compile as regex
		regexPattern := "^" + strings.ReplaceAll(regexp.QuoteMeta(pattern), `\*`, `.*`) + "$"
		if re, err := regexp.Compile(regexPattern); err == nil {
			c.URLPatterns = append(c.URLPatterns, URLPatternRule{
				Regex:    re,
				Route:    route,
				Priority: priority,
			})
		}
	} else {
		// Exact domain match
		c.ExactDomains[pattern] = route
	}
}

func compileURLRule(c *CompiledRules, pattern string, route proxy.Route, priority int) {
	// Convert simple glob patterns to regex
	regexStr := pattern
	if !strings.HasPrefix(pattern, "^") {
		// Simple glob pattern like /static/*, /api/*
		regexStr = "^" + strings.ReplaceAll(regexp.QuoteMeta(pattern), `\*`, `.*`)
	}

	re, err := regexp.Compile(regexStr)
	if err != nil {
		return // skip invalid patterns
	}

	c.URLPatterns = append(c.URLPatterns, URLPatternRule{
		Regex:    re,
		Route:    route,
		Priority: priority,
	})
}

func defaultStaticExtensions() map[string]bool {
	return map[string]bool{
		".css":   true,
		".js":    true,
		".png":   true,
		".jpg":   true,
		".jpeg":  true,
		".gif":   true,
		".svg":   true,
		".ico":   true,
		".woff":  true,
		".woff2": true,
		".ttf":   true,
		".eot":   true,
		".otf":   true,
		".mp4":   true,
		".webm":  true,
		".webp":  true,
		".avif":  true,
	}
}
