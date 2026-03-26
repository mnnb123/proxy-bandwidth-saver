package database

import (
	"crypto/rand"
	"database/sql"
	_ "embed"
	"encoding/hex"
	"fmt"
	"log"
	"strconv"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

type DB struct {
	Writer *sql.DB
	Reader *sql.DB
}

func NewDB(dbPath string) (*DB, error) {
	writer, err := openDB(dbPath, 1)
	if err != nil {
		return nil, fmt.Errorf("open writer: %w", err)
	}

	reader, err := openDB(dbPath, 4)
	if err != nil {
		writer.Close()
		return nil, fmt.Errorf("open reader: %w", err)
	}

	db := &DB{Writer: writer, Reader: reader}

	if err := db.applyPragmas(writer); err != nil {
		db.Close()
		return nil, fmt.Errorf("apply pragmas: %w", err)
	}
	if err := db.applyPragmas(reader); err != nil {
		db.Close()
		return nil, fmt.Errorf("apply reader pragmas: %w", err)
	}

	if err := db.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return db, nil
}

func openDB(path string, maxConns int) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(maxConns)
	return db, nil
}

func (db *DB) applyPragmas(conn *sql.DB) error {
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA synchronous=NORMAL",
		"PRAGMA foreign_keys=ON",
		"PRAGMA busy_timeout=5000",
		"PRAGMA cache_size=-8000", // 8MB
		"PRAGMA temp_store=MEMORY",
	}
	for _, p := range pragmas {
		if _, err := conn.Exec(p); err != nil {
			return fmt.Errorf("%s: %w", p, err)
		}
	}
	return nil
}

func (db *DB) migrate() error {
	_, err := db.Writer.Exec(schemaSQL)
	return err
}

func (db *DB) Close() error {
	var errs []error
	if db.Writer != nil {
		if err := db.Writer.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if db.Reader != nil {
		if err := db.Reader.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

// Settings helpers

func (db *DB) GetSetting(key string) (string, error) {
	var value string
	err := db.Reader.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

func (db *DB) SetSetting(key, value string) error {
	_, err := db.Writer.Exec(
		"INSERT INTO settings (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value",
		key, value,
	)
	return err
}

func (db *DB) GetAllSettings() (map[string]string, error) {
	rows, err := db.Reader.Query("SELECT key, value FROM settings")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		settings[k] = v
	}
	return settings, rows.Err()
}

// generateSecurePassword returns a cryptographically random hex string of
// the given byte-length (the returned string is twice as long).
func generateSecurePassword(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("crypto/rand: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func (db *DB) SeedDefaults() error {
	defaults := map[string]string{
		"http_port":               "8888",
		"socks5_port":             "8889",
		"bind_address":            "127.0.0.1",
		"max_concurrent":          "500",
		"cache_memory_mb":         "512",
		"cache_disk_mb":           "2048",
		"monthly_budget_gb":       "0",
		"cost_per_gb":             "0",
		"auto_pause":              "false",
		"alert_threshold_80":      "true",
		"alert_threshold_95":      "true",
		"mitm_enabled":            "false",
		"accept_encoding_enforce": "true",
		"header_stripping":        "true",
		"html_minification":       "false",
		"image_recompression":     "false",
		"image_quality":           "85",
		"log_retention_days":      "7",
		"log_level":               "normal",
		"request_timeout_sec":     "30",
		"dns_cache_ttl_sec":       "300",
		"singleflight_dedup":      "true",
		"rotation_strategy":       "round-robin",
		"sticky_session_minutes":  "5",
		"web_username":            "admin",
		// web_password is handled separately below.
	}

	tx, err := db.Writer.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT OR IGNORE INTO settings (key, value) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for k, v := range defaults {
		if _, err := stmt.Exec(k, v); err != nil {
			return err
		}
	}

	// Generate a random password only on first install (INSERT OR IGNORE
	// is a no-op when the key already exists).
	password, err := generateSecurePassword(16) // 32-char hex string
	if err != nil {
		return fmt.Errorf("generate password: %w", err)
	}
	res, err := stmt.Exec("web_password", password)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows > 0 {
		log.Println("=========================================")
		log.Println("  FIRST-RUN: generated web credentials")
		log.Printf("  Username: admin")
		log.Printf("  Password: %s", password)
		log.Println("  (change these in Settings after login)")
		log.Println("=========================================")
	}

	return tx.Commit()
}

// Helper to parse setting as int
func (db *DB) GetSettingInt(key string, fallback int) int {
	val, err := db.GetSetting(key)
	if err != nil || val == "" {
		return fallback
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}
	return n
}

// Helper to parse setting as bool
func (db *DB) GetSettingBool(key string, fallback bool) bool {
	val, err := db.GetSetting(key)
	if err != nil || val == "" {
		return fallback
	}
	return val == "true" || val == "1"
}
