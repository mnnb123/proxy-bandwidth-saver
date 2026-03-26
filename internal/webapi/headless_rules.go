package webapi

import (
	"fmt"
	"net/http"
	"strings"

	"proxy-bandwidth-saver/internal/classifier"
	"proxy-bandwidth-saver/internal/database"
)

func (a *HeadlessApp) GetRulesJSON() interface{} {
	return a.getRules()
}

func (a *HeadlessApp) getRules() []database.Rule {
	if a.db == nil {
		return nil
	}
	rows, err := a.db.Reader.Query(
		"SELECT id, rule_type, pattern, action, priority, enabled, hit_count, bytes_saved, created_at FROM rules ORDER BY priority DESC, id ASC",
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var rules []database.Rule
	for rows.Next() {
		var r database.Rule
		var enabled int
		var createdAt string
		if err := rows.Scan(&r.ID, &r.RuleType, &r.Pattern, &r.Action, &r.Priority, &enabled, &r.HitCount, &r.BytesSaved, &createdAt); err != nil {
			continue
		}
		r.Enabled = enabled == 1
		r.CreatedAt = createdAt
		rules = append(rules, r)
	}
	return rules
}

func (a *HeadlessApp) AddRule(ruleType, pattern, action string, priority int) error {
	if a.db == nil {
		return fmt.Errorf("not initialized")
	}
	_, err := classifier.CreateRule(a.db.Writer, ruleType, pattern, action, priority)
	if err == nil {
		a.reloadClassifier()
	}
	return err
}

func (a *HeadlessApp) UpdateRuleById(id int, ruleType, pattern, action string, priority int, enabled bool) error {
	if a.db == nil {
		return fmt.Errorf("not initialized")
	}
	err := classifier.UpdateRule(a.db.Writer, id, ruleType, pattern, action, priority, enabled)
	if err == nil {
		a.reloadClassifier()
	}
	return err
}

func (a *HeadlessApp) DeleteRule(id int) error {
	if a.db == nil {
		return fmt.Errorf("not initialized")
	}
	err := classifier.DeleteRule(a.db.Writer, id)
	if err == nil {
		a.reloadClassifier()
	}
	return err
}

func (a *HeadlessApp) ToggleRule(id int, enabled bool) error {
	if a.db == nil {
		return fmt.Errorf("not initialized")
	}
	err := classifier.ToggleRule(a.db.Writer, id, enabled)
	if err == nil {
		a.reloadClassifier()
	}
	return err
}

func (a *HeadlessApp) TestRule(domain, urlPath, contentType string) string {
	if a.classifier == nil {
		return "residential"
	}
	req, _ := http.NewRequest("GET", "http://"+domain+urlPath, nil)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	route := a.classifier.Classify(req)
	return string(route)
}

func (a *HeadlessApp) ClearAllRules() error {
	if a.db == nil {
		return fmt.Errorf("not initialized")
	}
	_, err := a.db.Writer.Exec("DELETE FROM rules")
	if err == nil {
		a.reloadClassifier()
	}
	return err
}

func (a *HeadlessApp) AddBulkRules(patterns []string, action string, priority int) (int, error) {
	if a.db == nil {
		return 0, fmt.Errorf("not initialized")
	}
	count := 0
	for _, p := range patterns {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if _, err := classifier.CreateRule(a.db.Writer, "domain", p, action, priority); err == nil {
			count++
		}
	}
	if count > 0 {
		a.reloadClassifier()
	}
	return count, nil
}

func (a *HeadlessApp) ImportRules(jsonStr string) int {
	if a.db == nil {
		return 0
	}
	var rules []struct {
		RuleType string `json:"ruleType"`
		Pattern  string `json:"pattern"`
		Action   string `json:"action"`
		Priority int    `json:"priority"`
		Enabled  bool   `json:"enabled"`
	}
	if err := jsonUnmarshal(jsonStr, &rules); err != nil {
		return 0
	}
	count := 0
	for _, r := range rules {
		if _, err := classifier.CreateRule(a.db.Writer, r.RuleType, r.Pattern, r.Action, r.Priority); err == nil {
			count++
		}
	}
	if count > 0 {
		a.reloadClassifier()
	}
	return count
}

func (a *HeadlessApp) ExportRules() string {
	rules := a.getRules()
	data, _ := jsonMarshal(rules)
	return string(data)
}
