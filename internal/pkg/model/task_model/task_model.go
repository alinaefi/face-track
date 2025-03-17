package task_model

import "mime/multipart"

// request types
type TaskIdReq struct {
	TaskId int `json:"id"`
}

// models for working with API and database
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

type Statistics struct {
	FacesTotal   int `db:"faces_total" json:"facesTotal"`
	FacesMale    int `db:"faces_male" json:"facesMale"`
	FacesFemale  int `db:"faces_female" json:"facesFemale"`
	AgeFemaleAvg int `db:"age_female_avg" json:"ageFemaleAvg"`
	AgeMaleAvg   int `db:"age_male_avg" json:"ageMaleAvg"`
}

type Image struct {
	Id        int     `db:"id" json:"-"`
	TaskId    int     `db:"task_id" json:"-"`
	ImageName string  `db:"image_name" json:"name"`
	DoneFlag  bool    `db:"done" json:"-"`
	Faces     []*Face `json:"faces"`
}

type Face struct {
	Id      int    `db:"id" json:"-"`
	ImageId int    `db:"image_id" json:"-"`
	Gender  string `db:"gender" json:"gender"`
	Age     int    `db:"age" json:"age"`
	Height  int    `db:"bbox_height" json:"-"`
	Width   int    `db:"bbox_width" json:"-"`
	X       int    `db:"bbox_x" json:"-"`
	Y       int    `db:"bbox_y" json:"-"`
	Bbox    Bbox   `json:"bbox"`
}

type FileData struct {
	File       multipart.File
	FileHeader *multipart.FileHeader
}

// models for working with face cloud
type FaceCloudLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type FaceCloudLoginResponse struct {
	Data       Data `json:"data"`
	StatusCode int  `json:"status_code"`
}

type Data struct {
	AccessToken string `json:"access_token"`
}

type FaceCloudDetectResponse struct {
	Data       []FaceData `json:"data"`
	Rotation   int        `json:"rotation"`
	StatusCode int        `json:"status_code"`
}

type FaceData struct {
	Attributes   Attributes   `json:"attributes"`
	Bbox         Bbox         `json:"bbox"`
	Demographics Demographics `json:"demographics"`
	Landmarks    []Landmark   `json:"landmarks"`
	Liveness     int          `json:"liveness"`
	Masks        Masks        `json:"masks"`
	Quality      Quality      `json:"quality"`
	Score        float64      `json:"score"`
}

type Attributes struct {
	FacialHair string `json:"facial_hair"`
	Glasses    string `json:"glasses"`
	HairColor  string `json:"hair_color"`
	HairType   string `json:"hair_type"`
	Headwear   string `json:"headwear"`
}

type Bbox struct {
	Height int `json:"height"`
	Width  int `json:"width"`
	X      int `json:"x"`
	Y      int `json:"y"`
}

type Demographics struct {
	Age       Age    `json:"age"`
	Ethnicity string `json:"ethnicity"`
	Gender    string `json:"gender"`
}

type Landmark struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Masks struct {
	FullFaceMask  int `json:"full_face_mask"`
	LowerFaceMask int `json:"lower_face_mask"`
	NoMask        int `json:"no_mask"`
	OtherMask     int `json:"other_mask"`
}

type Quality struct {
	Blurriness    int `json:"blurriness"`
	Overexposure  int `json:"overexposure"`
	Underexposure int `json:"underexposure"`
}

type Age struct {
	Mean     float64 `json:"mean"`
	Variance float64 `json:"variance"`
}
