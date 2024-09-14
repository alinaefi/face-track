package internal

import (
	"context"
	"face-track/internal/handler"
	"face-track/internal/service"
	"face-track/utils"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	serverAddrEnvName = "FACE_TRACK__SERVER_ADDRESS"
)

func RunServer(srvs *service.Service) {

	utils.CheckEnvs(serverAddrEnvName)
	serverAddr := os.Getenv(serverAddrEnvName)

	handler := handler.NewHandler(srvs)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	router := gin.Default()

	setupRoutes(router, handler)

	server := &http.Server{
		Addr:    serverAddr,
		Handler: router,
	}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("gin error: %s\n", err)
		}
	}()

	<-ctx.Done()

	stop()
	log.Println("shutting down gracefully, press Ctrl+C again to force")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Println("Server exiting")
}
