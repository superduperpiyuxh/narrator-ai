package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestNewRateLimiter(t *testing.T) {
	rl := NewRateLimiter(5, time.Minute)
	if rl == nil {
		t.Fatal("expected non-nil rate limiter")
	}
	if rl.limit != 5 {
		t.Errorf("expected limit 5, got %d", rl.limit)
	}
}

func TestRateLimiter_Allow(t *testing.T) {
	rl := NewRateLimiter(3, time.Minute)

	for i := 0; i < 3; i++ {
		if !rl.Allow("user-1") {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}

	// 4th request should be blocked
	if rl.Allow("user-1") {
		t.Error("expected 4th request to be blocked")
	}

	// Different key should still work
	if !rl.Allow("user-2") {
		t.Error("different key should be allowed")
	}
}

func TestRateLimiter_WindowReset(t *testing.T) {
	rl := NewRateLimiter(2, 50*time.Millisecond)

	rl.Allow("user-1")
	rl.Allow("user-1")

	// Blocked
	if rl.Allow("user-1") {
		t.Error("should be blocked")
	}

	// Wait for window to expire
	time.Sleep(60 * time.Millisecond)

	if !rl.Allow("user-1") {
		t.Error("should be allowed after window reset")
	}
}

func TestRateLimiter_EmptyKey(t *testing.T) {
	rl := NewRateLimiter(1, time.Minute)

	if !rl.Allow("") {
		t.Error("empty key should be allowed first time")
	}
	if rl.Allow("") {
		t.Error("empty key should be blocked second time")
	}
}

func TestRateLimitMiddleware_Allows(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rl := NewRateLimiter(5, time.Minute)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "test-user")
		c.Next()
	})
	router.Use(RateLimitMiddleware(rl))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestRateLimitMiddleware_Blocks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rl := NewRateLimiter(1, time.Minute)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "test-user")
		c.Next()
	})
	router.Use(RateLimitMiddleware(rl))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	// First request OK
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Second request blocked
	req = httptest.NewRequest("GET", "/test", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", w.Code)
	}
}

func TestRateLimitMiddleware_NoUserID_UsesIP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rl := NewRateLimiter(1, time.Minute)

	router := gin.New()
	router.Use(RateLimitMiddleware(rl))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Second request with no user_id (uses IP) should be blocked
	req = httptest.NewRequest("GET", "/test", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", w.Code)
	}
}
