package handler

import (
	"face-track/internal/pkg/service"
	"face-track/utils"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

const (
	serverAddrName = "FACE_TRACK__SERVER_ADDRESS"
)

type Handler struct {
	service *service.Service
}

func NewHandler(service *service.Service) *Handler {
	return &Handler{
		service: service,
	}
}

func NewServer(s *service.Service) *http.Server {
	utils.CheckEnvs(serverAddrName)

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
