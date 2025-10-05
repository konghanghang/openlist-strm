package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// tokenAuthMiddleware validates API token
func (s *Server) tokenAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("X-API-Token")
		if token == "" {
			token = c.GetHeader("Authorization")
		}

		if token != s.cfg.API.Token {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or missing API token",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
