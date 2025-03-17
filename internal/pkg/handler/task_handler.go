package handler

import (
	"face-track/internal/pkg/middleware"
	"face-track/internal/pkg/model"
	"face-track/internal/pkg/service/task_service"
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
		taskApiGroup.GET("/tasks/:id", h.getTask)
		taskApiGroup.POST("/tasks", h.createTask)
		taskApiGroup.DELETE("/tasks/:id", h.deleteTask)
		taskApiGroup.PATCH("/tasks/:id", h.addImageToTask)
		taskApiGroup.PATCH("/tasks/:id/process", h.processTask)
	}
}

func respond(c *gin.Context, resp *task_service.Response) {
	c.JSON(resp.Status, resp.Data)
}

func (h *Handler) getTask(c *gin.Context) {
	var err error
	var task *model.Task

	req := &struct {
		TaskId int `json:"id"`
	}{}

	err = c.BindJSON(&req)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err = h.service.GetTaskById(req.TaskId)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": task})
}

func (h *Handler) createTask(c *gin.Context) {

	resp := h.service.CreateTask()

	respond(c, resp)
}

func (h *Handler) deleteTask(c *gin.Context) {

	taskIdStr := c.Param("id")
	taskId, err := strconv.Atoi(taskIdStr)
	if err != nil {
		respond(c, &task_service.Response{
			Status: http.StatusBadRequest,
			Data:   gin.H{"error": "invalid task id format"},
		})
		return
	}

	resp := h.service.DeleteTask(taskId)

	respond(c, resp)
}

func (h *Handler) addImageToTask(c *gin.Context) {

	taskIdStr := c.Param("id")
	taskId, err := strconv.Atoi(taskIdStr)
	if err != nil {
		respond(c, &task_service.Response{
			Status: http.StatusBadRequest,
			Data:   gin.H{"error": "invalid task id format"},
		})
		return
	}

	imageName := c.PostForm("imageName")
	if len(imageName) == 0 {
		respond(c, &task_service.Response{
			Status: http.StatusBadRequest,
			Data:   gin.H{"error": "image name cannot be empty"},
		})
		return
	}

	fileData := &model.FileData{}
	fileData.File, fileData.FileHeader, err = c.Request.FormFile("image")
	defer fileData.File.Close()

	if err != nil {
		respond(c, &task_service.Response{
			Status: http.StatusBadRequest,
			Data:   gin.H{"error": "failed to get uploaded image"},
		})
		return
	}

	resp := h.service.AddImageToTask(taskId, imageName, fileData)

	respond(c, resp)
}

func (h *Handler) processTask(c *gin.Context) {

	taskIdStr := c.Param("id")
	taskId, err := strconv.Atoi(taskIdStr)
	if err != nil {
		respond(c, &task_service.Response{
			Status: http.StatusBadRequest,
			Data:   gin.H{"error": "invalid task id format"},
		})
		return
	}

	if err = h.service.UpdateTaskStatus(taskId, "in_progress"); err != nil {
		respond(c, &task_service.Response{
			Status: http.StatusInternalServerError,
			Data:   gin.H{"error": err.Error()},
		})
		return
	}

	respond(c, &task_service.Response{
		Status: http.StatusOK,
		Data:   gin.H{"message": "task is being processed"},
	})

	go func() {
		h.service.ProcessTask(taskId)
	}()
}
