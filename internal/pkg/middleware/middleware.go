// Package middleware provides HTTP middleware utilities for request handling.
package middleware

import (
	"net/http"
	"os"

	"face-track/tools"

	"github.com/gin-gonic/gin"
)

const (
	// apiAuthUsernameEnvName is the env variable key for the API username.
	apiAuthUsernameEnvName = "FACE_TRACK__API_USER"
	// apiAuthPasswordEnvName is the env variable key for the API password.
	apiAuthPasswordEnvName = "FACE_TRACK__API_PASS"
)

// AuthMiddleware handles basic authentication for API requests.
type AuthMiddleware struct {
	username string
	password string
}

// NewAuthMiddleware creates and returns an AuthMiddleware instance.
func NewAuthMiddleware() *AuthMiddleware {
	tools.CheckEnvs(apiAuthUsernameEnvName, apiAuthPasswordEnvName)

	return &AuthMiddleware{
		username: os.Getenv(apiAuthUsernameEnvName),
		password: os.Getenv(apiAuthPasswordEnvName),
	}
}

// BasicAuthMiddleware returns a Gin middleware that enforces basic authentication.
func (m *AuthMiddleware) BasicAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, pass, ok := c.Request.BasicAuth()
		if !ok || user != m.username || pass != m.password {
			c.Header("WWW-Authenticate", `Basic realm="Restricted"`)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Next()
	}
}
