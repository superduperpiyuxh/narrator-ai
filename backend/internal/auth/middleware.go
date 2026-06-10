package auth

import (
	"net/http"
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
				if err != nil {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
					c.Abort()
					return
				}
				c.Set("user_id", claims.UserID)
				c.Set("user_email", claims.Email)
				c.Next()
				return
			}
		}

		// Try X-API-Key header
		apiKey := c.GetHeader("X-API-Key")
		if apiKey != "" {
			user, err := svc.GetUserByAPIKey(apiKey)
			if err != nil || user == nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid API key"})
				c.Abort()
				return
			}
			c.Set("user_id", user.ID)
			c.Set("user_email", user.Email)
			c.Next()
			return
		}

		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		c.Abort()
	}
}

func GetUserID(c *gin.Context) string {
	userID, _ := c.Get("user_id")
	if id, ok := userID.(string); ok {
		return id
	}
	return ""
}
