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

func setupStreamTestDB(t *testing.T) *database.DB {
	t.Helper()
	dbPath := t.TempDir() + "/test.db"
	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	return db
}

func TestStreamHandler_IngestStream_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupStreamTestDB(t)
	defer db.Close()

	h := NewStreamHandler(db)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "stream-user")
		c.Next()
	})
	router.POST("/api/v1/stream", h.IngestStream)

	events := []map[string]interface{}{
		{
			"timestamp":  "2025-12-21T10:00:00Z",
			"hostname":   "dc01",
			"event_type": "authentication",
			"source_ip":  "10.0.0.1",
			"user_name":  "admin",
		},
		{
			"timestamp":  "2025-12-21T10:01:00Z",
			"hostname":   "dc01",
			"event_type": "process_activity",
			"source_ip":  "10.0.0.1",
			"user_name":  "admin",
		},
	}
	body, _ := json.Marshal(map[string]interface{}{"events": events})

	req := httptest.NewRequest("POST", "/api/v1/stream", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["imported"] != float64(2) {
		t.Errorf("expected 2 events imported, got %v", resp["imported"])
	}
}

func TestStreamHandler_IngestStream_EmptyEvents(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupStreamTestDB(t)
	defer db.Close()

	h := NewStreamHandler(db)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "stream-user")
		c.Next()
	})
	router.POST("/api/v1/stream", h.IngestStream)

	body, _ := json.Marshal(map[string]interface{}{"events": []interface{}{}})
	req := httptest.NewRequest("POST", "/api/v1/stream", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty events, got %d", w.Code)
	}
}

func TestStreamHandler_IngestStream_MissingFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupStreamTestDB(t)
	defer db.Close()

	h := NewStreamHandler(db)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "stream-user")
		c.Next()
	})
	router.POST("/api/v1/stream", h.IngestStream)

	events := []map[string]interface{}{
		{"timestamp": "2025-12-21T10:00:00Z"},
	}
	body, _ := json.Marshal(map[string]interface{}{"events": events})

	req := httptest.NewRequest("POST", "/api/v1/stream", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for no valid events, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "no valid events" {
		t.Errorf("expected 'no valid events' error, got %v", resp["error"])
	}
}

func TestStreamHandler_IngestStream_NoUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupStreamTestDB(t)
	defer db.Close()

	h := NewStreamHandler(db)
	router := gin.New()
	router.POST("/api/v1/stream", h.IngestStream)

	body, _ := json.Marshal(map[string]interface{}{"events": []interface{}{}})
	req := httptest.NewRequest("POST", "/api/v1/stream", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestStreamHandler_IngestSingle_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupStreamTestDB(t)
	defer db.Close()

	h := NewStreamHandler(db)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "stream-user")
		c.Next()
	})
	router.POST("/api/v1/event", h.IngestSingle)

	body, _ := json.Marshal(map[string]interface{}{
		"timestamp":  "2025-12-21T10:00:00Z",
		"hostname":   "web01",
		"event_type": "authentication",
		"source_ip":  "10.0.0.2",
	})

	req := httptest.NewRequest("POST", "/api/v1/event", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["imported"] != float64(1) {
		t.Errorf("expected 1 event imported, got %v", resp["imported"])
	}
}

func TestStreamHandler_IngestSingle_MissingFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupStreamTestDB(t)
	defer db.Close()

	h := NewStreamHandler(db)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "stream-user")
		c.Next()
	})
	router.POST("/api/v1/event", h.IngestSingle)

	body, _ := json.Marshal(map[string]interface{}{
		"timestamp": "2025-12-21T10:00:00Z",
	})

	req := httptest.NewRequest("POST", "/api/v1/event", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing fields, got %d", w.Code)
	}
}

func TestStreamHandler_IngestSingle_NoUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupStreamTestDB(t)
	defer db.Close()

	h := NewStreamHandler(db)
	router := gin.New()
	router.POST("/api/v1/event", h.IngestSingle)

	body, _ := json.Marshal(map[string]interface{}{
		"timestamp":  "2025-12-21T10:00:00Z",
		"hostname":   "web01",
		"event_type": "authentication",
	})

	req := httptest.NewRequest("POST", "/api/v1/event", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestStreamHandler_AutoGroupIncidents_NoEvents(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupStreamTestDB(t)
	defer db.Close()

	h := NewStreamHandler(db)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "test-user")
		c.Next()
	})
	router.POST("/api/v1/auto-group", func(c *gin.Context) {
		created, err := h.AutoGroupIncidents("test-user")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"incidents_created": created})
	})

	body, _ := json.Marshal(map[string]interface{}{})
	req := httptest.NewRequest("POST", "/api/v1/auto-group", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["incidents_created"] != float64(0) {
		t.Errorf("expected 0 incidents created, got %v", resp["incidents_created"])
	}
}

func TestStreamHandler_AutoGroupIncidents_WithEvents(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupStreamTestDB(t)
	defer db.Close()

	h := NewStreamHandler(db)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "test-user")
		c.Next()
	})
	router.POST("/api/v1/auto-group", func(c *gin.Context) {
		created, err := h.AutoGroupIncidents("test-user")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"incidents_created": created})
	})

	// Ingest events first
	events := []map[string]interface{}{
		{
			"timestamp":  "2025-12-21T10:00:00Z",
			"hostname":   "dc01",
			"event_type": "authentication",
			"source_ip":  "10.0.0.1",
			"user_name":  "admin",
		},
		{
			"timestamp":  "2025-12-21T10:01:00Z",
			"hostname":   "dc01",
			"event_type": "process_activity",
			"source_ip":  "10.0.0.1",
			"user_name":  "admin",
		},
	}
	body, _ := json.Marshal(map[string]interface{}{"events": events})
	db.InsertEvents([]database.Event{
		{UserID: "test-user", Timestamp: "2025-12-21T10:00:00Z", Hostname: "dc01", EventType: "authentication", SourceIP: "10.0.0.1", UserName: "admin"},
		{UserID: "test-user", Timestamp: "2025-12-21T10:01:00Z", Hostname: "dc01", EventType: "process_activity", SourceIP: "10.0.0.1", UserName: "admin"},
	})
	_ = body

	req := httptest.NewRequest("POST", "/api/v1/auto-group", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["incidents_created"] != float64(1) {
		t.Errorf("expected 1 incident created, got %v", resp["incidents_created"])
	}
}

func TestStreamHandler_BroadcastEvent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupStreamTestDB(t)
	defer db.Close()

	h := NewStreamHandler(db)

	// Create a client channel
	ch := make(chan database.Event, 10)
	h.mu.Lock()
	h.clients[ch] = struct{}{}
	h.mu.Unlock()

	// Broadcast an event
	event := database.Event{
		Timestamp: "2025-12-21T10:00:00Z",
		Hostname:  "dc01",
		EventType: "authentication",
	}
	h.BroadcastEvent(event)

	// Verify the event was received
	select {
	case received := <-ch:
		if received.Hostname != "dc01" {
			t.Errorf("expected hostname dc01, got %s", received.Hostname)
		}
	default:
		t.Error("expected event on channel")
	}

	// Clean up
	h.mu.Lock()
	delete(h.clients, ch)
	h.mu.Unlock()
	close(ch)
}

func TestStreamHandler_BroadcastSlowClient(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupStreamTestDB(t)
	defer db.Close()

	h := NewStreamHandler(db)

	// Create a full client channel (capacity 1, already has one event)
	ch := make(chan database.Event, 1)
	ch <- database.Event{} // fill it

	h.mu.Lock()
	h.clients[ch] = struct{}{}
	h.mu.Unlock()

	// Broadcast should not block (drops event for slow client)
	event := database.Event{Hostname: "dc01"}
	h.BroadcastEvent(event)

	// Clean up
	h.mu.Lock()
	delete(h.clients, ch)
	h.mu.Unlock()
	close(ch)
}
