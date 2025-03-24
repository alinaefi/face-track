package face_cloud_client

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	faceCloudApiUrlEnvName   = "FACE_CLOUD__API_URL"
	faceCloudUserEnvName     = "FACE_CLOUD__API_USER"
	faceCloudPasswordEnvName = "FACE_CLOUD__API_PASS"
)

// DetectFases sends a request to the Face Cloud API to detect faces in images.
func DetectFaces(file *os.File, token string) (b []byte, err error) {

	url := fmt.Sprintf("%s/detect?demographics=true", os.Getenv(faceCloudApiUrlEnvName))

	req, err := http.NewRequest("POST", url, file)
	if err != nil {
		return nil, err
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "image/jpeg")
	req.Header.Set("Authorization", "Bearer "+token)
	req.ContentLength = fileInfo.Size()

	return reqUrl(req)
}

// Login sends a request to the Face Cloud API to obtain a JWT token.
func Login(body []byte) (b []byte, err error) {

	url := fmt.Sprintf("%s/login", os.Getenv(faceCloudApiUrlEnvName))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		log.Println(err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return reqUrl(req)
}

// reqUrl makes an HTTP request and returns the response and an error.
func reqUrl(req *http.Request) (data []byte, err error) {

	client := &http.Client{Timeout: time.Second * time.Duration(10)}

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer resp.Body.Close()

	data, err = io.ReadAll(resp.Body)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	if resp.StatusCode >= 400 && resp.StatusCode <= 599 {
		return nil, errors.New("server returned error with status: " + resp.Status)
	}

	return data, nil
}
