package webapi

import (
	"fmt"
	"strings"
)

// GeneratePAC generates a PAC (Proxy Auto-Config) file content.
// Bypass domains connect directly (client's own IP).
// Everything else goes through the specified proxy address.
func (a *HeadlessApp) GeneratePAC(proxyAddr string) string {
	bypassDomains, bypassKeywords := a.getBypassPatterns()

	var sb strings.Builder
	sb.WriteString("function FindProxyForURL(url, host) {\n")

	// Bypass rules: exact domains + wildcards
	if len(bypassDomains) > 0 {
		sb.WriteString("  // Bypass domains - connect directly (your local IP)\n")
		for _, d := range bypassDomains {
			if strings.HasPrefix(d, "*.") {
				// Wildcard: *.example.com -> match example.com and all subdomains
				base := d[2:]
				sb.WriteString(fmt.Sprintf("  if (dnsDomainIs(host, \"%s\") || dnsDomainIs(host, \".%s\")) return \"DIRECT\";\n", base, base))
			} else {
				// Exact domain + subdomains
				sb.WriteString(fmt.Sprintf("  if (dnsDomainIs(host, \"%s\") || dnsDomainIs(host, \".%s\")) return \"DIRECT\";\n", d, d))
			}
		}
		sb.WriteString("\n")
	}

	// Keyword rules
	if len(bypassKeywords) > 0 {
		sb.WriteString("  // Keyword bypass - any domain containing these words\n")
		for _, kw := range bypassKeywords {
			sb.WriteString(fmt.Sprintf("  if (host.indexOf(\"%s\") >= 0) return \"DIRECT\";\n", kw))
		}
		sb.WriteString("\n")
	}

	// Default: use proxy
	sb.WriteString(fmt.Sprintf("  return \"PROXY %s\";\n", proxyAddr))
	sb.WriteString("}\n")

	return sb.String()
}

// getBypassPatterns returns all bypass domain patterns and keywords from the database.
func (a *HeadlessApp) getBypassPatterns() (domains []string, keywords []string) {
	if a.db == nil {
		return nil, nil
	}
	rows, err := a.db.Reader.Query(
		"SELECT pattern FROM rules WHERE action = 'bypass' AND rule_type = 'domain' AND enabled = 1",
	)
	if err != nil {
		return nil, nil
	}
	defer rows.Close()

	for rows.Next() {
		var pattern string
		if err := rows.Scan(&pattern); err != nil {
			continue
		}
		pattern = strings.TrimSpace(strings.ToLower(pattern))
		if pattern == "" {
			continue
		}
		if !strings.Contains(pattern, ".") && !strings.Contains(pattern, "*") {
			// Keyword: no dots, no wildcards
			keywords = append(keywords, pattern)
		} else {
			domains = append(domains, pattern)
		}
	}
	return domains, keywords
}
