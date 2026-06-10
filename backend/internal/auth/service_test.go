package auth

import (
	"context"
	"testing"
)

func TestService_Signup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	// Test would require a database connection
	// Placeholder for integration tests
}

func TestService_Login(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	// Test would require a database connection
	// Placeholder for integration tests
}

func TestService_GenerateToken(t *testing.T) {
	svc := &Service{
		jwtSecret: []byte("test-secret"),
	}

	token, err := svc.generateToken("user-123", "test@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestService_ValidateToken(t *testing.T) {
	svc := &Service{
		jwtSecret: []byte("test-secret"),
	}

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
}

func TestService_ValidateToken_Invalid(t *testing.T) {
	svc := &Service{
		jwtSecret: []byte("test-secret"),
	}

	_, err := svc.ValidateToken("invalid-token")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestService_ValidateToken_WrongSecret(t *testing.T) {
	svc1 := &Service{
		jwtSecret: []byte("secret-1"),
	}
	svc2 := &Service{
		jwtSecret: []byte("secret-2"),
	}

	token, err := svc1.generateToken("user-123", "test@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = svc2.ValidateToken(token)
	if err == nil {
		t.Fatal("expected error for wrong secret")
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
