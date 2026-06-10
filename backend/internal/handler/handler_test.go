package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/superduperpiyuxh/narrator-ai/backend/internal/database"
)

func setupTestDB(t *testing.T) *database.DB {
	dbPath := t.TempDir() + "/test.db"
	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	return db
}

func TestHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)
	defer db.Close()

	h := New(db, t.TempDir())
	router := gin.New()
	router.GET("/health", h.Health)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["status"] != "ok" {
		t.Errorf("expected status ok, got %s", resp["status"])
	}
}

func TestGetEvents(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)
	defer db.Close()

	events := []database.Event{
		{Timestamp: "2025-12-21T10:00:00Z", Hostname: "test", EventType: "auth", SourceIP: "10.0.0.1"},
	}
	db.InsertEvents(events)

	h := New(db, t.TempDir())
	router := gin.New()
	router.GET("/api/events", h.GetEvents)

	req := httptest.NewRequest("GET", "/api/events?limit=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["total"] != float64(1) {
		t.Errorf("expected 1 event, got %v", resp["total"])
	}
}

func TestGetStats(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)
	defer db.Close()

	events := []database.Event{
		{Timestamp: "2025-12-21T10:00:00Z", Hostname: "host1", EventType: "auth", SourceIP: "10.0.0.1"},
	}
	db.InsertEvents(events)

	h := New(db, t.TempDir())
	router := gin.New()
	router.GET("/api/stats", h.GetStats)

	req := httptest.NewRequest("GET", "/api/stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
