package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/superduperpiyuxh/narrator-ai/backend/internal/auth"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupTestRouter(svc *auth.Service) *gin.Engine {
	r := gin.New()
	h := NewAuthHandler(svc)
	r.POST("/api/auth/signup", h.Signup)
	r.POST("/api/auth/login", h.Login)
	return r
}

func TestAuthHandler_Signup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	// Would need a test database
}

func TestAuthHandler_Signup_BadRequest(t *testing.T) {
	svc := &auth.Service{}
	r := setupTestRouter(svc)

	body, _ := json.Marshal(map[string]string{})
	req := httptest.NewRequest("POST", "/api/auth/signup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAuthHandler_Signup_ShortPassword(t *testing.T) {
	svc := &auth.Service{}
	r := setupTestRouter(svc)

	body, _ := json.Marshal(map[string]string{
		"email":    "test@test.com",
		"password": "short",
	})
	req := httptest.NewRequest("POST", "/api/auth/signup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for short password, got %d", w.Code)
	}
}

func TestAuthHandler_Login_BadRequest(t *testing.T) {
	svc := &auth.Service{}
	r := setupTestRouter(svc)

	body, _ := json.Marshal(map[string]string{})
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	svc := &auth.Service{}
	r := setupTestRouter(svc)

	body, _ := json.Marshal(map[string]string{
		"email":    "nonexistent@test.com",
		"password": "password123",
	})
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}
