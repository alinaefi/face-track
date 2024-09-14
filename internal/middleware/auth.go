package middleware

import (
	"face-track/utils"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

const (
	apiAuthUsernameEnvName = "FACE_TRACK__API_USER"
	apiAuthPasswordEnvName = "FACE_TRACK__API_PASS"
)

type AuthMiddleware struct {
	username string
	password string
}

func NewAuthMiddleware() *AuthMiddleware {
	utils.CheckEnvs(apiAuthUsernameEnvName, apiAuthPasswordEnvName)

	return &AuthMiddleware{
		username: os.Getenv(apiAuthUsernameEnvName),
		password: os.Getenv(apiAuthPasswordEnvName),
	}
}

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
