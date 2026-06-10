package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

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

	conn, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000&_synchronous=NORMAL&_foreign_keys=ON")
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	conn.SetMaxOpenConns(1)
	conn.SetMaxIdleConns(1)
	conn.SetConnMaxLifetime(0)

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
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		email TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		api_key TEXT NOT NULL UNIQUE,
		openrouter_key TEXT DEFAULT '',
		created_at TEXT DEFAULT (datetime('now')),
		updated_at TEXT DEFAULT (datetime('now'))
	);

	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	CREATE INDEX IF NOT EXISTS idx_users_api_key ON users(api_key);

	CREATE TABLE IF NOT EXISTS events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT DEFAULT '',
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

	CREATE INDEX IF NOT EXISTS idx_events_user_id ON events(user_id);
	CREATE INDEX IF NOT EXISTS idx_events_timestamp ON events(timestamp);
	CREATE INDEX IF NOT EXISTS idx_events_hostname ON events(hostname);
	CREATE INDEX IF NOT EXISTS idx_events_event_type ON events(event_type);
	CREATE INDEX IF NOT EXISTS idx_events_source_ip ON events(source_ip);
	CREATE INDEX IF NOT EXISTS idx_events_user ON events(user_name);
	CREATE INDEX IF NOT EXISTS idx_events_success ON events(success);

	CREATE TABLE IF NOT EXISTS incidents (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT DEFAULT '',
		title TEXT NOT NULL,
		description TEXT,
		source_ip TEXT NOT NULL,
		start_time TEXT NOT NULL,
		end_time TEXT NOT NULL,
		event_count INTEGER NOT NULL DEFAULT 0,
		unique_users TEXT,
		unique_ips TEXT,
		unique_hostnames TEXT,
		severity TEXT DEFAULT 'low',
		status TEXT DEFAULT 'new',
		techniques TEXT,
		tactics TEXT,
		mitre_attack_ids TEXT,
		confidence REAL DEFAULT 0.0,
		raw_summary TEXT,
		created_at TEXT DEFAULT (datetime('now')),
		updated_at TEXT DEFAULT (datetime('now'))
	);

	CREATE INDEX IF NOT EXISTS idx_incidents_user_id ON incidents(user_id);
	CREATE INDEX IF NOT EXISTS idx_incidents_source_ip ON incidents(source_ip);
	CREATE INDEX IF NOT EXISTS idx_incidents_start_time ON incidents(start_time);
	CREATE INDEX IF NOT EXISTS idx_incidents_severity ON incidents(severity);
	CREATE INDEX IF NOT EXISTS idx_incidents_status ON incidents(status);

	CREATE TABLE IF NOT EXISTS incident_events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		incident_id INTEGER NOT NULL,
		event_id INTEGER NOT NULL,
		timestamp TEXT NOT NULL,
		source_ip TEXT,
		FOREIGN KEY (incident_id) REFERENCES incidents(id) ON DELETE CASCADE,
		FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE,
		UNIQUE(incident_id, event_id)
	);

	CREATE TABLE IF NOT EXISTS techniques (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		technique_id TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL,
		description TEXT,
		tactic TEXT,
		url TEXT,
		created_at TEXT DEFAULT (datetime('now'))
	);

	CREATE TABLE IF NOT EXISTS incident_techniques (
		incident_id INTEGER NOT NULL,
		technique_id TEXT NOT NULL,
		event_count INTEGER DEFAULT 0,
		PRIMARY KEY (incident_id, technique_id),
		FOREIGN KEY (incident_id) REFERENCES incidents(id) ON DELETE CASCADE,
		FOREIGN KEY (technique_id) REFERENCES techniques(technique_id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_incident_events_incident ON incident_events(incident_id);
	CREATE INDEX IF NOT EXISTS idx_incident_events_event ON incident_events(event_id);

	CREATE TABLE IF NOT EXISTS narratives (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		incident_id INTEGER NOT NULL,
		user_id TEXT DEFAULT '',
		summary TEXT NOT NULL,
		confidence REAL NOT NULL DEFAULT 0.0,
		sentences TEXT NOT NULL,
		model_used TEXT NOT NULL,
		temperature REAL NOT NULL,
		tokens_used INTEGER DEFAULT 0,
		generation_time_ms INTEGER DEFAULT 0,
		created_at TEXT DEFAULT (datetime('now')),
		updated_at TEXT DEFAULT (datetime('now')),
		FOREIGN KEY (incident_id) REFERENCES incidents(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_narratives_incident ON narratives(incident_id);
	CREATE INDEX IF NOT EXISTS idx_narratives_user_id ON narratives(user_id);

	CREATE TABLE IF NOT EXISTS feedback (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		narrative_id INTEGER NOT NULL,
		incident_id INTEGER NOT NULL,
		rating INTEGER NOT NULL CHECK(rating IN (-1, 1)),
		notes TEXT,
		user_id TEXT,
		created_at TEXT DEFAULT (datetime('now')),
		FOREIGN KEY (narrative_id) REFERENCES narratives(id) ON DELETE CASCADE,
		FOREIGN KEY (incident_id) REFERENCES incidents(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_feedback_narrative ON feedback(narrative_id);
	CREATE INDEX IF NOT EXISTS idx_feedback_incident ON feedback(incident_id);
	`

	_, err := db.conn.Exec(schema)
	if err != nil {
		return err
	}

	// Create demo user for backward compatibility with existing data
	var count int
	db.conn.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if count == 0 {
		_, err = db.conn.Exec(`
			INSERT INTO users (id, email, password_hash, api_key, created_at, updated_at)
			VALUES ('demo', 'demo@narrator.ai', '', 'nai_demo_key_12345', datetime('now'), datetime('now'))`)
		if err != nil {
			return fmt.Errorf("create demo user: %w", err)
		}
		// Assign all existing data to demo user
		db.conn.Exec("UPDATE events SET user_id = 'demo' WHERE user_id = '' OR user_id IS NULL")
		db.conn.Exec("UPDATE incidents SET user_id = 'demo' WHERE user_id = '' OR user_id IS NULL")
		db.conn.Exec("UPDATE narratives SET user_id = 'demo' WHERE user_id = '' OR user_id IS NULL")
	}

	return nil
}

func (db *DB) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	return db.conn.PingContext(ctx)
}

func (db *DB) Conn() *sql.DB {
	return db.conn
}
