// Package task_service provides the functionality to interact with task-related data in the repository.
package task_service

import (
	"errors"
	"face-track/internal/pkg/model/task_model"
	"face-track/internal/pkg/repo"
	"fmt"
	"image"
	"log"
	"os"
	"sync"

	"golang.org/x/sync/errgroup"
)

const (
	// foldersAmount defines the maximum number of nested folders for organizing task images.
	foldersAmount = 30000
)

// TaskService is a struct that holds methods for managing tasks and processing associated images.
type TaskService struct {
	repo *repo.Repo
}

// New creates a new instance of TaskService, initializing it with the provided repo.
func New(repo *repo.Repo) *TaskService {
	return &TaskService{
		repo: repo,
	}
}

// GetTaskById returns task data, images, and faces associated with it by task ID.
func (s *TaskService) GetTaskById(taskId int) (task *task_model.Task, err error) {
	return s.getFullTaskData(taskId)
}

// GetTaskById returns task data as an object.
func (s *TaskService) getFullTaskData(taskId int) (task *task_model.Task, err error) {

	task, err = s.repo.GetTaskById(taskId)
	if err != nil {
		return task, err
	}

	task.Images, err = s.repo.GetTaskImages(taskId)
	if err != nil {
		return task, err
	}

	if len(task.Images) > 0 {
		imageIds := make([]int, len(task.Images))
		for i, img := range task.Images {
			imageIds[i] = img.Id
		}

		faces, err := s.repo.GetFacesByImageIds(imageIds)
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
	return s.repo.CreateTask()
}

// DeleteTask deletes all task data from db and disk; returns error.
func (s *TaskService) DeleteTask(taskId int) (err error) {
	var task *task_model.Task

	task, err = s.repo.GetTaskById(taskId)
	if err != nil {
		return err
	}

	if task.Status == "in_progress" {
		return errors.New("unable to delete task: processing is in progress")
	}

	if err = s.repo.DeleteTask(taskId); err != nil {
		return err
	}

	if err = s.deleteTaskImagesFromDisk(task.Id); err != nil {
		log.Printf("error deleting images from disk: %v\n", err)
	}

	return nil
}

// deleteTaskImagesFromDisk removes task image folder with content from the disk; returns err.
func (s *TaskService) deleteTaskImagesFromDisk(taskId int) error {

	subFolderID := taskId % foldersAmount
	path := fmt.Sprintf("/face track/images/%d/%d", subFolderID, taskId)

	return os.RemoveAll(path)
}

// AddImageToTask validates and adds a new image to task: to disk and database.
func (s *TaskService) AddImageToTask(taskId int, fileData *task_model.FileData) (err error) {

	if err = s.validateTaskImage(taskId, fileData); err != nil {
		return err
	}

	// decode file to image type
	var image image.Image
	image, err = s.repo.DecodeFile(fileData)
	if err != nil {
		return err
	}

	// save image on disk
	fileName := fileData.FileHeader.Filename
	imageRow, err := s.repo.SaveImageDisk(taskId, image, fileName)
	if err != nil {
		return err
	}

	return s.repo.CreateImage(imageRow)
}

// validateImage validates the image and related task data; returns error.
func (s *TaskService) validateTaskImage(taskId int, fileData *task_model.FileData) error {

	// validate file extension
	if fileData.FileHeader.Header.Get("Content-Type") != "image/jpeg" {
		return errors.New("unsupported file extension")
	}

	// Check task status
	taskStatusNew := s.repo.ConfirmTaskStatus(taskId, "new")

	if !taskStatusNew {
		return errors.New("failed to add image to task: task status does not allow adding images")
	}

	return nil
}

// UpdateTaskStatus updates the task status to the specified value.
func (s *TaskService) UpdateTaskStatus(taskId int, status string) error {
	return s.repo.UpdateTaskStatus(taskId, status)
}

// ProcessTask processes tasks' images concurrently.
func (s *TaskService) ProcessTask(taskId int) {
	var err error
	var task *task_model.Task

	task, err = s.getFullTaskData(taskId)
	if err != nil {
		log.Println(err)
		_ = s.repo.UpdateTaskStatus(taskId, "error")
		return
	}
	if task.Status == "completed" {
		return
	}

	g := new(errgroup.Group)
	g.SetLimit(10)

	var Mu sync.RWMutex
	var facesToSave []*task_model.Face
	var imagesToSetDone []*task_model.Image

	if len(task.Images) > 0 {

		// get token for external API authentication
		token, err := s.repo.GetFaceCloudToken()
		if err != nil {
			log.Println(err)
			_ = s.repo.UpdateTaskStatus(taskId, "error")
			return
		}

		for _, img := range task.Images {
			// skip processed images
			if img.DoneFlag {
				continue
			}

			currImage := img
			g.Go(func() error {

				// send request to face cloud
				imageData, err := s.repo.GetFaceDetectionData(currImage, token)
				if err != nil {
					log.Println(err)
					return err
				}

				// process recognised faces data
				for _, faceData := range imageData.Data {
					newFace := &task_model.Face{
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

	// save processed images to db
	s.repo.SaveProcessedData(facesToSave, imagesToSetDone)

	if err != nil {
		log.Println(err)
		_ = s.repo.UpdateTaskStatus(taskId, "error")
		return
	}

	// request updated task data
	task, err = s.getFullTaskData(taskId)
	if err != nil {
		log.Println(err)
		return
	}

	// analize statistics data and save it to db
	s.concludeTask(task)
}

// concludeTask calculates task statistics and saves them to the database.
func (s *TaskService) concludeTask(task *task_model.Task) {

	var totalFaces, maleFaces, femaleFaces, totalMaleAge, totalFemaleAge int

	for _, image := range task.Images {
		for _, face := range image.Faces {
			totalFaces++

			switch face.Gender {
			case "male":
				maleFaces++
				totalMaleAge += face.Age
			case "female":
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

	err := s.repo.UpdateTaskStatistics(task)
	if err != nil {
		_ = s.repo.UpdateTaskStatus(task.Id, "error")
		return
	}
}
