package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/superduperpiyuxh/narrator-ai/backend/internal/database"
)

func setupFeedbackHandlerTestDB(t *testing.T) *database.DB {
	t.Helper()
	dbPath := t.TempDir() + "/test.db"
	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	return db
}

func TestSubmitFeedback_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupFeedbackHandlerTestDB(t)
	defer db.Close()

	inc := &database.Incident{Title: "Test", SourceIP: "10.0.0.1", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", Severity: "medium", Status: "new"}
	db.CreateIncident(inc, nil)
	narr := &database.Narrative{IncidentID: inc.ID, Summary: "Test", Confidence: 0.8, Sentences: "[]", ModelUsed: "test", Temperature: 0.2}
	db.CreateNarrative(narr)

	h := NewFeedbackHandler(db)
	router := gin.New()
	router.POST("/api/feedback", h.SubmitFeedback)

	body, _ := json.Marshal(map[string]interface{}{
		"narrative_id": narr.ID,
		"incident_id":  inc.ID,
		"rating":       1,
		"notes":        "Great narrative!",
	})
	req := httptest.NewRequest("POST", "/api/feedback", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestSubmitFeedback_InvalidRating(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupFeedbackHandlerTestDB(t)
	defer db.Close()

	h := NewFeedbackHandler(db)
	router := gin.New()
	router.POST("/api/feedback", h.SubmitFeedback)

	body, _ := json.Marshal(map[string]interface{}{
		"narrative_id": 1,
		"incident_id":  1,
		"rating":       0, // Invalid — must be -1 or 1
		"notes":        "test",
	})
	req := httptest.NewRequest("POST", "/api/feedback", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestSubmitFeedback_MissingNarrativeID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupFeedbackHandlerTestDB(t)
	defer db.Close()

	h := NewFeedbackHandler(db)
	router := gin.New()
	router.POST("/api/feedback", h.SubmitFeedback)

	body, _ := json.Marshal(map[string]interface{}{
		"narrative_id": 0,
		"incident_id":  1,
		"rating":       1,
	})
	req := httptest.NewRequest("POST", "/api/feedback", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestSubmitFeedback_NotesTruncated(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupFeedbackHandlerTestDB(t)
	defer db.Close()

	inc := &database.Incident{Title: "Test", SourceIP: "10.0.0.1", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", Severity: "medium", Status: "new"}
	db.CreateIncident(inc, nil)
	narr := &database.Narrative{IncidentID: inc.ID, Summary: "Test", Confidence: 0.8, Sentences: "[]", ModelUsed: "test", Temperature: 0.2}
	db.CreateNarrative(narr)

	h := NewFeedbackHandler(db)
	router := gin.New()
	router.POST("/api/feedback", h.SubmitFeedback)

	longNotes := ""
	for i := 0; i < 1100; i++ {
		longNotes += "a"
	}
	body, _ := json.Marshal(map[string]interface{}{
		"narrative_id": narr.ID,
		"incident_id":  inc.ID,
		"rating":       1,
		"notes":        longNotes,
	})
	req := httptest.NewRequest("POST", "/api/feedback", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	// Verify notes were truncated
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	fb := resp["feedback"].(map[string]interface{})
	notes := fb["notes"].(string)
	if len(notes) > 1000 {
		t.Errorf("expected notes <= 1000 chars, got %d", len(notes))
	}
}

func TestGetFeedback(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupFeedbackHandlerTestDB(t)
	defer db.Close()

	inc := &database.Incident{Title: "Test", SourceIP: "10.0.0.1", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", Severity: "medium", Status: "new"}
	db.CreateIncident(inc, nil)
	narr := &database.Narrative{IncidentID: inc.ID, Summary: "Test", Confidence: 0.8, Sentences: "[]", ModelUsed: "test", Temperature: 0.2}
	db.CreateNarrative(narr)
	fb := &database.Feedback{NarrativeID: narr.ID, IncidentID: inc.ID, Rating: 1, Notes: "Good"}
	db.CreateFeedback(fb)

	h := NewFeedbackHandler(db)
	router := gin.New()
	router.GET("/api/feedback/:narrative_id", h.GetFeedback)

	req := httptest.NewRequest("GET", "/api/feedback/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetFeedback_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupFeedbackHandlerTestDB(t)
	defer db.Close()

	h := NewFeedbackHandler(db)
	router := gin.New()
	router.GET("/api/feedback/:narrative_id", h.GetFeedback)

	req := httptest.NewRequest("GET", "/api/feedback/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 with null feedback, got %d", w.Code)
	}
}

func TestGetFeedback_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupFeedbackHandlerTestDB(t)
	defer db.Close()

	h := NewFeedbackHandler(db)
	router := gin.New()
	router.GET("/api/feedback/:narrative_id", h.GetFeedback)

	req := httptest.NewRequest("GET", "/api/feedback/abc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestSubmitFeedback_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupFeedbackHandlerTestDB(t)
	defer db.Close()

	h := NewFeedbackHandler(db)
	router := gin.New()
	router.POST("/api/feedback", h.SubmitFeedback)

	req := httptest.NewRequest("POST", "/api/feedback", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}
