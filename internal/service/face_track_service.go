package service

import (
	"database/sql"
	"face-track/internal/model"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

const (
	foldersAmount = 30000
)

type Response struct {
	Status int
	Data   interface{}
}

// GetTaskById возвращает данные о задании, изображениях и лицах, связанных с ним по идентификатору задания
func (s *Service) GetTaskById(taskId int) (resp *Response) {
	resp = &Response{Status: http.StatusInternalServerError}

	task, err := s.getFullTaskData(taskId)
	if err != nil {
		if err == sql.ErrNoRows {
			resp.Status = http.StatusNotFound
			resp.Data = gin.H{"error": fmt.Sprintf("task with id %d not found", taskId)}
			return resp
		}
		resp.Data = gin.H{"error": "failed to get task"}
		return resp
	}

	resp.Status = http.StatusOK
	resp.Data = task
	return resp
}

// GetTaskById возвращает данные о задании в виде объекта
func (s *Service) getFullTaskData(taskId int) (task *model.Task, err error) {

	task, err = s.repository.GetTaskById(taskId)
	if err != nil {
		return nil, err
	}

	task.Images, err = s.repository.GetTaskImages(taskId)
	if err != nil {
		return nil, err
	}

	if len(task.Images) > 0 {
		imageIds := make([]int, len(task.Images))
		for _, img := range task.Images {
			imageIds = append(imageIds, img.Id)
		}

		faces, err := s.repository.GetFacesByImageIds(imageIds)
		if err != nil {
			return nil, err
		}

		for _, image := range task.Images {
			image.Faces = faces[image.Id]
		}
	}

	return task, nil
}

// CreateTask создает новое задание
func (s *Service) CreateTask() (resp *Response) {
	resp = &Response{Status: http.StatusInternalServerError, Data: gin.H{"error": "failed to create a task"}}

	taskId, err := s.repository.CreateTask()
	if err != nil {
		return resp
	}

	if taskId == 0 {
		return resp
	}

	resp.Status = http.StatusOK
	resp.Data = gin.H{"taskId": taskId}
	return resp
}

// DeleteTask удаляет все данные о задании с диска и из бд
func (s *Service) DeleteTask(taskId int) (resp *Response) {
	resp = &Response{Status: http.StatusInternalServerError, Data: gin.H{"error": "failed to delete the task"}}

	task, err := s.repository.GetTaskById(taskId)
	if err != nil {
		if err == sql.ErrNoRows {
			resp.Status = http.StatusNotFound
			resp.Data = gin.H{"error": fmt.Sprintf("task with id %d not found", taskId)}
			return resp
		}
		return resp
	}

	if task.Status == "in_progress" {
		resp.Status = http.StatusBadRequest
		resp.Data = gin.H{"error": "unable to delete task: processing is in progress"}
		return resp
	}

	if err = s.repository.DeleteTask(taskId); err != nil {
		return resp
	}

	if err = s.deleteTaskImagesFromDisk(task.Id); err != nil {
		return resp
	}

	resp.Status = http.StatusOK
	resp.Data = gin.H{"message": "task was successfully deleted"}
	return resp
}

// deleteTaskImagesFromDisk удаляет папку с изображениями задания с диска
func (s *Service) deleteTaskImagesFromDisk(taskId int) (err error) {

	subFolderID := taskId % foldersAmount
	path := fmt.Sprintf("/face track/images/%d/%d", subFolderID, taskId)

	if err = os.RemoveAll(path); err != nil {
		return err
	}

	return nil
}

// AddImageToTask добавляет изображение в бд и на диск
func (s *Service) AddImageToTask(taskId int, imageName string, fileData *model.FileData) (resp *Response) {
	resp = &Response{Status: http.StatusInternalServerError, Data: gin.H{"error": "failed to add image to task"}}

	task, err := s.getFullTaskData(taskId)
	if err != nil {
		if err == sql.ErrNoRows {
			resp.Status = http.StatusNotFound
			resp.Data = gin.H{"error": fmt.Sprintf("task with id %d not found", taskId)}
			return resp
		}
		return resp
	}

	// проверяем статус задания
	if task.Status != "new" {
		resp.Status = http.StatusBadRequest
		resp.Data = gin.H{"error": "failed to add image to task: task processing in progress"}
		return resp
	}

	// проверяем уникальность имени
	for _, image := range task.Images {
		if image.ImageName == imageName {
			resp.Status = http.StatusBadRequest
			resp.Data = gin.H{"error": "failed to add image to task: image with specified name already exists"}
			return resp
		}
	}

	// валидируем расширение
	if fileData.FileHeader.Header.Get("Content-Type") != "image/jpeg" {
		resp.Status = http.StatusBadRequest
		resp.Data = gin.H{"error": "unsupported file extension"}
		return resp
	}

	// декодируем файл в изображение
	image, err := s.repository.DecodeFile(fileData)
	if err != nil {
		return resp
	}

	// сохраняем файл на диске
	imageRow, err := s.repository.SaveImageDisk(taskId, image, imageName)
	if err != nil {
		return resp
	}

	if err = s.repository.CreateImage(imageRow); err != nil {
		return resp
	}

	resp.Status = http.StatusOK
	resp.Data = gin.H{"message": "image was successfully added to task"}
	return resp
}

// UpdateTaskStatus обновляет статус задания на заданный
func (s *Service) UpdateTaskStatus(taskId int, status string) (err error) {
	errorMsg := fmt.Errorf("failed update task status")

	if err = s.repository.UpdateTaskStatus(taskId, status); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("task with id %d not found", taskId)
		}
		return errorMsg
	}

	return nil
}

// ProcessTask запускает параллельную обработку изображений задания
func (s *Service) ProcessTask(taskId int) {

	// запрашиваем все данные о задании, его изображениях и лицах
	task, err := s.getFullTaskData(taskId)
	if err != nil {
		log.Println(err)
		_ = s.repository.UpdateTaskStatus(taskId, "error")
		return
	}
	if task.Status == "completed" {
		return
	}

	g := new(errgroup.Group)
	g.SetLimit(10)

	var Mu sync.RWMutex
	var facesToSave []*model.Face
	var imagesToSetDone []*model.Image

	if len(task.Images) > 0 {

		// получаем токен для запросов к внешнему API
		token, err := s.repository.GetFaceCloudToken()
		if err != nil {
			log.Println(err)
			_ = s.repository.UpdateTaskStatus(taskId, "error")
			return
		}

		for _, img := range task.Images {
			// не обрабатываем изображения повторно
			if img.DoneFlag {
				continue
			}

			currImage := img
			g.Go(func() error {

				// отправляет запрос к face cloud
				imageData, err := s.repository.GetFaceDetectionData(currImage, token)
				if err != nil {
					log.Println(err)
					return err
				}

				// готовим данные о найденных лицах
				for _, faceData := range imageData.Data {
					newFace := &model.Face{
						ImageId: currImage.Id,
						Gender:  faceData.Demographics.Gender,
						Age:     int(faceData.Demographics.Age.Mean),
						Height:  faceData.Bbox.Height,
						Width:   faceData.Bbox.Width,
						X:       faceData.Bbox.X,
						Y:       faceData.Bbox.Y,
					}

					Mu.Lock()
					facesToSave = append(facesToSave, newFace)
					Mu.Unlock()
				}

				Mu.Lock()
				imagesToSetDone = append(imagesToSetDone, currImage)
				Mu.Unlock()

				return err
			})
		}
	}
	err = g.Wait()

	// сохраняем успешно обработанные данные в бд даже в случае ошибки
	s.repository.SaveProcessedData(facesToSave, imagesToSetDone)

	if err != nil {
		log.Println(err)
		_ = s.repository.UpdateTaskStatus(taskId, "error")
		return
	}

	// запрашиваем обновленные данные о задании
	task, _ = s.getFullTaskData(taskId)

	// подсчитываем статистику задания и вызываем сохранение финальных данных
	s.concludeTask(task)
}

// concludeTask подсчитывает статистические данные задания и сохраняет в бд
func (s *Service) concludeTask(task *model.Task) {

	var totalFaces, maleFaces, femaleFaces, totalMaleAge, totalFemaleAge int

	for _, image := range task.Images {
		for _, face := range image.Faces {
			totalFaces++
			if face.Gender == "male" {
				maleFaces++
				totalMaleAge += face.Age
			} else if face.Gender == "female" {
				femaleFaces++
				totalFemaleAge += face.Age
			}
		}
	}

	var avgMaleAge, avgFemaleAge int

	if maleFaces > 0 {
		avgMaleAge = totalMaleAge / maleFaces
	}
	if femaleFaces > 0 {
		avgFemaleAge = totalFemaleAge / femaleFaces
	}

	task.FacesTotal = totalFaces
	task.FacesMale = maleFaces
	task.FacesFemale = femaleFaces
	task.AgeMaleAvg = avgMaleAge
	task.AgeFemaleAvg = avgFemaleAge
	task.Status = "completed"

	err := s.repository.UpdateTaskStatistics(task)
	if err != nil {
		_ = s.repository.UpdateTaskStatus(task.Id, "error")
		return
	}
}
