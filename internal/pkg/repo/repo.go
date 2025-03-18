// Package repo provides functionality to interact with task-related data in the database.
// It includes methods for creating, updating, deleting, and fetching tasks, images, and face detection data.
package repo

import (
	"face-track/internal/pkg/model/face_cloud_model"
	"face-track/internal/pkg/model/task_model"
	"face-track/internal/pkg/repo/task_repo"
	"image"

	"github.com/jmoiron/sqlx"
)

// Repo is a struct that embeds the Task interface and allows interaction with task-related functions.
type Repo struct {
	Task
}

// NewRepo creates a new instance of Repo, initializing it with the TaskRepo interface implementation.
func NewRepo(db *sqlx.DB) *Repo {
	return &Repo{
		Task: task_repo.New(db),
	}
}

// Task defines the interface for interacting with task-related functions.
type Task interface {
	GetTaskById(taskId int) (taskRow *task_model.Task, err error)
	GetTaskImages(taskId int) (images []*task_model.Image, err error)
	GetFacesByImageIds(imageIds []int) (taskFaces map[int][]*task_model.Face, err error)
	CreateTask() (taskId int, err error)
	DeleteTask(taskId int) (err error)
	SaveImageDisk(taskId int, image image.Image, imageName string) (imageRow *task_model.Image, err error)
	CreateImage(image *task_model.Image) (err error)
	DecodeFile(fileData *task_model.FileData) (img image.Image, err error)
	UpdateTaskStatus(taskId int, status string) (err error)
	GetFaceDetectionData(image *task_model.Image, token string) (imageData *face_cloud_model.FaceCloudDetectResponse, err error)
	GetFaceCloudToken() (token string, err error)
	SaveProcessedData(processedFaces []*task_model.Face, processedImages []*task_model.Image)
	UpdateTaskStatistics(task *task_model.Task) (err error)
}
