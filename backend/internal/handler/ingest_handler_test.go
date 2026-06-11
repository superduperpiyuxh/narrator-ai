package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/superduperpiyuxh/narrator-ai/backend/internal/database"
)

func setupIngestTestDB(t *testing.T) *database.DB {
	t.Helper()
	dbPath := t.TempDir() + "/test.db"
	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	return db
}

func TestIngestHandler_IngestEvents_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupIngestTestDB(t)
	defer db.Close()

	h := NewIngestHandler(db, t.TempDir())
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "test-user")
		c.Next()
	})
	router.POST("/api/v1/ingest", h.IngestEvents)

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

	req := httptest.NewRequest("POST", "/api/v1/ingest", bytes.NewReader(body))
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

func TestIngestHandler_IngestEvents_EmptyBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupIngestTestDB(t)
	defer db.Close()

	h := NewIngestHandler(db, t.TempDir())
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "test-user")
		c.Next()
	})
	router.POST("/api/v1/ingest", h.IngestEvents)

	body, _ := json.Marshal(map[string]interface{}{"events": []interface{}{}})
	req := httptest.NewRequest("POST", "/api/v1/ingest", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty events, got %d", w.Code)
	}
}

func TestIngestHandler_IngestEvents_MissingFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupIngestTestDB(t)
	defer db.Close()

	h := NewIngestHandler(db, t.TempDir())
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "test-user")
		c.Next()
	})
	router.POST("/api/v1/ingest", h.IngestEvents)

	events := []map[string]interface{}{
		{"timestamp": "2025-12-21T10:00:00Z"},
	}
	body, _ := json.Marshal(map[string]interface{}{"events": events})

	req := httptest.NewRequest("POST", "/api/v1/ingest", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["imported"] != float64(0) {
		t.Errorf("expected 0 events imported (missing fields), got %v", resp["imported"])
	}
}

func TestIngestHandler_IngestFile_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupIngestTestDB(t)
	defer db.Close()

	dataDir := t.TempDir()
	jsonFile := filepath.Join(dataDir, "test.jsonl")
	os.WriteFile(jsonFile, []byte(`{"timestamp":"2025-12-21T10:00:00Z","hostname":"web01","event_type":"authentication","source_ip":"10.0.0.2"}
`), 0644)

	h := NewIngestHandler(db, dataDir)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "test-user")
		c.Next()
	})
	router.POST("/api/v1/ingest/file", h.IngestFile)

	body, _ := json.Marshal(map[string]string{"file_path": jsonFile})
	req := httptest.NewRequest("POST", "/api/v1/ingest/file", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestIngestHandler_IngestFile_PathTraversal(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupIngestTestDB(t)
	defer db.Close()

	dataDir := t.TempDir()
	h := NewIngestHandler(db, dataDir)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "test-user")
		c.Next()
	})
	router.POST("/api/v1/ingest/file", h.IngestFile)

	body, _ := json.Marshal(map[string]string{"file_path": "/etc/passwd"})
	req := httptest.NewRequest("POST", "/api/v1/ingest/file", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for path traversal, got %d", w.Code)
	}
}

func TestIngestHandler_IngestFile_RelativePath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupIngestTestDB(t)
	defer db.Close()

	h := NewIngestHandler(db, t.TempDir())
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "test-user")
		c.Next()
	})
	router.POST("/api/v1/ingest/file", h.IngestFile)

	body, _ := json.Marshal(map[string]string{"file_path": "../etc/passwd"})
	req := httptest.NewRequest("POST", "/api/v1/ingest/file", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for relative path, got %d", w.Code)
	}
}
