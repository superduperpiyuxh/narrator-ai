package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/superduperpiyuxh/narrator-ai/backend/internal/database"
)

func setupHandlerTestDB(t *testing.T) *database.DB {
	t.Helper()
	dbPath := t.TempDir() + "/test.db"
	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	return db
}

func TestHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupHandlerTestDB(t)
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
}

func TestGetEvents(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupHandlerTestDB(t)
	defer db.Close()

	events := []database.Event{
		{Timestamp: "2025-12-21T10:00:00Z", Hostname: "host1", EventType: "auth", SourceIP: "10.0.0.1"},
		{Timestamp: "2025-12-21T10:01:00Z", Hostname: "host2", EventType: "network", SourceIP: "10.0.0.2"},
	}
	db.InsertEvents(events)

	h := New(db, t.TempDir())
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "")
		c.Next()
	})
	router.GET("/api/events", h.GetEvents)

	req := httptest.NewRequest("GET", "/api/events?limit=10&offset=0", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["total"] != float64(2) {
		t.Errorf("expected 2 events, got %v", resp["total"])
	}
}

func TestGetEvents_InvalidParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupHandlerTestDB(t)
	defer db.Close()

	h := New(db, t.TempDir())
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "")
		c.Next()
	})
	router.GET("/api/events", h.GetEvents)

	req := httptest.NewRequest("GET", "/api/events?limit=-1&offset=-5", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 (defaults applied), got %d", w.Code)
	}
}

func TestSearchEvents(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupHandlerTestDB(t)
	defer db.Close()

	events := []database.Event{
		{Timestamp: "2025-12-21T10:00:00Z", Hostname: "dc01", EventType: "auth", SourceIP: "10.0.0.1", UserName: "admin"},
	}
	db.InsertEvents(events)

	h := New(db, t.TempDir())
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "")
		c.Next()
	})
	router.GET("/api/events/search", h.SearchEvents)

	req := httptest.NewRequest("GET", "/api/events/search?q=dc01", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestSearchEvents_MissingQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupHandlerTestDB(t)
	defer db.Close()

	h := New(db, t.TempDir())
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "")
		c.Next()
	})
	router.GET("/api/events/search", h.SearchEvents)

	req := httptest.NewRequest("GET", "/api/events/search", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestSearchEvents_QueryTooLong(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupHandlerTestDB(t)
	defer db.Close()

	h := New(db, t.TempDir())
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "")
		c.Next()
	})
	router.GET("/api/events/search", h.SearchEvents)

	longQuery := make([]byte, 201)
	for i := range longQuery {
		longQuery[i] = 'a'
	}
	req := httptest.NewRequest("GET", "/api/events/search?q="+string(longQuery), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for long query, got %d", w.Code)
	}
}

func TestGetEventsByHost(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupHandlerTestDB(t)
	defer db.Close()

	events := []database.Event{
		{Timestamp: "2025-12-21T10:00:00Z", Hostname: "dc01", EventType: "auth", SourceIP: "10.0.0.1"},
	}
	db.InsertEvents(events)

	h := New(db, t.TempDir())
	router := gin.New()
	router.GET("/api/events/host/:hostname", h.GetEventsByHost)

	req := httptest.NewRequest("GET", "/api/events/host/dc01", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetEventsByType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupHandlerTestDB(t)
	defer db.Close()

	events := []database.Event{
		{Timestamp: "2025-12-21T10:00:00Z", Hostname: "dc01", EventType: "authentication", SourceIP: "10.0.0.1"},
	}
	db.InsertEvents(events)

	h := New(db, t.TempDir())
	router := gin.New()
	router.GET("/api/events/type/:eventType", h.GetEventsByType)

	req := httptest.NewRequest("GET", "/api/events/type/authentication", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetStats(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupHandlerTestDB(t)
	defer db.Close()

	events := []database.Event{
		{Timestamp: "2025-12-21T10:00:00Z", Hostname: "host1", EventType: "auth", SourceIP: "10.0.0.1"},
	}
	db.InsertEvents(events)

	h := New(db, t.TempDir())
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "")
		c.Next()
	})
	router.GET("/api/stats", h.GetStats)

	req := httptest.NewRequest("GET", "/api/stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}
