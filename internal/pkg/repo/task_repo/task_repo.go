// Package task_repo provides methods for managing tasks related data in the database, and interacting
// with the Face Cloud API for face detection and processing.
package task_repo

import (
	"database/sql"
	"encoding/json"
	"errors"
	"face-track/internal/pkg/clients/face_cloud_client"
	"face-track/internal/pkg/model/face_cloud_model"
	"face-track/internal/pkg/model/task_model"
	"face-track/tools"
	"fmt"
	"image"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/jmoiron/sqlx"
)

const (
	// foldersAmount defines the maximum number of nested folders for organizing images.
	foldersAmount = 30000

	// faceCloudApiUrlEnvName is the env variable key for the Face Cloud API URL.
	faceCloudApiUrlEnvName = "FACE_CLOUD__API_URL"

	// faceCloudUserEnvName is the env variable key for the Face Cloud API user's email.
	faceCloudUserEnvName = "FACE_CLOUD__API_USER"

	// faceCloudPasswordEnvName is the env variable key for the Face Cloud API user's password.
	faceCloudPasswordEnvName = "FACE_CLOUD__API_PASS"
)

// TaskRepo represents a repository for managing tasks and interacting with the database.
// It provides methods for CRUD operations on tasks, image management, and communication with the Face Cloud API.
type TaskRepo struct {
	db *sqlx.DB
}

// New creates a new TaskRepo instance with the provided database connection.
func New(db *sqlx.DB) (repo *TaskRepo) {
	return &TaskRepo{
		db: db,
	}
}

// GetTaskById retrieves a task by its ID from the database.
func (r *TaskRepo) GetTaskById(taskId int) (task *task_model.Task, err error) {
	task = &task_model.Task{}

	query := `SELECT 
				id, 
				task_status, 
				faces_total, 
				faces_female, 
				faces_male, 
				age_female_avg, 
				age_male_avg 
			FROM task 
			WHERE id=$1`

	if err = r.db.QueryRow(query, taskId).Scan(
		&task.Id,
		&task.Status,
		&task.Statistics.FacesTotal,
		&task.Statistics.FacesFemale,
		&task.Statistics.FacesMale,
		&task.Statistics.AgeFemaleAvg,
		&task.Statistics.AgeMaleAvg,
	); err != nil {
		return nil, err
	}

	return task, err
}

// GetTaskImages retrieves all images associated with a given task ID from the database.
func (r *TaskRepo) GetTaskImages(taskId int) (images []*task_model.Image, err error) {

	query := `SELECT 
				id, 
				task_id, 
				image_name, 
				done 
			FROM task_image 
			WHERE task_id=$1`

	if err = r.db.Select(&images, query, taskId); err != nil {
		return nil, err
	}

	return images, err
}

// GetFacesByImageIds retrieves faces associated with the given image IDs.
func (r *TaskRepo) GetFacesByImageIds(imageIds []int) (taskFaces map[int][]*task_model.Face, err error) {
	var rows *sqlx.Rows
	var inArgs []interface{}
	taskFaces = make(map[int][]*task_model.Face)

	query := `SELECT 
				id, 
				image_id, 
				gender, 
				age, 
				bbox_height, 
				bbox_width, 
				bbox_x, 
				bbox_y 
			FROM face 
			WHERE image_id IN (?)`

	query, inArgs, err = sqlx.In(query, imageIds)
	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query)

	rows, err = r.db.Queryx(query, inArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var face task_model.Face
		if err := rows.Scan(
			&face.Id,
			&face.ImageId,
			&face.Gender,
			&face.Age,
			&face.Bbox.Height,
			&face.Bbox.Width,
			&face.Bbox.X,
			&face.Bbox.Y,
		); err != nil {
			return nil, err
		}
		taskFaces[face.ImageId] = append(taskFaces[face.ImageId], &face)
	}

	if rows.Err() != nil {
		return nil, err
	}

	return taskFaces, err
}

// CreateTask creates a new task and returns the task ID.
func (r *TaskRepo) CreateTask() (taskId int, err error) {

	query := `INSERT INTO task 
				(
				task_status, 
				faces_total, 
				faces_female, 
				faces_male, 
				age_female_avg, 
				age_male_avg
				) 
			VALUES ('new', 0, 0, 0, 0, 0) 
			RETURNING id`

	row := r.db.QueryRowx(query)
	if err = row.Scan(&taskId); err != nil {
		return 0, err
	}

	return taskId, err
}

// DeleteTask deletes a task by its ID from the database.
func (r *TaskRepo) DeleteTask(taskId int) (err error) {
	var result sql.Result
	var rowsDeleted int64

	query := `DELETE FROM task WHERE id=($1)`

	result, err = r.db.Exec(query, taskId)
	if err != nil {
		return err
	}

	rowsDeleted, err = result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsDeleted == 0 {
		return errors.New("operation unsuccessful: row not found")
	}

	return err
}

// SaveImageDisk saves an image to disk and returns an image record with task ID and image name.
func (r *TaskRepo) SaveImageDisk(taskId int, image image.Image, imageName string) (imageRow *task_model.Image, err error) {

	// Create unique file name
	uniqueFileName := getUniqueFilename(imageName)

	imageRow = &task_model.Image{
		TaskId:    taskId,
		ImageName: uniqueFileName,
	}

	path := r.getImagePath(imageRow)

	if err := tools.SaveImg(image, path); err != nil {
		return nil, err
	}

	return imageRow, nil
}

func (r *TaskRepo) getImagePath(imageRow *task_model.Image) (path string) {

	homeDir, _ := os.UserHomeDir() // Get the home directory
	subFolderID := imageRow.TaskId % foldersAmount
	folderToSave := fmt.Sprintf("%s/face-track/images/%d/%d", homeDir, subFolderID, imageRow.TaskId)

	tools.CreateFolderIfNotExist(folderToSave) // Ensure folder exists

	return fmt.Sprintf("%s/%s", folderToSave, imageRow.ImageName)
}

func getUniqueFilename(filename string) string {

	ext := filepath.Ext(filename)             // Get file extension
	name := filename[:len(filename)-len(ext)] // Get name without extension
	timestamp := time.Now().UnixNano()        // Use nanoseconds to avoid collisions

	return fmt.Sprintf("%s_%d%s", name, timestamp, ext)
}

// CreateImage inserts a new image record into the task_image table.
func (r *TaskRepo) CreateImage(image *task_model.Image) (err error) {
	var result sql.Result
	var rowsAffected int64

	query := `INSERT INTO task_image 
				(
				task_id, 
				image_name
				) 
			VALUES ($1, $2)`

	result, err = r.db.Exec(query, image.TaskId, image.ImageName)
	if err != nil {
		return err
	}

	rowsAffected, err = result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("operation unsuccessful: row not found")
	}

	return err
}

// DecodeFile decodes the image file from the provided file data and returns the decoded image.
func (r *TaskRepo) DecodeFile(fileData *task_model.FileData) (img image.Image, err error) {

	img, _, err = image.Decode(fileData.File)
	if err != nil {
		return nil, err
	}

	return img, err
}

// UpdateTaskStatus updates the status of a task by its ID.
func (r *TaskRepo) UpdateTaskStatus(taskId int, status string) (err error) {
	var result sql.Result
	var rowsAffected int64

	query := `UPDATE task 
				SET task_status=$1 
				WHERE id=$2`

	result, err = r.db.Exec(query, status, taskId)
	if err != nil {
		return err
	}

	rowsAffected, err = result.RowsAffected()
	if err != nil {
		log.Println(err)
	}

	if rowsAffected == 0 {
		return errors.New("operation unsuccessful: row not found")
	}

	return err
}

// GetFaceDetectionData requests Face Cloud API to detect faces on the specified image and returns response or error.
func (r *TaskRepo) GetFaceDetectionData(image *task_model.Image, token string) (imageData *face_cloud_model.FaceCloudDetectResponse, err error) {

	// prepare image
	imagePath := r.getImagePath(image)
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// send request
	data, err := face_cloud_client.DetectFaces(file, token)
	if err != nil {
		return nil, err
	}

	// process response data
	if err = json.Unmarshal(data, &imageData); err != nil {
		return nil, err
	}

	return imageData, err
}

// GetFaceCloudToken requests login to Face Cloud and returns JWT access token or error.
func (r *TaskRepo) GetFaceCloudToken() (token string, err error) {

	// prepare request params
	tools.CheckEnvs(faceCloudApiUrlEnvName, faceCloudUserEnvName, faceCloudPasswordEnvName)
	reqBody := face_cloud_model.FaceCloudLoginRequest{
		Email:    os.Getenv(faceCloudUserEnvName),
		Password: os.Getenv(faceCloudPasswordEnvName),
	}

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return token, err
	}

	// send request
	data, err := face_cloud_client.Login(reqBodyBytes)
	if err != nil {
		return token, err
	}

	var response face_cloud_model.FaceCloudLoginResponse

	// process response data
	if err = json.Unmarshal(data, &response); err != nil {
		return token, err
	}

	return response.Data.AccessToken, err
}

// SaveProcessedData saves processed face data and marks images as "done" in the database.
func (r *TaskRepo) SaveProcessedData(processedFaces []*task_model.Face, processedImages []*task_model.Image) {
	var err error

	if len(processedFaces) > 0 {
		query := `INSERT INTO face 
						(
						image_id, 
						gender, 
						age, 
						bbox_height, 
						bbox_width, 
						bbox_x, 
						bbox_y
						) 
					VALUES 
						(
						:image_id, 
						:gender, 
						:age, 
						:bbox_height, 
						:bbox_width, 
						:bbox_x, 
						:bbox_y
					)`

		_, err = r.db.NamedExec(query, processedFaces)
		if err != nil {
			panic(err)
		}
	}

	if len(processedImages) > 0 {
		for _, image := range processedImages {
			query := `UPDATE task_image 
					SET done=true 
					WHERE id=($1)`

			_, err = r.db.Exec(query, image.Id)
			if err != nil {
				panic(err)
			}
		}
	}
}

// UpdateTaskStatistics updates the statistics for a task, including gender and age data.
func (r *TaskRepo) UpdateTaskStatistics(task *task_model.Task) (err error) {
	var result sql.Result
	var rowsAffected int64

	query := `
		UPDATE task 
		SET task_status = :task_status, 
		    faces_total = :faces_total, 
		    faces_male = :faces_male, 
		    faces_female = :faces_female, 
		    age_female_avg = :age_female_avg, 
		    age_male_avg = :age_male_avg 
		WHERE id = :id`

	result, err = r.db.NamedExec(query, task)
	if err != nil {
		return err
	}

	rowsAffected, err = result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("operation unsuccessful: row not found")
	}

	return err
}
