package auth

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

func setupMiddlewareTestService(t *testing.T) *Service {
	t.Helper()
	dbPath := t.TempDir() + "/test.db"
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_foreign_keys=ON")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	// Create users table (same schema as database/migrate)
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		email TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		api_key TEXT NOT NULL UNIQUE,
		openrouter_key TEXT DEFAULT '',
		created_at TEXT DEFAULT (datetime('now')),
		updated_at TEXT DEFAULT (datetime('now'))
	)`)
	if err != nil {
		t.Fatalf("failed to create users table: %v", err)
	}

	return NewService(db, "test-secret-middleware")
}

func TestAuthMiddleware_NoAuth_Returns401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &Service{}
	mw := AuthMiddleware(svc)

	router := gin.New()
	router.Use(mw)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"user_id": c.GetString("user_id")})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_BearerToken_Valid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := setupMiddlewareTestService(t)

	_, err := svc.Signup("test@example.com", "password123")
	if err != nil {
		t.Fatalf("signup failed: %v", err)
	}
	token, _, err := svc.Login("test@example.com", "password123")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	mw := AuthMiddleware(svc)
	router := gin.New()
	router.Use(mw)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"user_id": c.GetString("user_id")})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAuthMiddleware_BearerToken_Invalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &Service{}
	mw := AuthMiddleware(svc)

	router := gin.New()
	router.Use(mw)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_XAPIKey_Invalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &Service{}
	mw := AuthMiddleware(svc)

	router := gin.New()
	router.Use(mw)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "invalid-key")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_XAPIKey_Valid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := setupMiddlewareTestService(t)

	user, err := svc.Signup("test@example.com", "password123")
	if err != nil {
		t.Fatalf("signup failed: %v", err)
	}

	mw := AuthMiddleware(svc)
	router := gin.New()
	router.Use(mw)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"user_id": c.GetString("user_id")})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", user.APIKey)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetUserID_FromContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c := &gin.Context{}
	c.Set("user_id", "test-user-123")

	id := GetUserID(c)
	if id != "test-user-123" {
		t.Errorf("expected 'test-user-123', got '%s'", id)
	}
}

func TestGetUserID_NotSet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c := &gin.Context{}

	id := GetUserID(c)
	if id != "" {
		t.Errorf("expected empty string, got '%s'", id)
	}
}

func TestGetUserID_NonString(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c := &gin.Context{}
	c.Set("user_id", 12345)

	id := GetUserID(c)
	if id != "" {
		t.Errorf("expected empty string for non-string, got '%s'", id)
	}
}
