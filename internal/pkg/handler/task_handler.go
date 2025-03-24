package handler

import (
	"face-track/internal/pkg/middleware"
	"face-track/internal/pkg/model/task_model"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *Handler) setTaskGroup(api *gin.RouterGroup) {
	taskApiGroup := api.Group("tasks")
	authMiddleware := middleware.NewAuthMiddleware()
	taskApiGroup.Use(authMiddleware.BasicAuthMiddleware())
	{
		taskApiGroup.GET("/:id", h.getTask)
		taskApiGroup.POST("/", h.createTask)
		taskApiGroup.DELETE("/:id", h.deleteTask)
		taskApiGroup.PATCH("/:id", h.addImageToTask)
		taskApiGroup.PATCH("/:id/process", h.processTask)
	}
}

func (h *Handler) getTask(c *gin.Context) {

	var taskId int
	var err error
	var task *task_model.Task

	taskIdStr := c.Param("id")

	taskId, err = strconv.Atoi(taskIdStr)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err = h.service.GetTaskById(taskId)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": task})
}

func (h *Handler) createTask(c *gin.Context) {

	taskId, err := h.service.CreateTask()
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": taskId})
}

func (h *Handler) deleteTask(c *gin.Context) {

	var taskId int
	var err error

	taskIdStr := c.Param("id")

	taskId, err = strconv.Atoi(taskIdStr)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.service.DeleteTask(taskId)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": "task was successfully deleted"})
}

func (h *Handler) addImageToTask(c *gin.Context) {
	var req *task_model.TaskIdRequest
	var err error

	err = c.BindJSON(&req)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	imageName := c.PostForm("imageName")
	if len(imageName) == 0 {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "image name cannot be empty"})
		return
	}

	fileData := &task_model.FileData{}
	fileData.File, fileData.FileHeader, err = c.Request.FormFile("image")
	defer fileData.File.Close()

	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to get uploaded image"})
		return
	}

	err = h.service.AddImageToTask(req.TaskId, imageName, fileData)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": "image was successfully added to task"})
}

func (h *Handler) processTask(c *gin.Context) {
	var req *task_model.TaskIdRequest
	var err error

	err = c.BindJSON(&req)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.service.UpdateTaskStatus(req.TaskId, "in_progress")
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": "task is being processed"})

	h.service.ProcessTask(req.TaskId)
}
