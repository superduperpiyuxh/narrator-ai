package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type Event struct {
	ID            int64                  `json:"id"`
	Timestamp     string                 `json:"timestamp"`
	Hostname      string                 `json:"hostname"`
	EventType     string                 `json:"event_type"`
	EventID       string                 `json:"event_id"`
	UserName      string                 `json:"user_name"`
	SourceIP      string                 `json:"source_ip"`
	DestIP        string                 `json:"dest_ip"`
	ProcessName   string                 `json:"process_name"`
	CommandLine   string                 `json:"command_line"`
	ParentProcess string                 `json:"parent_process"`
	LogType       string                 `json:"log_type"`
	SessionID     string                 `json:"session_id"`
	Department    string                 `json:"department"`
	Location      string                 `json:"location"`
	DeviceType    string                 `json:"device_type"`
	Success       bool                   `json:"success"`
	Port          string                 `json:"port"`
	Protocol      string                 `json:"protocol"`
	FilePath      string                 `json:"file_path"`
	Severity      string                 `json:"severity"`
	Error         string                 `json:"error"`
	RawJSON       map[string]interface{} `json:"raw_json"`
	CreatedAt     string                 `json:"created_at"`
}

func (db *DB) InsertEvents(events []Event) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO events 
		(timestamp, hostname, event_type, event_id, user_name, source_ip, dest_ip,
		 process_name, command_line, parent_process, log_type, session_id,
		 department, location, device_type, success, port, protocol, file_path,
		 severity, error, raw_json)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, e := range events {
		rawJSON, _ := json.Marshal(e.RawJSON)
		_, err := stmt.Exec(
			e.Timestamp, e.Hostname, e.EventType, e.EventID, e.UserName,
			e.SourceIP, e.DestIP, e.ProcessName, e.CommandLine, e.ParentProcess,
			e.LogType, e.SessionID, e.Department, e.Location, e.DeviceType,
			e.Success, e.Port, e.Protocol, e.FilePath, e.Severity, e.Error,
			string(rawJSON),
		)
		if err != nil {
			return fmt.Errorf("insert event: %w", err)
		}
	}

	return tx.Commit()
}

func (db *DB) GetEvents(limit, offset int) ([]Event, int, error) {
	var total int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM events").Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count events: %w", err)
	}

	rows, err := db.conn.Query(`
		SELECT id, timestamp, hostname, event_type, event_id, user_name,
		       source_ip, dest_ip, process_name, command_line, parent_process,
		       log_type, session_id, department, location, device_type,
		       success, port, protocol, file_path, severity, error, raw_json, created_at
		FROM events ORDER BY timestamp DESC LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query events: %w", err)
	}
	defer rows.Close()

	events, err := scanEvents(rows)
	return events, total, err
}

func (db *DB) GetEventsByHost(hostname string) ([]Event, error) {
	rows, err := db.conn.Query(`
		SELECT id, timestamp, hostname, event_type, event_id, user_name,
		       source_ip, dest_ip, process_name, command_line, parent_process,
		       log_type, session_id, department, location, device_type,
		       success, port, protocol, file_path, severity, error, raw_json, created_at
		FROM events WHERE hostname = ? ORDER BY timestamp DESC
	`, hostname)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEvents(rows)
}

func (db *DB) GetEventsByType(eventType string) ([]Event, error) {
	rows, err := db.conn.Query(`
		SELECT id, timestamp, hostname, event_type, event_id, user_name,
		       source_ip, dest_ip, process_name, command_line, parent_process,
		       log_type, session_id, department, location, device_type,
		       success, port, protocol, file_path, severity, error, raw_json, created_at
		FROM events WHERE event_type = ? ORDER BY timestamp DESC
	`, eventType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEvents(rows)
}

func (db *DB) SearchEvents(query string) ([]Event, error) {
	like := "%" + query + "%"
	rows, err := db.conn.Query(`
		SELECT id, timestamp, hostname, event_type, event_id, user_name,
		       source_ip, dest_ip, process_name, command_line, parent_process,
		       log_type, session_id, department, location, device_type,
		       success, port, protocol, file_path, severity, error, raw_json, created_at
		FROM events 
		WHERE hostname LIKE ? OR event_type LIKE ? OR user_name LIKE ? 
		   OR source_ip LIKE ? OR command_line LIKE ?
		ORDER BY timestamp DESC LIMIT 100
	`, like, like, like, like, like)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEvents(rows)
}

func (db *DB) GetStats() (map[string]interface{}, error) {
	stats := map[string]interface{}{}

	var count int
	db.conn.QueryRow("SELECT COUNT(*) FROM events").Scan(&count)
	stats["total_events"] = count

	var hosts int
	db.conn.QueryRow("SELECT COUNT(DISTINCT hostname) FROM events").Scan(&hosts)
	stats["unique_hosts"] = hosts

	var users int
	db.conn.QueryRow("SELECT COUNT(DISTINCT user_name) FROM events").Scan(&users)
	stats["unique_users"] = users

	var ipCount int
	db.conn.QueryRow("SELECT COUNT(DISTINCT source_ip) FROM events").Scan(&ipCount)
	stats["unique_ips"] = ipCount

	rows, err := db.conn.Query("SELECT event_type, COUNT(*) FROM events GROUP BY event_type ORDER BY COUNT(*) DESC LIMIT 10")
	if err == nil {
		defer rows.Close()
		eventTypes := map[string]int{}
		for rows.Next() {
			var et string
			var c int
			rows.Scan(&et, &c)
			eventTypes[et] = c
		}
		stats["event_types"] = eventTypes
	}

	var minTS, maxTS string
	db.conn.QueryRow("SELECT MIN(timestamp), MAX(timestamp) FROM events").Scan(&minTS, &maxTS)
	stats["time_range"] = map[string]string{"start": minTS, "end": maxTS}

	return stats, nil
}

func scanEvents(rows *sql.Rows) ([]Event, error) {
	var events []Event
	for rows.Next() {
		var e Event
		var rawJSON string
		err := rows.Scan(
			&e.ID, &e.Timestamp, &e.Hostname, &e.EventType, &e.EventID,
			&e.UserName, &e.SourceIP, &e.DestIP, &e.ProcessName, &e.CommandLine,
			&e.ParentProcess, &e.LogType, &e.SessionID, &e.Department, &e.Location,
			&e.DeviceType, &e.Success, &e.Port, &e.Protocol, &e.FilePath,
			&e.Severity, &e.Error, &rawJSON, &e.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		if rawJSON != "" {
			json.Unmarshal([]byte(rawJSON), &e.RawJSON)
		}
		events = append(events, e)
	}
	return events, nil
}

func init() {
	_ = time.Now()
}
