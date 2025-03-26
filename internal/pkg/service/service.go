// Package service provides the service layer for managing tasks in the face tracking system.
// It includes methods for interacting with task-related functionality, such as creating, deleting,
// updating tasks, processing tasks, and adding images to tasks.
package service

import (
	"face-track/internal/pkg/database"
	"face-track/internal/pkg/model/task_model"
	"face-track/internal/pkg/repo"
	"face-track/internal/pkg/service/task_service"
	"face-track/tools"
	"log"
	"os"
)

const (
	// pgDbEnvName is the env variable key for the PostgreSQL database name.
	pgDbEnvName = "FACE_TRACK__PG_NAME"

	// pgDbUserName is the env variable key for the PostgreSQL database username.
	pgDbUserName = "FACE_TRACK__PG_USER"

	// pgPassEnvName is the env variable key for the PostgreSQL database password.
	pgPassEnvName = "FACE_TRACK__PG_PASS"
)

// Service is a struct that embeds the Task interface and provides methods to interact
// with task-related functionalities.
type Service struct {
	Task
}

// NewServiceWithRepo creates a new instance of Service, initializing it with the task service
// and connecting to the database using environment variables for PostgreSQL credentials.
func NewServiceWithRepo() (srvs *Service) {
	tools.CheckEnvs(pgDbEnvName, pgDbUserName, pgPassEnvName)

	db, err := database.GetDatabase(os.Getenv(pgDbEnvName), os.Getenv(pgDbUserName), os.Getenv(pgPassEnvName))
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}

	repo := repo.NewRepo(db)

	return &Service{
		Task: task_service.New(repo),
	}
}

// Task defines the interface for interacting with task-related functionalities.
type Task interface {
	GetTaskById(taskId int) (task *task_model.Task, err error)
	CreateTask() (taskId int, err error)
	DeleteTask(taskId int) (err error)
	AddImageToTask(taskId int, fileData *task_model.FileData) (err error)
	UpdateTaskStatus(taskId int, status string) (err error)
	ProcessTask(taskId int)
}
