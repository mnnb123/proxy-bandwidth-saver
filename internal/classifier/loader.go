package classifier

import (
	"database/sql"
	"fmt"
	"log"
)

// LoadRulesFromDB loads all enabled rules from the database
func LoadRulesFromDB(reader *sql.DB) ([]Rule, error) {
	rows, err := reader.Query(
		"SELECT id, rule_type, pattern, action, priority, enabled FROM rules WHERE enabled = 1 ORDER BY priority DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []Rule
	for rows.Next() {
		var r Rule
		var enabled int
		if err := rows.Scan(&r.ID, &r.RuleType, &r.Pattern, &r.Action, &r.Priority, &enabled); err != nil {
			return nil, err
		}
		r.Enabled = enabled == 1
		rules = append(rules, r)
	}
	return rules, rows.Err()
}

// SeedDefaultRules inserts default rules if the rules table is empty
func SeedDefaultRules(writer *sql.DB) error {
	var count int
	if err := writer.QueryRow("SELECT COUNT(*) FROM rules").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil // already seeded
	}

	defaults := DefaultRules()
	tx, err := writer.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(
		"INSERT INTO rules (rule_type, pattern, action, priority, enabled) VALUES (?, ?, ?, ?, 1)",
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, r := range defaults {
		if _, err := stmt.Exec(r.RuleType, r.Pattern, r.Action, r.Priority); err != nil {
			return err
		}
	}

	log.Printf("Seeded %d default rules", len(defaults))
	return tx.Commit()
}

// CRUD operations

func CreateRule(writer *sql.DB, ruleType, pattern, action string, priority int) (int64, error) {
	// Check for duplicate (same type + pattern)
	var existingID int64
	err := writer.QueryRow(
		"SELECT id FROM rules WHERE rule_type = ? AND pattern = ?",
		ruleType, pattern,
	).Scan(&existingID)
	if err == nil {
		// Rule already exists, skip
		return existingID, fmt.Errorf("rule already exists: %s %s", ruleType, pattern)
	}

	result, err := writer.Exec(
		"INSERT INTO rules (rule_type, pattern, action, priority) VALUES (?, ?, ?, ?)",
		ruleType, pattern, action, priority,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func UpdateRule(writer *sql.DB, id int, ruleType, pattern, action string, priority int, enabled bool) error {
	enabledInt := 0
	if enabled {
		enabledInt = 1
	}
	_, err := writer.Exec(
		"UPDATE rules SET rule_type = ?, pattern = ?, action = ?, priority = ?, enabled = ? WHERE id = ?",
		ruleType, pattern, action, priority, enabledInt, id,
	)
	return err
}

func DeleteRule(writer *sql.DB, id int) error {
	_, err := writer.Exec("DELETE FROM rules WHERE id = ?", id)
	return err
}

func ToggleRule(writer *sql.DB, id int, enabled bool) error {
	enabledInt := 0
	if enabled {
		enabledInt = 1
	}
	_, err := writer.Exec("UPDATE rules SET enabled = ? WHERE id = ?", enabledInt, id)
	return err
}

func IncrementRuleHit(writer *sql.DB, id int, bytesSaved int64) {
	_, _ = writer.Exec(
		"UPDATE rules SET hit_count = hit_count + 1, bytes_saved = bytes_saved + ? WHERE id = ?",
		bytesSaved, id,
	)
}
