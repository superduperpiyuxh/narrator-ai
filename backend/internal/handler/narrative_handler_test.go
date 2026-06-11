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

func setupNarrativeHandlerTestDB(t *testing.T) *database.DB {
	t.Helper()
	dbPath := t.TempDir() + "/test.db"
	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	return db
}

func TestGetNarrative(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupNarrativeHandlerTestDB(t)
	defer db.Close()

	inc := &database.Incident{Title: "Test", SourceIP: "10.0.0.1", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", Severity: "medium", Status: "new"}
	db.CreateIncident(inc, nil)
	narr := &database.Narrative{IncidentID: inc.ID, Summary: "Test narrative", Confidence: 0.8, Sentences: "[]", ModelUsed: "test", Temperature: 0.2}
	db.CreateNarrative(narr)

	h := NewNarrativeHandler(db, nil, "")
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "")
		c.Next()
	})
	router.GET("/api/incidents/:id/narrative", h.GetNarrative)

	req := httptest.NewRequest("GET", "/api/incidents/1/narrative", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetNarrative_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupNarrativeHandlerTestDB(t)
	defer db.Close()

	h := NewNarrativeHandler(db, nil, "")
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "")
		c.Next()
	})
	router.GET("/api/incidents/:id/narrative", h.GetNarrative)

	req := httptest.NewRequest("GET", "/api/incidents/abc/narrative", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetNarrative_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupNarrativeHandlerTestDB(t)
	defer db.Close()

	h := NewNarrativeHandler(db, nil, "")
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "")
		c.Next()
	})
	router.GET("/api/incidents/:id/narrative", h.GetNarrative)

	req := httptest.NewRequest("GET", "/api/incidents/999/narrative", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestGetNarrativeSourceEvents(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupNarrativeHandlerTestDB(t)
	defer db.Close()

	events := []database.Event{
		{Timestamp: "2025-12-21T10:00:00Z", Hostname: "dc01", EventType: "auth", SourceIP: "10.0.0.1"},
	}
	db.InsertEvents(events)

	inc := &database.Incident{Title: "Test", SourceIP: "10.0.0.1", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", Severity: "medium", Status: "new"}
	db.CreateIncident(inc, nil)
	narr := &database.Narrative{IncidentID: inc.ID, Summary: "Test", Confidence: 0.8, Sentences: `{"sentences":[{"text":"Test","timestamp":"2025-12-21T10:00:00Z","source_event_ids":[1],"confidence":0.9}]}`, ModelUsed: "test", Temperature: 0.2}
	db.CreateNarrative(narr)

	h := NewNarrativeHandler(db, nil, "")
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "")
		c.Next()
	})
	router.GET("/api/narratives/:id/source-events", h.GetNarrativeSourceEvents)

	req := httptest.NewRequest("GET", "/api/narratives/1/source-events", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetNarrativeSourceEvents_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupNarrativeHandlerTestDB(t)
	defer db.Close()

	h := NewNarrativeHandler(db, nil, "")
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "")
		c.Next()
	})
	router.GET("/api/narratives/:id/source-events", h.GetNarrativeSourceEvents)

	req := httptest.NewRequest("GET", "/api/narratives/abc/source-events", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetNarrativeSourceEvents_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupNarrativeHandlerTestDB(t)
	defer db.Close()

	h := NewNarrativeHandler(db, nil, "")
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "")
		c.Next()
	})
	router.GET("/api/narratives/:id/source-events", h.GetNarrativeSourceEvents)

	req := httptest.NewRequest("GET", "/api/narratives/999/source-events", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestGenerateNarrative_NoLLMKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupNarrativeHandlerTestDB(t)
	defer db.Close()

	inc := &database.Incident{Title: "Test", SourceIP: "10.0.0.1", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", Severity: "medium", Status: "new"}
	db.CreateIncident(inc, nil)

	h := NewNarrativeHandler(db, nil, "")
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "")
		c.Next()
	})
	router.POST("/api/incidents/:id/narrative", h.GenerateNarrative)

	body, _ := json.Marshal(map[string]string{})
	req := httptest.NewRequest("POST", "/api/incidents/1/narrative", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503 (no LLM key), got %d: %s", w.Code, w.Body.String())
	}
}

func TestGenerateNarrative_IncidentNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupNarrativeHandlerTestDB(t)
	defer db.Close()

	h := NewNarrativeHandler(db, nil, "fake-key")
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "")
		c.Next()
	})
	router.POST("/api/incidents/:id/narrative", h.GenerateNarrative)

	body, _ := json.Marshal(map[string]string{})
	req := httptest.NewRequest("POST", "/api/incidents/999/narrative", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGenerateNarrative_CachedNarrative(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupNarrativeHandlerTestDB(t)
	defer db.Close()

	inc := &database.Incident{Title: "Test", SourceIP: "10.0.0.1", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", Severity: "medium", Status: "new"}
	db.CreateIncident(inc, nil)
	narr := &database.Narrative{IncidentID: inc.ID, Summary: "Cached", Confidence: 0.8, Sentences: "[]", ModelUsed: "test", Temperature: 0.2}
	db.CreateNarrative(narr)

	h := NewNarrativeHandler(db, nil, "fake-key")
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "")
		c.Next()
	})
	router.POST("/api/incidents/:id/narrative", h.GenerateNarrative)

	body, _ := json.Marshal(map[string]string{})
	req := httptest.NewRequest("POST", "/api/incidents/1/narrative", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 (cached), got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["cached"] != true {
		t.Error("expected cached=true")
	}
}
