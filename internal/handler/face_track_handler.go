package handler

import (
	"face-track/internal/model"
	"face-track/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func respond(c *gin.Context, resp *service.Response) {
	c.JSON(resp.Status, resp.Data)
}

func (h *Handler) HandleGetTask(c *gin.Context) {

	taskIdStr := c.Param("id")
	taskId, err := strconv.Atoi(taskIdStr)
	if err != nil {
		respond(c, &service.Response{
			Status: http.StatusBadRequest,
			Data:   gin.H{"error": "invalid task id format"},
		})
		return
	}

	resp := h.service.GetTaskById(taskId)

	respond(c, resp)
}

func (h *Handler) HandleCreateTask(c *gin.Context) {

	resp := h.service.CreateTask()

	respond(c, resp)
}

func (h *Handler) HandleDeleteTask(c *gin.Context) {

	taskIdStr := c.Param("id")
	taskId, err := strconv.Atoi(taskIdStr)
	if err != nil {
		respond(c, &service.Response{
			Status: http.StatusBadRequest,
			Data:   gin.H{"error": "invalid task id format"},
		})
		return
	}

	resp := h.service.DeleteTask(taskId)

	respond(c, resp)
}

func (h *Handler) HandleAddImageToTask(c *gin.Context) {

	taskIdStr := c.Param("id")
	taskId, err := strconv.Atoi(taskIdStr)
	if err != nil {
		respond(c, &service.Response{
			Status: http.StatusBadRequest,
			Data:   gin.H{"error": "invalid task id format"},
		})
		return
	}

	imageName := c.PostForm("imageName")
	if len(imageName) == 0 {
		respond(c, &service.Response{
			Status: http.StatusBadRequest,
			Data:   gin.H{"error": "image name cannot be empty"},
		})
		return
	}

	fileData := &model.FileData{}
	fileData.File, fileData.FileHeader, err = c.Request.FormFile("image")
	defer fileData.File.Close()

	if err != nil {
		respond(c, &service.Response{
			Status: http.StatusBadRequest,
			Data:   gin.H{"error": "failed to get uploaded image"},
		})
		return
	}

	resp := h.service.AddImageToTask(taskId, imageName, fileData)

	respond(c, resp)
}

func (h *Handler) HandleProcessTask(c *gin.Context) {

	taskIdStr := c.Param("id")
	taskId, err := strconv.Atoi(taskIdStr)
	if err != nil {
		respond(c, &service.Response{
			Status: http.StatusBadRequest,
			Data:   gin.H{"error": "invalid task id format"},
		})
		return
	}

	if err = h.service.UpdateTaskStatus(taskId, "in_progress"); err != nil {
		respond(c, &service.Response{
			Status: http.StatusInternalServerError,
			Data:   gin.H{"error": err.Error()},
		})
		return
	}

	respond(c, &service.Response{
		Status: http.StatusOK,
		Data:   gin.H{"message": "task is being processed"},
	})

	go func() {
		h.service.ProcessTask(taskId)
	}()
}
