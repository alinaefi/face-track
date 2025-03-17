package service

import (
	"face-track/internal/pkg/database"
	"face-track/internal/pkg/model"
	"face-track/internal/pkg/repo"
	"face-track/internal/pkg/service/task_service"
	"face-track/utils"
	"log"
	"os"
)

const (
	pgDbEnvName   = "FACE_TRACK__PG_NAME"
	pgDbUserName  = "FACE_TRACK__PG_USER"
	pgPassEnvName = "FACE_TRACK__PG_PASS"
)

type Service struct {
	Task
}

func NewServiceWithRepo() (srvs *Service) {
	utils.CheckEnvs(pgDbEnvName, pgDbUserName, pgPassEnvName)

	db, err := database.GetDatabase(os.Getenv(pgDbEnvName), os.Getenv(pgDbUserName), os.Getenv(pgPassEnvName))
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}

	repo := repo.NewRepo(db)

	return &Service{
		Task: task_service.New(repo),
	}
}

type Task interface {
	GetTaskById(taskId int) (task *model.Task, err error)
	CreateTask() (resp *task_service.Response)
	DeleteTask(taskId int) (resp *task_service.Response)
	AddImageToTask(taskId int, imageName string, fileData *model.FileData) (resp *task_service.Response)
	UpdateTaskStatus(taskId int, status string) (err error)
	ProcessTask(taskId int)
}
