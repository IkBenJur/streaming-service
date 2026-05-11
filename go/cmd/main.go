package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"github.com/IkBenJur/streaming-service/internal/videoProcessing"
	"github.com/gin-gonic/gin"
)

func run(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	router := gin.Default()

	// Ignore CORS for now
	// TODO SETUP CORS
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Next()
	})

	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "OK",
		})
	})

	router.MaxMultipartMemory = 8 << 20

	videoProcessingHandler := videoProcessing.NewHandler(videoProcessing.NewLocalStorage("./files"))
	router.POST("/upload-video", videoProcessingHandler.UploadVideo)
	router.GET("/stream-video/:fileName", videoProcessingHandler.StreamVideo)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		<-ctx.Done()
		srv.Shutdown(context.Background())
	}()

	slog.Info("Starting server")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("Server shutdown", "error", err)
		return err
	}
	return nil
}

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(0)
	}
}
