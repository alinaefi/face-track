package cmd

import (
	"face-track/internal"
	"face-track/internal/service"
	"log"
)

func Execute() {

	action := "run-server"

	processAction(action)
}

func processAction(arg string) {
	log.Println("Processing action:", arg)

	service := service.NewService()

	switch arg {
	case "run-server":
		internal.RunServer(service)
	default:
		panic("invalid action")
	}
}
