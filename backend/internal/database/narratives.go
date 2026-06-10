package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

type Narrative struct {
	ID               int64   `json:"id"`
	IncidentID       int64   `json:"incident_id"`
	UserID           string  `json:"user_id"`
	Summary          string  `json:"summary"`
	Confidence       float64 `json:"confidence"`
	Sentences        string  `json:"sentences"`
	ModelUsed        string  `json:"model_used"`
	Temperature      float64 `json:"temperature"`
	TokensUsed       int     `json:"tokens_used"`
	GenerationTimeMs int64   `json:"generation_time_ms"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

func (db *DB) CreateNarrative(n *Narrative) error {
	result, err := db.conn.Exec(`
		INSERT INTO narratives (incident_id, user_id, summary, confidence, sentences, model_used, temperature, tokens_used, generation_time_ms)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		n.IncidentID, n.UserID, n.Summary, n.Confidence, n.Sentences, n.ModelUsed, n.Temperature, n.TokensUsed, n.GenerationTimeMs)
	if err != nil {
		return fmt.Errorf("insert narrative: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get narrative id: %w", err)
	}

	n.ID = id
	return nil
}

func (db *DB) GetNarrativeByIncidentID(incidentID int64) (*Narrative, error) {
	var n Narrative
	err := db.conn.QueryRow(`
		SELECT id, incident_id, user_id, summary, confidence, sentences, model_used, temperature, tokens_used, generation_time_ms, created_at, updated_at
		FROM narratives WHERE incident_id = ?`, incidentID).Scan(&n.ID, &n.IncidentID, &n.UserID, &n.Summary, &n.Confidence, &n.Sentences, &n.ModelUsed, &n.Temperature, &n.TokensUsed, &n.GenerationTimeMs, &n.CreatedAt, &n.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get narrative: %w", err)
	}
	return &n, nil
}

func (db *DB) GetNarrativeByID(id int64) (*Narrative, error) {
	var n Narrative
	err := db.conn.QueryRow(`
		SELECT id, incident_id, user_id, summary, confidence, sentences, model_used, temperature, tokens_used, generation_time_ms, created_at, updated_at
		FROM narratives WHERE id = ?`, id).Scan(&n.ID, &n.IncidentID, &n.UserID, &n.Summary, &n.Confidence, &n.Sentences, &n.ModelUsed, &n.Temperature, &n.TokensUsed, &n.GenerationTimeMs, &n.CreatedAt, &n.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get narrative: %w", err)
	}
	return &n, nil
}

func (db *DB) GetNarrativeSourceEvents(narrativeID int64) ([]Event, error) {
	var n Narrative
	err := db.conn.QueryRow("SELECT sentences FROM narratives WHERE id = ?", narrativeID).Scan(&n.Sentences)
	if err != nil {
		return nil, fmt.Errorf("get narrative sentences: %w", err)
	}

	var result struct {
		Sentences []struct {
			SourceEventIDs []int64 `json:"source_event_ids"`
		} `json:"sentences"`
	}
	if err := json.Unmarshal([]byte(n.Sentences), &result); err != nil {
		return nil, fmt.Errorf("unmarshal sentences: %w", err)
	}

	eventIDs := make(map[int64]bool)
	for _, s := range result.Sentences {
		for _, id := range s.SourceEventIDs {
			eventIDs[id] = true
		}
	}

	if len(eventIDs) == 0 {
		return nil, nil
	}

	ids := make([]interface{}, 0, len(eventIDs))
	placeholders := make([]string, 0, len(eventIDs))
	for id := range eventIDs {
		ids = append(ids, id)
		placeholders = append(placeholders, "?")
	}

	query := fmt.Sprintf(`
		SELECT id, user_id, timestamp, hostname, event_type, event_id, user_name, source_ip, dest_ip,
			process_name, command_line, parent_process, log_type, session_id, department,
			location, device_type, success, port, protocol, file_path, severity, error, raw_json, created_at
		FROM events WHERE id IN (%s)
		ORDER BY timestamp ASC`, strings.Join(placeholders, ","))

	rows, err := db.conn.Query(query, ids...)
	if err != nil {
		return nil, fmt.Errorf("query source events: %w", err)
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
