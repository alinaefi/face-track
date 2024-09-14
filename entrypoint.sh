#!/bin/sh

echo "Waiting for PostgreSQL to be ready..."
until nc -z -v -w30 ${FACE_TRACK__PG_HOST} ${FACE_TRACK__PG_PORT}; do
  echo "Waiting for PostgreSQL..."
  sleep 1
done

echo "Running migrations..."
migrate -path=internal/database/migrations -database "postgresql://${FACE_TRACK__PG_USER}:${FACE_TRACK__PG_PASS}@${FACE_TRACK__PG_HOST}:${FACE_TRACK__PG_PORT}/${FACE_TRACK__PG_NAME}?sslmode=disable" up

echo "Starting Go application..."
exec "$@"
