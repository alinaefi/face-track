package tools

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	foldersAmount            = 30000
	faceCloudApiUrlEnvName   = "FACE_CLOUD__API_URL"
	faceCloudUserEnvName     = "FACE_CLOUD__API_USER"
	faceCloudPasswordEnvName = "FACE_CLOUD__API_PASS"
)

// CheckEnvs проверяет переменные окружения на валидность
func CheckEnvs(envNames ...string) {
	for _, env := range envNames {
		envStr, exist := os.LookupEnv(env)

		if !exist {
			log.Fatalf("переменная окружения не найдена: %s", env)
			os.Exit(1)
		}

		if envStr == "" {
			log.Fatalf("переменная окружения не инициализирована: %s", env)
			os.Exit(1)
		}
	}
}

// CreateFolderIfNotExist создаёт папку, путь которой задан в переменной path
func CreateFolderIfNotExist(path string) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			panic(err)
		}
	} else if err != nil {
		panic(err)
	}
}

// SaveImg сохраняет изображение на диске
func SaveImg(img image.Image, imgPath string) (err error) {
	out, err := os.Create(imgPath)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	if err = jpeg.Encode(out, img, nil); err != nil {
		return err
	}

	return nil
}

// FaceCloudDetectRequest отправляет запрос к API Face Cloud на обнаружение лиц на изображениях
func FaceCloudDetectRequest(file *os.File, token string) (b []byte, err error) {

	url := fmt.Sprintf("%s/v1/detect?demographics=true", os.Getenv(faceCloudApiUrlEnvName))

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

// FaceCloudLoginRequest отправляет запрос к API Face Cloud на получение jwt токена
func FaceCloudLoginRequest(body []byte) (b []byte, err error) {

	url := fmt.Sprintf("%s/v1/login", os.Getenv(faceCloudApiUrlEnvName))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		log.Println(err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return reqUrl(req)
}

// reqUrl делает HTTP запрос и возвращает ошибку и ответ на завпрос
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
		return nil, errors.New("Сервер вернул ошибку на запрос со статусом: " + resp.Status)
	}

	return data, nil
}
