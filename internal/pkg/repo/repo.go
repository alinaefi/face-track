package repo

import (
	"face-track/internal/pkg/model"
	"face-track/internal/pkg/repo/task_repo"
	"image"

	"github.com/jmoiron/sqlx"
)

type Repo struct {
	Task
}

func NewRepo(db *sqlx.DB) *Repo {
	return &Repo{
		Task: task_repo.New(db),
	}
}

type Task interface {
	GetTaskById(taskId int) (taskRow *model.Task, err error)
	GetTaskImages(taskId int) (images []*model.Image, err error)
	GetFacesByImageIds(imageIds []int) (taskFaces map[int][]*model.Face, err error)
	CreateTask() (taskId int, err error)
	DeleteTask(taskId int) (err error)
	SaveImageDisk(taskId int, image image.Image, imageName string) (imageRow *model.Image, err error)
	CreateImage(image *model.Image) (err error)
	DecodeFile(fileData *model.FileData) (img image.Image, err error)
	UpdateTaskStatus(taskId int, status string) (err error)
	GetFaceDetectionData(image *model.Image, token string) (imageData *model.FaceCloudDetectResponse, err error)
	GetFaceCloudToken() (token string, err error)
	SaveProcessedData(processedFaces []*model.Face, processedImages []*model.Image)
	UpdateTaskStatistics(task *model.Task) (err error)
}
