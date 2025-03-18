package task_repo

import (
	"database/sql"
	"encoding/json"
	"errors"
	"face-track/internal/pkg/clients/face_cloud_client"
	"face-track/internal/pkg/model/task_model"
	"face-track/tools"
	"fmt"
	"image"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
)

const (
	foldersAmount            = 30000 // limit number of nested folders
	faceCloudApiUrlEnvName   = "FACE_CLOUD__API_URL"
	faceCloudUserEnvName     = "FACE_CLOUD__API_USER"
	faceCloudPasswordEnvName = "FACE_CLOUD__API_PASS"
)

type TaskRepo struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) (repo *TaskRepo) {
	return &TaskRepo{
		db: db,
	}
}

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

func (r *TaskRepo) SaveImageDisk(taskId int, image image.Image, imageName string) (imageRow *task_model.Image, err error) {

	imageRow = &task_model.Image{
		TaskId:    taskId,
		ImageName: imageName,
	}

	path := r.getImagePath(imageRow)

	if err := tools.SaveImg(image, path); err != nil {
		return nil, err
	}

	return imageRow, nil
}

func (r *TaskRepo) getImagePath(imageRow *task_model.Image) (path string) {

	subFolderID := imageRow.TaskId % foldersAmount
	folderToSave := fmt.Sprintf("/face track/images/%d/%d", subFolderID, imageRow.TaskId)
	tools.CreateFolderIfNotExist(folderToSave)

	return fmt.Sprintf("%s/%s.jpeg", folderToSave, imageRow.ImageName)
}

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

func (r *TaskRepo) DecodeFile(fileData *task_model.FileData) (img image.Image, err error) {

	img, _, err = image.Decode(fileData.File)
	if err != nil {
		return nil, err
	}

	return img, err
}

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
func (r *TaskRepo) GetFaceDetectionData(image *task_model.Image, token string) (imageData *task_model.FaceCloudDetectResponse, err error) {

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
	reqBody := task_model.FaceCloudLoginRequest{
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

	var response task_model.FaceCloudLoginResponse

	// process response data
	if err = json.Unmarshal(data, &response); err != nil {
		return token, err
	}

	return response.Data.AccessToken, err
}

func (r *TaskRepo) SaveProcessedData(processedFaces []*task_model.Face, processedImages []*task_model.Image) {

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

		_, err := r.db.NamedExec(query, processedFaces)
		if err != nil {
			panic(err)
		}
	}

	if len(processedImages) > 0 {
		for _, image := range processedImages {
			query := `UPDATE task_image 
					SET done=true 
					WHERE id=($1)`

			r.db.Exec(query, image.Id)
		}
	}
}

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
