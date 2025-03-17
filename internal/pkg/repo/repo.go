package repo

import (
	"face-track/internal/pkg/repo/task_repo"

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
}
