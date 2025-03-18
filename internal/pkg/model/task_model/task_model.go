// Package task_model defines data structures for handling tasks, images, and face detection.
package task_model

import (
	"face-track/internal/pkg/model/face_cloud_model"
	"mime/multipart"
)

// Models for working with API and database

// Request types
// TaskIdRequest represents a request containing a task ID.
type TaskIdRequest struct {
	TaskId int `json:"id"`
}

// Task represents a task with its status, images, and statistics.
type Task struct {
	Id           int        `db:"id" json:"id"`
	Status       string     `db:"task_status" json:"taskStatus"`
	Images       []*Image   `json:"images"`
	FacesTotal   int        `db:"faces_total" json:"-"`
	FacesMale    int        `db:"faces_male" json:"-"`
	FacesFemale  int        `db:"faces_female" json:"-"`
	AgeFemaleAvg int        `db:"age_female_avg" json:"-"`
	AgeMaleAvg   int        `db:"age_male_avg" json:"-"`
	Statistics   Statistics `json:"statistics"`
}

// Statistics holds aggregated face detection data.
type Statistics struct {
	FacesTotal   int `db:"faces_total" json:"facesTotal"`
	FacesMale    int `db:"faces_male" json:"facesMale"`
	FacesFemale  int `db:"faces_female" json:"facesFemale"`
	AgeFemaleAvg int `db:"age_female_avg" json:"ageFemaleAvg"`
	AgeMaleAvg   int `db:"age_male_avg" json:"ageMaleAvg"`
}

// Image represents an image linked to a task.
type Image struct {
	Id        int     `db:"id" json:"-"`
	TaskId    int     `db:"task_id" json:"-"`
	ImageName string  `db:"image_name" json:"name"`
	DoneFlag  bool    `db:"done" json:"-"`
	Faces     []*Face `json:"faces"`
}

// Face represents detected facial attributes within an image.
type Face struct {
	Id      int                   `db:"id" json:"-"`
	ImageId int                   `db:"image_id" json:"-"`
	Gender  string                `db:"gender" json:"gender"`
	Age     int                   `db:"age" json:"age"`
	Height  int                   `db:"bbox_height" json:"-"`
	Width   int                   `db:"bbox_width" json:"-"`
	X       int                   `db:"bbox_x" json:"-"`
	Y       int                   `db:"bbox_y" json:"-"`
	Bbox    face_cloud_model.Bbox `json:"bbox"`
}

// FileData represents a file uploaded via multipart form.
type FileData struct {
	File       multipart.File
	FileHeader *multipart.FileHeader
}
