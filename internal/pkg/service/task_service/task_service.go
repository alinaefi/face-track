package task_service

import (
	"database/sql"
	"errors"
	"face-track/internal/pkg/model"
	"face-track/internal/pkg/repo"
	"fmt"
	"image"
	"log"
	"os"
	"sync"

	"golang.org/x/sync/errgroup"
)

const (
	foldersAmount = 30000
)

type TaskService struct {
	repo *repo.Repo
}

func New(repo *repo.Repo) *TaskService {
	return &TaskService{
		repo: repo,
	}
}

type Response struct {
	Status int
	Data   interface{}
}

// GetTaskById returns task data, images, and faces associated with it by task ID.
func (s *TaskService) GetTaskById(taskId int) (task *model.Task, err error) {
	return s.getFullTaskData(taskId)
}

// GetTaskById returns task data as an object.
func (s *TaskService) getFullTaskData(taskId int) (task *model.Task, err error) {

	task, err = s.repo.Task.GetTaskById(taskId)
	if err != nil {
		return task, err
	}

	task.Images, err = s.repo.Task.GetTaskImages(taskId)
	if err != nil {
		return task, err
	}

	if len(task.Images) > 0 {
		imageIds := make([]int, len(task.Images))
		for _, img := range task.Images {
			imageIds = append(imageIds, img.Id)
		}

		faces, err := s.repo.Task.GetFacesByImageIds(imageIds)
		if err != nil {
			return task, err
		}

		for _, image := range task.Images {
			image.Faces = faces[image.Id]
		}
	}

	return task, err
}

// CreateTask creates new task and returns its ID.
func (s *TaskService) CreateTask() (taskId int, err error) {
	return s.repo.Task.CreateTask()
}

// DeleteTask deletes all task data from db and disk; returns error.
func (s *TaskService) DeleteTask(taskId int) (err error) {
	var task *model.Task

	task, err = s.repo.Task.GetTaskById(taskId)
	if err != nil {
		return err
	}

	if task.Status == "in_progress" {
		return errors.New("unable to delete task: processing is in progress")
	}

	if err = s.repo.Task.DeleteTask(taskId); err != nil {
		return err
	}

	s.deleteTaskImagesFromDisk(task.Id)

	return err
}

// deleteTaskImagesFromDisk removes task image folder with content from the disk; returns err.
func (s *TaskService) deleteTaskImagesFromDisk(taskId int) (err error) {

	subFolderID := taskId % foldersAmount
	path := fmt.Sprintf("/face track/images/%d/%d", subFolderID, taskId)

	return os.RemoveAll(path)
}

// AddImageToTask validates and adds a new image to task: to disk and database.
func (s *TaskService) AddImageToTask(taskId int, imageName string, fileData *model.FileData) (err error) {

	if err = s.validateTaskImage(taskId, imageName, fileData); err != nil {
		return err
	}

	// decode file to image type
	var image image.Image
	image, err = s.repo.Task.DecodeFile(fileData)
	if err != nil {
		return err
	}

	// save image on disk
	imageRow, err := s.repo.Task.SaveImageDisk(taskId, image, imageName)
	if err != nil {
		return err
	}

	return s.repo.Task.CreateImage(imageRow)
}

// validateImage validates the image and related task data; returns error.
func (s *TaskService) validateTaskImage(taskId int, imageName string, fileData *model.FileData) (err error) {
	var task *model.Task

	// validate file extension
	if fileData.FileHeader.Header.Get("Content-Type") != "image/jpeg" {
		return errors.New("unsupported file extension")
	}

	task, err = s.getFullTaskData(taskId)
	if err != nil {
		return err
	}

	// check task status
	if task.Status != "new" {
		return errors.New("failed to add image to task: task processing is in progress")
	}

	// TO DO allow same names for images
	// validate image file name
	for _, image := range task.Images {
		if image.ImageName == imageName {
			return errors.New("failed to add image to task: image with specified name already exists")
		}
	}
	return err
}

// UpdateTaskStatus обновляет статус задания на заданный
func (s *TaskService) UpdateTaskStatus(taskId int, status string) (err error) {
	errorMsg := fmt.Errorf("failed update task status")

	if err = s.repo.Task.UpdateTaskStatus(taskId, status); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("task with id %d not found", taskId)
		}
		return errorMsg
	}

	return nil
}

// ProcessTask запускает параллельную обработку изображений задания
func (s *TaskService) ProcessTask(taskId int) {

	// запрашиваем все данные о задании, его изображениях и лицах
	task, err := s.getFullTaskData(taskId)
	if err != nil {
		log.Println(err)
		_ = s.repo.Task.UpdateTaskStatus(taskId, "error")
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
		token, err := s.repo.Task.GetFaceCloudToken()
		if err != nil {
			log.Println(err)
			_ = s.repo.Task.UpdateTaskStatus(taskId, "error")
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
				imageData, err := s.repo.Task.GetFaceDetectionData(currImage, token)
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
	s.repo.Task.SaveProcessedData(facesToSave, imagesToSetDone)

	if err != nil {
		log.Println(err)
		_ = s.repo.Task.UpdateTaskStatus(taskId, "error")
		return
	}

	// запрашиваем обновленные данные о задании
	task, _ = s.getFullTaskData(taskId)

	// подсчитываем статистику задания и вызываем сохранение финальных данных
	s.concludeTask(task)
}

// concludeTask подсчитывает статистические данные задания и сохраняет в бд
func (s *TaskService) concludeTask(task *model.Task) {

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

	err := s.repo.Task.UpdateTaskStatistics(task)
	if err != nil {
		_ = s.repo.Task.UpdateTaskStatus(task.Id, "error")
		return
	}
}
