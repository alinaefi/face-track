package main

import (
	"context"
	"face-track/internal/pkg/handler"
	"face-track/internal/pkg/service"
	"log"
	"net/http"
	"os/signal"
	"syscall"
)

func main() {
	s := service.NewServiceWithRepo()

	runServer(s)
}

func runServer(s *service.Service) {

	signalCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	server := handler.NewServer(s)
	defer server.Close()

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("gin error: %s\n", err)
		}
	}()

	<-signalCtx.Done()

	log.Println("Server exiting")
}
