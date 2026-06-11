package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupTestService() *Service {
	return &Service{
		jwtSecret: []byte("test-secret"),
	}
}

func setupRouter(svc *Service) *gin.Engine {
	r := gin.New()
	r.Use(AuthMiddleware(svc))
	r.GET("/test", func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		userEmail, _ := c.Get("user_email")
		c.JSON(http.StatusOK, gin.H{
			"user_id": userID,
			"email":   userEmail,
		})
	})
	return r
}

func TestAuthMiddleware_BearerToken(t *testing.T) {
	svc := setupTestService()
	r := setupRouter(svc)

	token, err := svc.generateToken("user-abc", "test@example.com")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["user_id"] != "user-abc" {
		t.Errorf("expected user_id 'user-abc', got '%s'", resp["user_id"])
	}
	if resp["email"] != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got '%s'", resp["email"])
	}
}

func TestAuthMiddleware_NoAuth(t *testing.T) {
	svc := setupTestService()
	r := setupRouter(svc)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	svc := setupTestService()
	r := setupRouter(svc)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token-here")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_WrongSecretToken(t *testing.T) {
	svc1 := &Service{jwtSecret: []byte("secret-1")}
	svc2 := &Service{jwtSecret: []byte("secret-2")}
	r := setupRouter(svc2)

	token, _ := svc1.generateToken("user-abc", "test@example.com")

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for wrong secret, got %d", w.Code)
	}
}

func TestAuthMiddleware_InvalidAPIKey(t *testing.T) {
	svc := &Service{
		jwtSecret: []byte("test-secret"),
		db:        nil,
	}
	r := setupRouter(svc)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "nai_invalid_key")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_MalformedBearer(t *testing.T) {
	svc := setupTestService()
	r := setupRouter(svc)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer ")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// "Bearer " with empty token — validation fails → 401
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_CaseInsensitiveBearer(t *testing.T) {
	svc := setupTestService()
	r := setupRouter(svc)

	token, _ := svc.generateToken("user-123", "user@test.com")

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["user_id"] != "user-123" {
		t.Errorf("expected user_id 'user-123', got '%s'", resp["user_id"])
	}
}

func TestService_GenerateToken(t *testing.T) {
	svc := setupTestService()

	token, err := svc.generateToken("user-123", "test@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestService_ValidateToken(t *testing.T) {
	svc := setupTestService()

	token, err := svc.generateToken("user-123", "test@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	claims, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if claims.UserID != "user-123" {
		t.Errorf("expected user_id 'user-123', got '%s'", claims.UserID)
	}
	if claims.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got '%s'", claims.Email)
	}
	if claims.Issuer != "narrator-ai" {
		t.Errorf("expected issuer 'narrator-ai', got '%s'", claims.Issuer)
	}
}

func TestService_ValidateToken_Invalid(t *testing.T) {
	svc := setupTestService()

	_, err := svc.ValidateToken("invalid-token")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestService_ValidateToken_WrongSecret(t *testing.T) {
	svc1 := &Service{jwtSecret: []byte("secret-1")}
	svc2 := &Service{jwtSecret: []byte("secret-2")}

	token, err := svc1.generateToken("user-123", "test@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = svc2.ValidateToken(token)
	if err == nil {
		t.Fatal("expected error for wrong secret")
	}
}

func TestService_ValidateToken_EmptyString(t *testing.T) {
	svc := setupTestService()

	_, err := svc.ValidateToken("")
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestContextWithUserID(t *testing.T) {
	ctx := ContextWithUserID(context.Background(), "user-123")
	if ctx == nil {
		t.Fatal("expected non-nil context")
	}

	userID, ok := UserIDFromContext(ctx)
	if !ok {
		t.Fatal("expected ok to be true")
	}
	if userID != "user-123" {
		t.Errorf("expected 'user-123', got '%s'", userID)
	}
}

func TestUserIDFromContext_Empty(t *testing.T) {
	userID, ok := UserIDFromContext(context.Background())
	if ok {
		t.Fatal("expected ok to be false")
	}
	if userID != "" {
		t.Errorf("expected empty string, got '%s'", userID)
	}
}

func TestGetUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", "test-user")

	userID := GetUserID(c)
	if userID != "test-user" {
		t.Errorf("expected 'test-user', got '%s'", userID)
	}
}

func TestGetUserID_Missing(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	userID := GetUserID(c)
	if userID != "" {
		t.Errorf("expected empty string, got '%s'", userID)
	}
}

func TestGetUserID_WrongType(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", 12345)

	userID := GetUserID(c)
	if userID != "" {
		t.Errorf("expected empty string for wrong type, got '%s'", userID)
	}
}
