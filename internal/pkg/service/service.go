package service

import (
	"face-track/internal/pkg/database"
	"face-track/internal/pkg/repo"
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

type Task interface{}
