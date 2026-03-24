package classifier

// DefaultRules returns the standard bypass rules for CDNs, analytics, fonts, tracking
func DefaultRules() []Rule {
	rules := make([]Rule, 0, 80)
	id := 0

	add := func(ruleType, pattern, action string, priority int) {
		id++
		rules = append(rules, Rule{
			ID:       id,
			RuleType: ruleType,
			Pattern:  pattern,
			Action:   action,
			Priority: priority,
			Enabled:  true,
		})
	}

	// === CDN domains -> DIRECT (priority 100) ===
	cdnDomains := []string{
		"*.cloudflare.com", "*.cloudflare-dns.com",
		"*.cloudfront.net", "*.fastly.net", "*.akamaized.net",
		"*.jsdelivr.net", "*.unpkg.com", "*.cdnjs.cloudflare.com",
		"*.bootstrapcdn.com", "*.googleapis.com", "*.gstatic.com",
		"*.google.com", "*.googlevideo.com",
		"cdn.jsdelivr.net", "unpkg.com",
	}
	for _, d := range cdnDomains {
		add("domain", d, "direct", 100)
	}

	// === Analytics & Tracking -> DIRECT (priority 90) ===
	analyticsDomains := []string{
		"*.google-analytics.com", "*.googletagmanager.com",
		"*.googleadservices.com", "*.googlesyndication.com",
		"*.doubleclick.net", "*.adsense.com",
		"*.facebook.net", "*.facebook.com",
		"*.hotjar.com", "*.clarity.ms",
		"*.mixpanel.com", "*.segment.io", "*.segment.com",
		"*.amplitude.com", "*.heap.io",
		"*.newrelic.com", "*.nr-data.net",
		"*.sentry.io", "*.bugsnag.com",
	}
	for _, d := range analyticsDomains {
		add("domain", d, "direct", 90)
	}

	// === Fonts -> DIRECT (priority 95) ===
	fontDomains := []string{
		"fonts.googleapis.com", "fonts.gstatic.com",
		"use.fontawesome.com", "*.typekit.net",
		"use.typekit.net",
	}
	for _, d := range fontDomains {
		add("domain", d, "direct", 95)
	}

	// === Social media / common services -> DIRECT (priority 80) ===
	socialDomains := []string{
		"*.twitter.com", "*.x.com",
		"*.youtube.com", "*.ytimg.com",
		"*.instagram.com", "*.whatsapp.com",
		"*.tiktok.com", "*.tiktokcdn.com",
		"*.reddit.com", "*.redd.it",
		"*.github.com", "*.githubusercontent.com",
		"*.stackoverflow.com",
	}
	for _, d := range socialDomains {
		add("domain", d, "direct", 80)
	}

	// === Content-Type rules (priority 50) ===
	// Static content -> DIRECT
	add("content_type", "image/*", "direct", 50)
	add("content_type", "font/*", "direct", 50)
	add("content_type", "text/css", "direct", 50)
	add("content_type", "application/javascript", "direct", 50)
	add("content_type", "application/x-javascript", "direct", 50)

	// Dynamic content -> RESIDENTIAL
	add("content_type", "text/html", "residential", 40)
	add("content_type", "application/json", "residential", 40)
	add("content_type", "application/xml", "residential", 40)

	// === URL pattern rules (priority 60) ===
	// Static paths -> DIRECT
	staticPaths := []string{
		"/static/*", "/assets/*", "/cdn-cgi/*",
		"/_next/static/*", "/_next/image/*",
		"/wp-content/uploads/*", "/wp-content/themes/*",
		"/media/*", "/images/*", "/img/*",
		"/fonts/*", "/css/*", "/js/*",
	}
	for _, p := range staticPaths {
		add("url_pattern", p, "direct", 60)
	}

	// API paths -> RESIDENTIAL
	apiPaths := []string{
		"/api/*", "/graphql", "/graphql/*",
		"/v1/*", "/v2/*", "/v3/*",
	}
	for _, p := range apiPaths {
		add("url_pattern", p, "residential", 55)
	}

	return rules
}

func defaultCompiledRules() *CompiledRules {
	return CompileRules(DefaultRules())
}
