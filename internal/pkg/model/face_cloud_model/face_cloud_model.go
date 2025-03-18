// Package face_cloud_model provides models for interacting with the face detection service.
package face_cloud_model

// FaceCloudLoginRequest represents a login request for face detection services.
type FaceCloudLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// FaceCloudLoginResponse contains authentication data from the face detection service.
type FaceCloudLoginResponse struct {
	Data       Data `json:"data"`
	StatusCode int  `json:"status_code"`
}

// Data contains authentication tokens.
type Data struct {
	AccessToken string `json:"access_token"`
}

// FaceCloudDetectResponse holds the response from a face detection request.
type FaceCloudDetectResponse struct {
	Data       []FaceData `json:"data"`
	Rotation   int        `json:"rotation"`
	StatusCode int        `json:"status_code"`
}

// FaceData contains detailed information about detected faces.
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

// Attributes represents various facial features.
type Attributes struct {
	FacialHair string `json:"facial_hair"`
	Glasses    string `json:"glasses"`
	HairColor  string `json:"hair_color"`
	HairType   string `json:"hair_type"`
	Headwear   string `json:"headwear"`
}

// Bbox represents the bounding box of a detected face.
type Bbox struct {
	Height int `json:"height"`
	Width  int `json:"width"`
	X      int `json:"x"`
	Y      int `json:"y"`
}

// Demographics contains personal attributes of a detected face.
type Demographics struct {
	Age       Age    `json:"age"`
	Ethnicity string `json:"ethnicity"`
	Gender    string `json:"gender"`
}

// Landmark represents a facial landmark coordinate.
type Landmark struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Masks contains information on face mask detection.
type Masks struct {
	FullFaceMask  int `json:"full_face_mask"`
	LowerFaceMask int `json:"lower_face_mask"`
	NoMask        int `json:"no_mask"`
	OtherMask     int `json:"other_mask"`
}

// Quality represents image quality metrics for face detection.
type Quality struct {
	Blurriness    int `json:"blurriness"`
	Overexposure  int `json:"overexposure"`
	Underexposure int `json:"underexposure"`
}

// Age provides age estimation data.
type Age struct {
	Mean     float64 `json:"mean"`
	Variance float64 `json:"variance"`
}
