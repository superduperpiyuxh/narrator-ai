package auth

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type RateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Remove old entries
	requests := rl.requests[key]
	valid := requests[:0]
	for _, t := range requests {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= rl.limit {
		rl.requests[key] = valid
		return false
	}

	rl.requests[key] = append(valid, now)
	return true
}

func RateLimitMiddleware(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == "" {
			userID = c.ClientIP()
		}

		if !limiter.Allow(userID) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded. Try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
