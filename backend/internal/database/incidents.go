package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

type Incident struct {
	ID              int64          `json:"id"`
	UserID          string         `json:"user_id"`
	Title           string         `json:"title"`
	Description     string         `json:"description"`
	SourceIP        string         `json:"source_ip"`
	StartTime       string         `json:"start_time"`
	EndTime         string         `json:"end_time"`
	EventCount      int            `json:"event_count"`
	UniqueUsers     []string       `json:"unique_users"`
	UniqueIPs       []string       `json:"unique_ips"`
	UniqueHostnames []string       `json:"unique_hostnames"`
	Severity        string         `json:"severity"`
	Status          string         `json:"status"`
	Techniques      []TechniqueRef `json:"techniques"`
	Tactics         []string       `json:"tactics"`
	MitreAttackIDs  []string       `json:"mitre_attack_ids"`
	Confidence      float64        `json:"confidence"`
	RawSummary      string         `json:"raw_summary"`
	CreatedAt       string         `json:"created_at"`
	UpdatedAt       string         `json:"updated_at"`
}

type TechniqueRef struct {
	TechniqueID string `json:"technique_id"`
	Name        string `json:"name"`
	Tactic      string `json:"tactic"`
	EventCount  int    `json:"event_count"`
}

func (db *DB) CreateIncident(inc *Incident, eventIDs []int64) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	usersJSON, _ := json.Marshal(inc.UniqueUsers)
	ipsJSON, _ := json.Marshal(inc.UniqueIPs)
	hostnamesJSON, _ := json.Marshal(inc.UniqueHostnames)
	techniquesJSON, _ := json.Marshal(inc.Techniques)
	tacticsJSON, _ := json.Marshal(inc.Tactics)
	mitreJSON, _ := json.Marshal(inc.MitreAttackIDs)

	result, err := tx.Exec(`
		INSERT INTO incidents (user_id, title, description, source_ip, start_time, end_time, event_count,
			unique_users, unique_ips, unique_hostnames, severity, status, techniques, tactics,
			mitre_attack_ids, confidence, raw_summary)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		inc.UserID, inc.Title, inc.Description, inc.SourceIP, inc.StartTime, inc.EndTime, inc.EventCount,
		string(usersJSON), string(ipsJSON), string(hostnamesJSON), inc.Severity, inc.Status,
		string(techniquesJSON), string(tacticsJSON), string(mitreJSON), inc.Confidence, inc.RawSummary)
	if err != nil {
		return fmt.Errorf("insert incident: %w", err)
	}

	incidentID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get incident id: %w", err)
	}

	for _, eventID := range eventIDs {
		_, err := tx.Exec(`
			INSERT OR IGNORE INTO incident_events (incident_id, event_id, timestamp, source_ip)
			SELECT ?, id, timestamp, source_ip FROM events WHERE id = ?`,
			incidentID, eventID)
		if err != nil {
			return fmt.Errorf("insert incident event: %w", err)
		}
	}

	for _, tech := range inc.Techniques {
		_, err := tx.Exec(`
			INSERT OR IGNORE INTO incident_techniques (incident_id, technique_id, event_count)
			VALUES (?, ?, ?)`,
			incidentID, tech.TechniqueID, tech.EventCount)
		if err != nil {
			return fmt.Errorf("insert incident technique: %w", err)
		}
	}

	inc.ID = incidentID
	return tx.Commit()
}

func (db *DB) GetIncidents(limit, offset int, severity, status, sourceIP string) ([]Incident, int, error) {
	where := []string{}
	args := []interface{}{}

	if severity != "" {
		where = append(where, "severity = ?")
		args = append(args, severity)
	}
	if status != "" {
		where = append(where, "status = ?")
		args = append(args, status)
	}
	if sourceIP != "" {
		where = append(where, "source_ip = ?")
		args = append(args, sourceIP)
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM incidents %s", whereClause)
	err := db.conn.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count incidents: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT id, user_id, title, description, source_ip, start_time, end_time, event_count,
			unique_users, unique_ips, unique_hostnames, severity, status, techniques,
			tactics, mitre_attack_ids, confidence, raw_summary, created_at, updated_at
		FROM incidents %s
		ORDER BY start_time DESC
		LIMIT ? OFFSET ?`, whereClause)

	args = append(args, limit, offset)
	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query incidents: %w", err)
	}
	defer rows.Close()

	var incidents []Incident
	for rows.Next() {
		var inc Incident
		var usersJSON, ipsJSON, hostnamesJSON, techniquesJSON, tacticsJSON, mitreJSON string
		err := rows.Scan(&inc.ID, &inc.UserID, &inc.Title, &inc.Description, &inc.SourceIP, &inc.StartTime,
			&inc.EndTime, &inc.EventCount, &usersJSON, &ipsJSON, &hostnamesJSON, &inc.Severity,
			&inc.Status, &techniquesJSON, &tacticsJSON, &mitreJSON, &inc.Confidence,
			&inc.RawSummary, &inc.CreatedAt, &inc.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("scan incident: %w", err)
		}
		json.Unmarshal([]byte(usersJSON), &inc.UniqueUsers)
		json.Unmarshal([]byte(ipsJSON), &inc.UniqueIPs)
		json.Unmarshal([]byte(hostnamesJSON), &inc.UniqueHostnames)
		json.Unmarshal([]byte(techniquesJSON), &inc.Techniques)
		json.Unmarshal([]byte(tacticsJSON), &inc.Tactics)
		json.Unmarshal([]byte(mitreJSON), &inc.MitreAttackIDs)
		incidents = append(incidents, inc)
	}

	return incidents, total, nil
}

func (db *DB) GetIncidentsByUserID(userID string, limit, offset int, severity, status, sourceIP string) ([]Incident, int, error) {
	where := []string{"user_id = ?"}
	args := []interface{}{userID}

	if severity != "" {
		where = append(where, "severity = ?")
		args = append(args, severity)
	}
	if status != "" {
		where = append(where, "status = ?")
		args = append(args, status)
	}
	if sourceIP != "" {
		where = append(where, "source_ip = ?")
		args = append(args, sourceIP)
	}

	whereClause := "WHERE " + strings.Join(where, " AND ")

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM incidents %s", whereClause)
	err := db.conn.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count incidents: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT id, user_id, title, description, source_ip, start_time, end_time, event_count,
			unique_users, unique_ips, unique_hostnames, severity, status, techniques,
			tactics, mitre_attack_ids, confidence, raw_summary, created_at, updated_at
		FROM incidents %s
		ORDER BY start_time DESC
		LIMIT ? OFFSET ?`, whereClause)

	args = append(args, limit, offset)
	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query incidents: %w", err)
	}
	defer rows.Close()

	var incidents []Incident
	for rows.Next() {
		var inc Incident
		var usersJSON, ipsJSON, hostnamesJSON, techniquesJSON, tacticsJSON, mitreJSON string
		err := rows.Scan(&inc.ID, &inc.UserID, &inc.Title, &inc.Description, &inc.SourceIP, &inc.StartTime,
			&inc.EndTime, &inc.EventCount, &usersJSON, &ipsJSON, &hostnamesJSON, &inc.Severity,
			&inc.Status, &techniquesJSON, &tacticsJSON, &mitreJSON, &inc.Confidence,
			&inc.RawSummary, &inc.CreatedAt, &inc.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("scan incident: %w", err)
		}
		json.Unmarshal([]byte(usersJSON), &inc.UniqueUsers)
		json.Unmarshal([]byte(ipsJSON), &inc.UniqueIPs)
		json.Unmarshal([]byte(hostnamesJSON), &inc.UniqueHostnames)
		json.Unmarshal([]byte(techniquesJSON), &inc.Techniques)
		json.Unmarshal([]byte(tacticsJSON), &inc.Tactics)
		json.Unmarshal([]byte(mitreJSON), &inc.MitreAttackIDs)
		incidents = append(incidents, inc)
	}

	return incidents, total, nil
}

func (db *DB) GetIncidentByID(id int64) (*Incident, error) {
	var inc Incident
	var usersJSON, ipsJSON, hostnamesJSON, techniquesJSON, tacticsJSON, mitreJSON string

	err := db.conn.QueryRow(`
		SELECT id, user_id, title, description, source_ip, start_time, end_time, event_count,
			unique_users, unique_ips, unique_hostnames, severity, status, techniques,
			tactics, mitre_attack_ids, confidence, raw_summary, created_at, updated_at
		FROM incidents WHERE id = ?`, id).Scan(&inc.ID, &inc.UserID, &inc.Title, &inc.Description,
		&inc.SourceIP, &inc.StartTime, &inc.EndTime, &inc.EventCount, &usersJSON,
		&ipsJSON, &hostnamesJSON, &inc.Severity, &inc.Status, &techniquesJSON,
		&tacticsJSON, &mitreJSON, &inc.Confidence, &inc.RawSummary, &inc.CreatedAt, &inc.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get incident: %w", err)
	}

	json.Unmarshal([]byte(usersJSON), &inc.UniqueUsers)
	json.Unmarshal([]byte(ipsJSON), &inc.UniqueIPs)
	json.Unmarshal([]byte(hostnamesJSON), &inc.UniqueHostnames)
	json.Unmarshal([]byte(techniquesJSON), &inc.Techniques)
	json.Unmarshal([]byte(tacticsJSON), &inc.Tactics)
	json.Unmarshal([]byte(mitreJSON), &inc.MitreAttackIDs)

	return &inc, nil
}

func (db *DB) GetIncidentEvents(incidentID int64) ([]Event, error) {
	rows, err := db.conn.Query(`
		SELECT e.id, e.user_id, e.timestamp, e.hostname, e.event_type, e.event_id, e.user_name,
			e.source_ip, e.dest_ip, e.process_name, e.command_line, e.parent_process,
			e.log_type, e.session_id, e.department, e.location, e.device_type, e.success,
			e.port, e.protocol, e.file_path, e.severity, e.error, e.raw_json, e.created_at
		FROM events e
		INNER JOIN incident_events ie ON e.id = ie.event_id
		WHERE ie.incident_id = ?
		ORDER BY e.timestamp ASC`, incidentID)
	if err != nil {
		return nil, fmt.Errorf("query incident events: %w", err)
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		var rawJSON string
		err := rows.Scan(&e.ID, &e.UserID, &e.Timestamp, &e.Hostname, &e.EventType, &e.EventID,
			&e.UserName, &e.SourceIP, &e.DestIP, &e.ProcessName, &e.CommandLine,
			&e.ParentProcess, &e.LogType, &e.SessionID, &e.Department, &e.Location,
			&e.DeviceType, &e.Success, &e.Port, &e.Protocol, &e.FilePath, &e.Severity,
			&e.Error, &rawJSON, &e.CreatedAt)
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

func (db *DB) GetIncidentStats() (map[string]interface{}, error) {
	stats := map[string]interface{}{}

	var total int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM incidents").Scan(&total)
	if err != nil {
		return nil, err
	}
	stats["total_incidents"] = total

	rows, err := db.conn.Query("SELECT severity, COUNT(*) FROM incidents GROUP BY severity")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	bySeverity := map[string]int{}
	for rows.Next() {
		var sev string
		var count int
		rows.Scan(&sev, &count)
		bySeverity[sev] = count
	}
	stats["by_severity"] = bySeverity

	var avgEvents float64
	db.conn.QueryRow("SELECT COALESCE(AVG(event_count), 0) FROM incidents").Scan(&avgEvents)
	stats["avg_events_per_incident"] = avgEvents

	return stats, nil
}

func (db *DB) GetIncidentStatsByUserID(userID string) (map[string]interface{}, error) {
	stats := map[string]interface{}{}

	var total int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM incidents WHERE user_id = ?", userID).Scan(&total)
	if err != nil {
		return nil, err
	}
	stats["total_incidents"] = total

	rows, err := db.conn.Query("SELECT severity, COUNT(*) FROM incidents WHERE user_id = ? GROUP BY severity", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	bySeverity := map[string]int{}
	for rows.Next() {
		var sev string
		var count int
		rows.Scan(&sev, &count)
		bySeverity[sev] = count
	}
	stats["by_severity"] = bySeverity

	var avgEvents float64
	db.conn.QueryRow("SELECT COALESCE(AVG(event_count), 0) FROM incidents WHERE user_id = ?", userID).Scan(&avgEvents)
	stats["avg_events_per_incident"] = avgEvents

	return stats, nil
}

func (db *DB) GetUnprocessedEvents() ([]Event, error) {
	rows, err := db.conn.Query(`
		SELECT e.id, e.user_id, e.timestamp, e.hostname, e.event_type, e.event_id, e.user_name,
			e.source_ip, e.dest_ip, e.process_name, e.command_line, e.parent_process,
			e.log_type, e.session_id, e.department, e.location, e.device_type, e.success,
			e.port, e.protocol, e.file_path, e.severity, e.error, e.raw_json, e.created_at
		FROM events e
		LEFT JOIN incident_events ie ON e.id = ie.event_id
		WHERE ie.id IS NULL AND e.source_ip IS NOT NULL AND e.source_ip != ''
		ORDER BY e.source_ip ASC, e.timestamp ASC`)
	if err != nil {
		return nil, fmt.Errorf("query unprocessed events: %w", err)
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		var rawJSON string
		err := rows.Scan(&e.ID, &e.UserID, &e.Timestamp, &e.Hostname, &e.EventType, &e.EventID,
			&e.UserName, &e.SourceIP, &e.DestIP, &e.ProcessName, &e.CommandLine,
			&e.ParentProcess, &e.LogType, &e.SessionID, &e.Department, &e.Location,
			&e.DeviceType, &e.Success, &e.Port, &e.Protocol, &e.FilePath, &e.Severity,
			&e.Error, &rawJSON, &e.CreatedAt)
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

func (db *DB) SeedTechniques(techniques []TechniqueRef) error {
	for _, t := range techniques {
		_, err := db.conn.Exec(`
			INSERT OR IGNORE INTO techniques (technique_id, name, tactic)
			VALUES (?, ?, ?)`, t.TechniqueID, t.Name, t.Tactic)
		if err != nil {
			return fmt.Errorf("seed technique: %w", err)
		}
	}
	return nil
}
