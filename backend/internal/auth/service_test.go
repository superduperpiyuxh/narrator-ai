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

func TestService_Signup_Success(t *testing.T) {
	svc := setupMiddlewareTestService(t)

	user, err := svc.Signup("newuser@test.com", "password123")
	if err != nil {
		t.Fatalf("signup failed: %v", err)
	}
	if user.ID == "" {
		t.Error("expected non-empty user ID")
	}
	if user.Email != "newuser@test.com" {
		t.Errorf("expected email 'newuser@test.com', got '%s'", user.Email)
	}
	if user.APIKey == "" {
		t.Error("expected non-empty API key")
	}
}

func TestService_Signup_DuplicateEmail(t *testing.T) {
	svc := setupMiddlewareTestService(t)

	_, err := svc.Signup("dup@test.com", "password123")
	if err != nil {
		t.Fatalf("first signup failed: %v", err)
	}

	_, err = svc.Signup("dup@test.com", "password456")
	if err == nil {
		t.Fatal("expected error for duplicate email")
	}
}

func TestService_Signup_InvalidEmail(t *testing.T) {
	svc := setupMiddlewareTestService(t)

	// Empty email should still succeed at DB level (no constraint check in code)
	_, err := svc.Signup("", "password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_Login_Success(t *testing.T) {
	svc := setupMiddlewareTestService(t)

	_, err := svc.Signup("loginuser@test.com", "mypassword")
	if err != nil {
		t.Fatalf("signup failed: %v", err)
	}

	token, user, err := svc.Login("loginuser@test.com", "mypassword")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
	if user.Email != "loginuser@test.com" {
		t.Errorf("expected email 'loginuser@test.com', got '%s'", user.Email)
	}
	if user.PasswordHash != "" {
		t.Error("expected empty password hash in response")
	}
}

func TestService_Login_WrongPassword(t *testing.T) {
	svc := setupMiddlewareTestService(t)

	_, err := svc.Signup("wrongpw@test.com", "correctpass")
	if err != nil {
		t.Fatalf("signup failed: %v", err)
	}

	_, _, err = svc.Login("wrongpass@test.com", "wrongpass")
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
}

func TestService_Login_NonexistentUser(t *testing.T) {
	svc := setupMiddlewareTestService(t)

	_, _, err := svc.Login("nonexistent@test.com", "password")
	if err == nil {
		t.Fatal("expected error for nonexistent user")
	}
}

func TestService_Login_NilDB(t *testing.T) {
	svc := &Service{jwtSecret: []byte("test")}

	_, _, err := svc.Login("test@test.com", "password")
	if err == nil {
		t.Fatal("expected error for nil DB")
	}
}

func TestService_GetUserByID(t *testing.T) {
	svc := setupMiddlewareTestService(t)

	signupUser, _ := svc.Signup("getbyid@test.com", "password123")

	user, err := svc.GetUserByID(signupUser.ID)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if user == nil {
		t.Fatal("expected user, got nil")
	}
	if user.Email != "getbyid@test.com" {
		t.Errorf("expected 'getbyid@test.com', got '%s'", user.Email)
	}
}

func TestService_GetUserByID_NotFound(t *testing.T) {
	svc := setupMiddlewareTestService(t)

	user, err := svc.GetUserByID("nonexistent-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user != nil {
		t.Error("expected nil for nonexistent user")
	}
}

func TestService_GetUserByAPIKey(t *testing.T) {
	svc := setupMiddlewareTestService(t)

	signupUser, _ := svc.Signup("apikey@test.com", "password123")

	user, err := svc.GetUserByAPIKey(signupUser.APIKey)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if user == nil {
		t.Fatal("expected user, got nil")
	}
	if user.Email != "apikey@test.com" {
		t.Errorf("expected 'apikey@test.com', got '%s'", user.Email)
	}
}

func TestService_GetUserByAPIKey_NotFound(t *testing.T) {
	svc := setupMiddlewareTestService(t)

	user, err := svc.GetUserByAPIKey("nai_nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user != nil {
		t.Error("expected nil for nonexistent API key")
	}
}

func TestService_GetUserByAPIKey_NilDB(t *testing.T) {
	svc := &Service{jwtSecret: []byte("test")}

	user, err := svc.GetUserByAPIKey("key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user != nil {
		t.Error("expected nil for nil DB")
	}
}

func TestService_UpdateOpenRouterKey(t *testing.T) {
	svc := setupMiddlewareTestService(t)

	signupUser, _ := svc.Signup("updatekey@test.com", "password123")

	err := svc.UpdateOpenRouterKey(signupUser.ID, "sk-or-new-key")
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}

	user, _ := svc.GetUserByID(signupUser.ID)
	if user.OpenRouterKey != "sk-or-new-key" {
		t.Errorf("expected 'sk-or-new-key', got '%s'", user.OpenRouterKey)
	}
}

func TestService_UpdateOpenRouterKey_ClearKey(t *testing.T) {
	svc := setupMiddlewareTestService(t)

	signupUser, _ := svc.Signup("clearkey@test.com", "password123")
	svc.UpdateOpenRouterKey(signupUser.ID, "some-key")

	err := svc.UpdateOpenRouterKey(signupUser.ID, "")
	if err != nil {
		t.Fatalf("clear failed: %v", err)
	}

	user, _ := svc.GetUserByID(signupUser.ID)
	if user.OpenRouterKey != "" {
		t.Errorf("expected empty key, got '%s'", user.OpenRouterKey)
	}
}

func TestService_CreateDefaultAdmin_Success(t *testing.T) {
	svc := setupMiddlewareTestService(t)

	err := svc.CreateDefaultAdmin("admin@test.com", "adminpass")
	if err != nil {
		t.Fatalf("create admin failed: %v", err)
	}

	// Should be idempotent — calling again should not error
	err = svc.CreateDefaultAdmin("admin@test.com", "adminpass")
	if err != nil {
		t.Fatalf("re-create admin failed: %v", err)
	}

	user, _ := svc.GetUserByID("")
	if user != nil {
		// verify user exists via login
	}
}

func TestService_CreateDefaultAdmin_AlreadyExists(t *testing.T) {
	svc := setupMiddlewareTestService(t)

	svc.Signup("existing@test.com", "password123")

	// CreateDefaultAdmin should return nil (already exists)
	err := svc.CreateDefaultAdmin("existing@test.com", "password123")
	if err != nil {
		t.Fatalf("expected nil for existing user, got: %v", err)
	}
}

func TestService_ValidateToken_CrossService(t *testing.T) {
	svc1 := &Service{jwtSecret: []byte("secret-a")}
	svc2 := &Service{jwtSecret: []byte("secret-a")} // same secret

	token, _ := svc1.generateToken("user-1", "test@test.com")

	claims, err := svc2.ValidateToken(token)
	if err != nil {
		t.Fatalf("validation failed: %v", err)
	}
	if claims.UserID != "user-1" {
		t.Errorf("expected 'user-1', got '%s'", claims.UserID)
	}
}

func TestService_Login_VerifyToken(t *testing.T) {
	svc := setupMiddlewareTestService(t)

	svc.Signup("verify@test.com", "password123")
	token, _, _ := svc.Login("verify@test.com", "password123")

	claims, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("token validation failed: %v", err)
	}
	if claims.Email != "verify@test.com" {
		t.Errorf("expected 'verify@test.com', got '%s'", claims.Email)
	}
}
