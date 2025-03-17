package tools

import (
	"image"
	"image/jpeg"
	"log"
	"os"
)

const (
	foldersAmount = 30000
)

// CheckEnvs checks the environment variables.
func CheckEnvs(envNames ...string) {
	for _, env := range envNames {
		envStr, exist := os.LookupEnv(env)

		if !exist {
			log.Fatalf("env variable not found: %s", env)
			os.Exit(1)
		}

		if envStr == "" {
			log.Fatalf("env variable not initialized: %s", env)
			os.Exit(1)
		}
	}
}

// CreateFolderIfNotExist creates a folder at the path specified in the path variable.
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

// SaveImg saves the image to disk.
func SaveImg(img image.Image, imgPath string) (err error) {
	out, err := os.Create(imgPath)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	return jpeg.Encode(out, img, nil)
}
