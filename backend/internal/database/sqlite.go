package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	conn *sql.DB
}

func New(dbPath string) (*DB, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	conn, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000&_synchronous=NORMAL")
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return db, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp TEXT NOT NULL,
		hostname TEXT NOT NULL,
		event_type TEXT NOT NULL,
		event_id TEXT,
		user_name TEXT,
		source_ip TEXT,
		dest_ip TEXT,
		process_name TEXT,
		command_line TEXT,
		parent_process TEXT,
		log_type TEXT,
		session_id TEXT,
		department TEXT,
		location TEXT,
		device_type TEXT,
		success BOOLEAN DEFAULT 0,
		port TEXT,
		protocol TEXT,
		file_path TEXT,
		severity TEXT,
		error TEXT,
		raw_json TEXT,
		created_at TEXT DEFAULT (datetime('now')),
		UNIQUE(hostname, timestamp, event_type)
	);

	CREATE INDEX IF NOT EXISTS idx_events_timestamp ON events(timestamp);
	CREATE INDEX IF NOT EXISTS idx_events_hostname ON events(hostname);
	CREATE INDEX IF NOT EXISTS idx_events_event_type ON events(event_type);
	CREATE INDEX IF NOT EXISTS idx_events_source_ip ON events(source_ip);
	CREATE INDEX IF NOT EXISTS idx_events_user ON events(user_name);
	CREATE INDEX IF NOT EXISTS idx_events_success ON events(success);
	`

	_, err := db.conn.Exec(schema)
	return err
}
