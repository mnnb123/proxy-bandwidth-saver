package main

import (
	"encoding/json"
	"fmt"
	"log"

	"proxy-bandwidth-saver/internal/classifier"
	"proxy-bandwidth-saver/internal/database"
	"proxy-bandwidth-saver/internal/proxy"
)

// GetRules returns all classification rules ordered by priority.
func (a *App) GetRules() []database.Rule {
	if a.db == nil {
		return nil
	}
	rows, err := a.db.Reader.Query(
		"SELECT id, rule_type, pattern, action, priority, enabled, hit_count, bytes_saved, created_at FROM rules ORDER BY priority DESC, id ASC",
	)
	if err != nil {
		log.Printf("Failed to get rules: %v", err)
		return nil
	}
	defer rows.Close()

	var rules []database.Rule
	for rows.Next() {
		var r database.Rule
		var enabled int
		if err := rows.Scan(&r.ID, &r.RuleType, &r.Pattern, &r.Action, &r.Priority, &enabled, &r.HitCount, &r.BytesSaved, &r.CreatedAt); err != nil {
			continue
		}
		r.Enabled = enabled == 1
		rules = append(rules, r)
	}
	return rules
}

func (a *App) AddRule(ruleType, pattern, action string, priority int) error {
	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}
	_, err := classifier.CreateRule(a.db.Writer, ruleType, pattern, action, priority)
	if err != nil {
		return err
	}
	a.reloadClassifier()
	return nil
}

func (a *App) UpdateRuleById(id int, ruleType, pattern, action string, priority int, enabled bool) error {
	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}
	if err := classifier.UpdateRule(a.db.Writer, id, ruleType, pattern, action, priority, enabled); err != nil {
		return err
	}
	a.reloadClassifier()
	return nil
}

func (a *App) DeleteRule(id int) error {
	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}
	if err := classifier.DeleteRule(a.db.Writer, id); err != nil {
		return err
	}
	a.reloadClassifier()
	return nil
}

func (a *App) ToggleRule(id int, enabled bool) error {
	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}
	if err := classifier.ToggleRule(a.db.Writer, id, enabled); err != nil {
		return err
	}
	a.reloadClassifier()
	return nil
}

func (a *App) TestRule(domain, urlPath, contentType string) string {
	if a.classifier == nil {
		return string(proxy.RouteResidential)
	}
	return string(a.classifier.TestClassify(domain, urlPath, contentType))
}

func (a *App) ImportRules(jsonStr string) (int, error) {
	var rules []struct {
		RuleType string `json:"ruleType"`
		Pattern  string `json:"pattern"`
		Action   string `json:"action"`
		Priority int    `json:"priority"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &rules); err != nil {
		return 0, fmt.Errorf("invalid JSON: %w", err)
	}
	count := 0
	for _, r := range rules {
		_, err := classifier.CreateRule(a.db.Writer, r.RuleType, r.Pattern, r.Action, r.Priority)
		if err == nil {
			count++
		}
	}
	a.reloadClassifier()
	return count, nil
}

func (a *App) ExportRules() string {
	rules := a.GetRules()
	data, _ := json.MarshalIndent(rules, "", "  ")
	return string(data)
}
