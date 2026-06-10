package database

import (
	"fmt"
	"time"
)

type Feedback struct {
	ID          int64  `json:"id"`
	NarrativeID int64  `json:"narrative_id"`
	IncidentID  int64  `json:"incident_id"`
	Rating      int    `json:"rating"` // -1 or 1
	Notes       string `json:"notes"`
	UserID      string `json:"user_id"`
	CreatedAt   string `json:"created_at"`
}

func (db *DB) CreateFeedback(f *Feedback) error {
	result, err := db.conn.Exec(`
		INSERT INTO feedback (narrative_id, incident_id, rating, notes, user_id)
		VALUES (?, ?, ?, ?, ?)`,
		f.NarrativeID, f.IncidentID, f.Rating, f.Notes, f.UserID)
	if err != nil {
		return fmt.Errorf("insert feedback: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get feedback id: %w", err)
	}

	f.ID = id
	if f.CreatedAt == "" {
		f.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	return nil
}

func (db *DB) GetFeedbackByNarrativeID(narrativeID int64) (*Feedback, error) {
	var f Feedback
	err := db.conn.QueryRow(`
		SELECT id, narrative_id, incident_id, rating, notes, user_id, created_at
		FROM feedback WHERE narrative_id = ?`, narrativeID).Scan(
		&f.ID, &f.NarrativeID, &f.IncidentID, &f.Rating, &f.Notes, &f.UserID, &f.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get feedback: %w", err)
	}
	return &f, nil
}

func (db *DB) GetFeedbackByIncidentID(incidentID int64) ([]Feedback, error) {
	rows, err := db.conn.Query(`
		SELECT id, narrative_id, incident_id, rating, notes, user_id, created_at
		FROM feedback WHERE incident_id = ?
		ORDER BY created_at DESC`, incidentID)
	if err != nil {
		return nil, fmt.Errorf("query feedback: %w", err)
	}
	defer rows.Close()

	var feedbacks []Feedback
	for rows.Next() {
		var f Feedback
		err := rows.Scan(&f.ID, &f.NarrativeID, &f.IncidentID, &f.Rating, &f.Notes, &f.UserID, &f.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan feedback: %w", err)
		}
		feedbacks = append(feedbacks, f)
	}
	return feedbacks, nil
}
