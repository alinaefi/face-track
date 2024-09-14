package internal

import (
	"face-track/internal/handler"
	"face-track/internal/middleware"

	"github.com/gin-gonic/gin"
)

func setupRoutes(router *gin.Engine, h *handler.Handler) {
	apiGroup := router.Group("/api")

	{
		authMiddleware := middleware.NewAuthMiddleware()

		apiGroup.Use(authMiddleware.BasicAuthMiddleware())

		apiGroup.GET("/tasks/:id", func(c *gin.Context) {
			h.HandleGetTask(c)
		})
		apiGroup.POST("/tasks", func(c *gin.Context) {
			h.HandleCreateTask(c)
		})
		apiGroup.DELETE("/tasks/:id", func(c *gin.Context) {
			h.HandleDeleteTask(c)
		})
		apiGroup.PATCH("/tasks/:id", func(c *gin.Context) {
			h.HandleAddImageToTask(c)
		})
		apiGroup.PATCH("/tasks/:id/process", func(c *gin.Context) {
			h.HandleProcessTask(c)
		})
	}
}
