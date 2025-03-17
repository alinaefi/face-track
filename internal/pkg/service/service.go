package service

import (
	"face-track/internal/pkg/database"
	"face-track/internal/pkg/repository"
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
	repository *repository.FaceTrackRepo
}

func NewService() (srvs *Service) {
	utils.CheckEnvs(pgDbEnvName, pgDbUserName, pgPassEnvName)

	db, err := database.GetDatabase(os.Getenv(pgDbEnvName), os.Getenv(pgDbUserName), os.Getenv(pgPassEnvName))
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}

	repo := repository.NewFaceTrackRepo(db)

	return &Service{
		repository: repo,
	}
}
