package task_repo

import (
	"database/sql"
	"encoding/json"
	"face-track/internal/pkg/model"
	"face-track/utils"
	"fmt"
	"image"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
)

const (
	foldersAmount            = 30000 // не храним в одной папке более 32000 файлов/папок
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

// GetTaskById делает запрос к бд и возвращает данные о задании по идентификатору задания
// или ошибку
func (r *TaskRepo) GetTaskById(taskId int) (taskRow *model.Task, err error) {

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

	var task model.Task

	err = r.db.QueryRow(query, taskId).Scan(
		&task.Id,
		&task.Status,
		&task.Statistics.FacesTotal,
		&task.Statistics.FacesFemale,
		&task.Statistics.FacesMale,
		&task.Statistics.AgeFemaleAvg,
		&task.Statistics.AgeMaleAvg,
	)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// GetTaskImages делает запрос к бд и возвращает данные об изображениях задания по идентификатору задания
// или ошибку
func (r *TaskRepo) GetTaskImages(taskId int) (images []*model.Image, err error) {

	query := `SELECT 
				id, 
				task_id, 
				image_name, 
				done 
			FROM task_image 
			WHERE task_id=$1`

	if err = r.db.Select(&images, query, taskId); err != nil {
		return images, err
	}

	return images, nil
}

// GetFacesByImageIds делает запрос к бд и возвращает данные о лицах на изображениях задания по идентификаторам изображений
// или ошибку
func (r *TaskRepo) GetFacesByImageIds(imageIds []int) (taskFaces map[int][]*model.Face, err error) {

	taskFaces = make(map[int][]*model.Face)
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

	query, inArgs, err := sqlx.In(query, imageIds)
	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query)

	rows, err := r.db.Queryx(query, inArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var face model.Face
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

	return taskFaces, nil
}

// CreateTask осуществляет вставку нового задания в бд и возвращает идентификатор задания
// или ошибку
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

	return taskId, nil
}

// DeleteTask удаляет из бд задание и все связанные с ним данные об изображениях и лицах
// возвращает ошибку, если удаление не удалось
func (r *TaskRepo) DeleteTask(taskId int) (err error) {

	query := `DELETE FROM task WHERE id=($1)`

	result, err := r.db.Exec(query, taskId)
	if err != nil {
		return err
	}

	rowsDeleted, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsDeleted == 0 {
		return err
	}

	return nil
}

// SaveImageDisk сохраняет изображение в файловой системе и возвращает его или ошибку
func (r *TaskRepo) SaveImageDisk(taskId int, image image.Image, imageName string) (imageRow *model.Image, err error) {

	imageRow = &model.Image{
		TaskId:    taskId,
		ImageName: imageName,
	}

	path := r.getImagePath(imageRow)

	if err := utils.SaveImg(image, path); err != nil {
		return nil, err
	}

	return imageRow, nil
}

// getImagePath создает папки для изображений задания и возвращает путь к изображению на диске
func (r *TaskRepo) getImagePath(imageRow *model.Image) (path string) {

	subFolderID := imageRow.TaskId % foldersAmount
	folderToSave := fmt.Sprintf("/face track/images/%d/%d", subFolderID, imageRow.TaskId)
	utils.CreateFolderIfNotExist(folderToSave)

	return fmt.Sprintf("%s/%s.jpeg", folderToSave, imageRow.ImageName)
}

// CreateImage сохраняет данные об изображении в бд, возвращает ошибку в случае неудачного запроса
func (r *TaskRepo) CreateImage(image *model.Image) (err error) {

	query := `INSERT INTO task_image 
				(
				task_id, 
				image_name
				) 
			VALUES ($1, $2)`

	result, err := r.db.Exec(query, image.TaskId, image.ImageName)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("failed to insert image data")
	}

	return nil
}

// DecodeFile декодирует объект типа файл в объект типа Image
// возвращает объект типа Image или ошибку
func (r *TaskRepo) DecodeFile(fileData *model.FileData) (img image.Image, err error) {

	img, _, err = image.Decode(fileData.File)
	if err != nil {
		return nil, err
	}

	return img, nil
}

// UpdateTaskStatus обновляет статус задания
func (r *TaskRepo) UpdateTaskStatus(taskId int, status string) (err error) {

	query := `UPDATE task 
				SET task_status=$1 
				WHERE id=$2`

	result, err := r.db.Exec(query, status, taskId)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Println(err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// GetFaceDetectionData отправляет запрос к API Face Cloud и возвращает ответ с API или ошибку
func (r *TaskRepo) GetFaceDetectionData(image *model.Image, token string) (imageData *model.FaceCloudDetectResponse, err error) {

	// готовим изображение
	imagePath := r.getImagePath(image)
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// готовим и отправляем запрос
	data, err := utils.FaceCloudDetectRequest(file, token)
	if err != nil {
		return nil, err
	}

	// распаковываем ответ с API
	if err = json.Unmarshal(data, &imageData); err != nil {
		return nil, err
	}

	return imageData, nil
}

// GetFaceCloudToken отправляет запрос к API Face Cloud на получение токена авторизации
// возвращает токен или ошибку
func (r *TaskRepo) GetFaceCloudToken() (token string, err error) {

	// готовим параметры запроса
	utils.CheckEnvs(faceCloudApiUrlEnvName, faceCloudUserEnvName, faceCloudPasswordEnvName)
	reqBody := model.FaceCloudLoginRequest{
		Email:    os.Getenv(faceCloudUserEnvName),
		Password: os.Getenv(faceCloudPasswordEnvName),
	}

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return token, err
	}

	// готовим и отправляем запрос
	data, err := utils.FaceCloudLoginRequest(reqBodyBytes)
	if err != nil {
		return token, err
	}

	var response model.FaceCloudLoginResponse

	// распаковываем ответ с API
	if err = json.Unmarshal(data, &response); err != nil {
		return token, err
	}

	return response.Data.AccessToken, nil
}

// SaveProcessedData сохраняет в бд данные об обработанных изображениях и лицах
func (r *TaskRepo) SaveProcessedData(processedFaces []*model.Face, processedImages []*model.Image) {

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

// UpdateTaskStatistics обновляет в бд данные статистики задания
// возвращает ошибку в случае неудачного запроса
func (r *TaskRepo) UpdateTaskStatistics(task *model.Task) (err error) {
	query := `
		UPDATE task 
		SET task_status = :task_status, 
		    faces_total = :faces_total, 
		    faces_male = :faces_male, 
		    faces_female = :faces_female, 
		    age_female_avg = :age_female_avg, 
		    age_male_avg = :age_male_avg 
		WHERE id = :id`

	res, err := r.db.NamedExec(query, task)
	if err != nil {
		log.Println(err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Println(err)
	}

	if rowsAffected == 0 {
		return err
	}

	return nil
}
