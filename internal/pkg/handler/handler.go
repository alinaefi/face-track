// Package handler provides HTTP server setup and request handling for the application.
package handler

import (
	"face-track/internal/pkg/service"
	"net/http"
	"os"

	"face-track/tools"

	"github.com/gin-gonic/gin"
)

// serverAddrName is an env variable key for the Face Track server address.
const serverAddrName = "FACE_TRACK__SERVER_ADDRESS"

// Handler is responsible for handling incoming HTTP requests and routing them
// to the appropriate service methods.
type Handler struct {
	service *service.Service
}

// NewHandler returns a new Handler instance.
func NewHandler(service *service.Service) *Handler {
	return &Handler{
		service: service,
	}
}

// NewServer initializes and returns an HTTP server.
func NewServer(s *service.Service) *http.Server {
	tools.CheckEnvs(serverAddrName)

	serverAddress := os.Getenv(serverAddrName)

	router := gin.Default()

	handler := NewHandler(s)

	taskApi := router.Group("/api")

	handler.setTaskGroup(taskApi)

	return &http.Server{
		Addr:    serverAddress,
		Handler: router,
	}
}
