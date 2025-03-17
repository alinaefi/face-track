package database

import (
	"fmt"
	"os"
	"time"

	"face-track/tools"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const (
	pgHostEnvName = "FACE_TRACK__PG_HOST"
	pgPortEnvName = "FACE_TRACK__PG_PORT"
)

func GetDatabase(dbName, user, password string) (db *sqlx.DB, err error) {

	tools.CheckEnvs(pgHostEnvName, pgPortEnvName)

	connStr := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		os.Getenv(pgHostEnvName),
		os.Getenv(pgPortEnvName),
		dbName,
		user,
		password,
	)

	db, err = sqlx.Connect("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db.SetMaxIdleConns(10)
	db.SetConnMaxIdleTime(5 * time.Minute)

	return db, nil
}
