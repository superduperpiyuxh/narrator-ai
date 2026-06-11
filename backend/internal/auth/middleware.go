package auth

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try Authorization header first (Bearer token)
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
				claims, err := svc.ValidateToken(parts[1])
				if err == nil {
					c.Set("user_id", claims.UserID)
					c.Set("user_email", claims.Email)
					c.Next()
					return
				}
				// Token invalid/expired — fall through to demo mode
			}
		}

		// Try X-API-Key header
		apiKey := c.GetHeader("X-API-Key")
		if apiKey != "" {
			user, err := svc.GetUserByAPIKey(apiKey)
			if err == nil && user != nil {
				c.Set("user_id", user.ID)
				c.Set("user_email", user.Email)
				c.Next()
				return
			}
			// Invalid API key — fall through to demo mode
		}

		// No auth — pass through in demo mode (user_id = "")
		c.Set("user_id", "")
		c.Set("user_email", "")
		c.Next()
	}
}

func GetUserID(c *gin.Context) string {
	userID, _ := c.Get("user_id")
	if id, ok := userID.(string); ok {
		return id
	}
	return ""
}
