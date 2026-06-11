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

func setupIncidentHandlerTestDB(t *testing.T) *database.DB {
	t.Helper()
	dbPath := t.TempDir() + "/test.db"
	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	return db
}

func TestGetIncidents(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupIncidentHandlerTestDB(t)
	defer db.Close()

	db.CreateIncident(&database.Incident{Title: "Test", SourceIP: "10.0.0.1", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", Severity: "high", Status: "new"}, nil)

	h := NewIncidentHandler(db)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "")
		c.Next()
	})
	router.GET("/api/incidents", h.GetIncidents)

	req := httptest.NewRequest("GET", "/api/incidents?limit=10&offset=0", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["total"] != float64(1) {
		t.Errorf("expected 1 incident, got %v", resp["total"])
	}
}

func TestGetIncidents_FilterBySeverity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupIncidentHandlerTestDB(t)
	defer db.Close()

	db.CreateIncident(&database.Incident{Title: "High", SourceIP: "10.0.0.1", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", Severity: "high", Status: "new"}, nil)
	db.CreateIncident(&database.Incident{Title: "Low", SourceIP: "10.0.0.2", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", Severity: "low", Status: "new"}, nil)

	h := NewIncidentHandler(db)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "")
		c.Next()
	})
	router.GET("/api/incidents", h.GetIncidents)

	req := httptest.NewRequest("GET", "/api/incidents?severity=high", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["total"] != float64(1) {
		t.Errorf("expected 1 high incident, got %v", resp["total"])
	}
}

func TestGetIncident(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupIncidentHandlerTestDB(t)
	defer db.Close()

	inc := &database.Incident{Title: "Test", SourceIP: "10.0.0.1", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", Severity: "high", Status: "new"}
	db.CreateIncident(inc, nil)

	h := NewIncidentHandler(db)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "")
		c.Next()
	})
	router.GET("/api/incidents/:id", h.GetIncident)

	req := httptest.NewRequest("GET", "/api/incidents/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetIncident_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupIncidentHandlerTestDB(t)
	defer db.Close()

	h := NewIncidentHandler(db)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "")
		c.Next()
	})
	router.GET("/api/incidents/:id", h.GetIncident)

	req := httptest.NewRequest("GET", "/api/incidents/abc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetIncident_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupIncidentHandlerTestDB(t)
	defer db.Close()

	h := NewIncidentHandler(db)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "")
		c.Next()
	})
	router.GET("/api/incidents/:id", h.GetIncident)

	req := httptest.NewRequest("GET", "/api/incidents/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestGetIncidentEvents(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupIncidentHandlerTestDB(t)
	defer db.Close()

	events := []database.Event{
		{Timestamp: "2025-12-21T10:00:00Z", Hostname: "dc01", EventType: "auth", SourceIP: "10.0.0.1"},
	}
	db.InsertEvents(events)

	inc := &database.Incident{Title: "Test", SourceIP: "10.0.0.1", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", Severity: "high", Status: "new"}
	db.CreateIncident(inc, []int64{1})

	h := NewIncidentHandler(db)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "")
		c.Next()
	})
	router.GET("/api/incidents/:id/events", h.GetIncidentEvents)

	req := httptest.NewRequest("GET", "/api/incidents/1/events", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetIncidentStats(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupIncidentHandlerTestDB(t)
	defer db.Close()

	db.CreateIncident(&database.Incident{Title: "Test", SourceIP: "10.0.0.1", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", EventCount: 10, Severity: "high", Status: "new"}, nil)

	h := NewIncidentHandler(db)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "")
		c.Next()
	})
	router.GET("/api/incidents/stats", h.GetIncidentStats)

	req := httptest.NewRequest("GET", "/api/incidents/stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetTechniques(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupIncidentHandlerTestDB(t)
	defer db.Close()

	h := NewIncidentHandler(db)
	router := gin.New()
	router.GET("/api/techniques", h.GetTechniques)

	req := httptest.NewRequest("GET", "/api/techniques", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGroupIncidents(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupIncidentHandlerTestDB(t)
	defer db.Close()

	// Insert events with same source_ip within 15 minutes
	events := []database.Event{
		{Timestamp: "2025-12-21T10:00:00Z", Hostname: "dc01", EventType: "auth", SourceIP: "10.0.0.1", UserName: "admin"},
		{Timestamp: "2025-12-21T10:01:00Z", Hostname: "dc01", EventType: "process", SourceIP: "10.0.0.1", UserName: "admin"},
	}
	db.InsertEvents(events)

	h := NewIncidentHandler(db)
	router := gin.New()
	router.POST("/api/incidents/group", h.GroupIncidents)

	req := httptest.NewRequest("POST", "/api/incidents/group", nil)
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

func TestGetIncident_OwnershipCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupIncidentHandlerTestDB(t)
	defer db.Close()

	// Create incident with user_id
	inc := &database.Incident{UserID: "user-1", Title: "Owned", SourceIP: "10.0.0.1", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", Severity: "high", Status: "new"}
	db.CreateIncident(inc, nil)

	h := NewIncidentHandler(db)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "user-2") // Different user
		c.Next()
	})
	router.GET("/api/incidents/:id", h.GetIncident)

	body, _ := json.Marshal(map[string]string{})
	req := httptest.NewRequest("GET", "/api/incidents/1", bytes.NewReader(body))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for wrong user, got %d", w.Code)
	}
}

func TestGetIncident_OwnerCanAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupIncidentHandlerTestDB(t)
	defer db.Close()

	inc := &database.Incident{UserID: "user-1", Title: "Owned", SourceIP: "10.0.0.1", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", Severity: "high", Status: "new"}
	db.CreateIncident(inc, nil)

	h := NewIncidentHandler(db)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "user-1")
		c.Next()
	})
	router.GET("/api/incidents/:id", h.GetIncident)

	req := httptest.NewRequest("GET", "/api/incidents/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for owner, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetIncident_EmptyUserID_AccessAll(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupIncidentHandlerTestDB(t)
	defer db.Close()

	// Incident with no user_id (empty) — should be accessible by anyone
	inc := &database.Incident{UserID: "", Title: "Public", SourceIP: "10.0.0.1", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", Severity: "high", Status: "new"}
	db.CreateIncident(inc, nil)

	h := NewIncidentHandler(db)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "user-1")
		c.Next()
	})
	router.GET("/api/incidents/:id", h.GetIncident)

	req := httptest.NewRequest("GET", "/api/incidents/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for empty user_id incident, got %d", w.Code)
	}
}

func TestGetIncidentEvents_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupIncidentHandlerTestDB(t)
	defer db.Close()

	h := NewIncidentHandler(db)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "")
		c.Next()
	})
	router.GET("/api/incidents/:id/events", h.GetIncidentEvents)

	req := httptest.NewRequest("GET", "/api/incidents/abc/events", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetIncidentEvents_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupIncidentHandlerTestDB(t)
	defer db.Close()

	h := NewIncidentHandler(db)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "")
		c.Next()
	})
	router.GET("/api/incidents/:id/events", h.GetIncidentEvents)

	req := httptest.NewRequest("GET", "/api/incidents/999/events", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for nonexistent incident, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetIncidentStats_Empty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupIncidentHandlerTestDB(t)
	defer db.Close()

	h := NewIncidentHandler(db)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "")
		c.Next()
	})
	router.GET("/api/incidents/stats", h.GetIncidentStats)

	req := httptest.NewRequest("GET", "/api/incidents/stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["total_incidents"] != float64(0) {
		t.Errorf("expected 0, got %v", resp["total_incidents"])
	}
}

func TestGetIncidentStats_WithData(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupIncidentHandlerTestDB(t)
	defer db.Close()

	db.CreateIncident(&database.Incident{Title: "High", SourceIP: "10.0.0.1", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", EventCount: 10, Severity: "high", Status: "new"}, nil)
	db.CreateIncident(&database.Incident{Title: "Low", SourceIP: "10.0.0.2", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", EventCount: 5, Severity: "low", Status: "new"}, nil)

	h := NewIncidentHandler(db)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "")
		c.Next()
	})
	router.GET("/api/incidents/stats", h.GetIncidentStats)

	req := httptest.NewRequest("GET", "/api/incidents/stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["total_incidents"] != float64(2) {
		t.Errorf("expected 2, got %v", resp["total_incidents"])
	}
}

func TestGetTechniques_Empty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupIncidentHandlerTestDB(t)
	defer db.Close()

	h := NewIncidentHandler(db)
	router := gin.New()
	router.GET("/api/techniques", h.GetTechniques)

	req := httptest.NewRequest("GET", "/api/techniques", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	techniques := resp["techniques"].([]interface{})
	// GetTechniques returns hardcoded ATT&CK techniques, not from DB
	if len(techniques) == 0 {
		t.Error("expected non-empty techniques list")
	}
}

func TestGetTechniqueCounts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupIncidentHandlerTestDB(t)
	defer db.Close()

	db.SeedTechniques([]database.TechniqueRef{
		{TechniqueID: "T1110", Name: "Brute Force", Tactic: "credential-access"},
	})

	inc := &database.Incident{Title: "Test", SourceIP: "10.0.0.1", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", Severity: "high", Status: "new", Techniques: []database.TechniqueRef{{TechniqueID: "T1110", EventCount: 5}}}
	db.CreateIncident(inc, nil)

	h := NewIncidentHandler(db)
	router := gin.New()
	router.GET("/api/techniques/counts", h.GetTechniqueCounts)

	req := httptest.NewRequest("GET", "/api/techniques/counts", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	counts := resp["counts"].(map[string]interface{})
	if counts["T1110"] != float64(1) {
		t.Errorf("expected 1 for T1110, got %v", counts["T1110"])
	}
}

func TestGetTechniqueCounts_Empty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupIncidentHandlerTestDB(t)
	defer db.Close()

	h := NewIncidentHandler(db)
	router := gin.New()
	router.GET("/api/techniques/counts", h.GetTechniqueCounts)

	req := httptest.NewRequest("GET", "/api/techniques/counts", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestGroupIncidents_NoEvents(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupIncidentHandlerTestDB(t)
	defer db.Close()

	h := NewIncidentHandler(db)
	router := gin.New()
	router.POST("/api/incidents/group", h.GroupIncidents)

	req := httptest.NewRequest("POST", "/api/incidents/group", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["incidents_created"] != float64(0) {
		t.Errorf("expected 0 incidents, got %v", resp["incidents_created"])
	}
}

func TestGetIncidents_MultipleFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupIncidentHandlerTestDB(t)
	defer db.Close()

	db.CreateIncident(&database.Incident{Title: "A", SourceIP: "10.0.0.1", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", Severity: "high", Status: "new"}, nil)
	db.CreateIncident(&database.Incident{Title: "B", SourceIP: "10.0.0.2", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", Severity: "low", Status: "closed"}, nil)

	h := NewIncidentHandler(db)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "")
		c.Next()
	})
	router.GET("/api/incidents", h.GetIncidents)

	req := httptest.NewRequest("GET", "/api/incidents?severity=high&status=new", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["total"] != float64(1) {
		t.Errorf("expected 1, got %v", resp["total"])
	}
}
