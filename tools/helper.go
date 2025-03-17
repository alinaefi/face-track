package tools

import (
	"image"
	"image/jpeg"
	"log"
	"os"
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
