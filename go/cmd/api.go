package main

import (
	"context"
	"fmt"
	"net/http"

	repo "github.com/IkBenJur/streaming-service/internal/postgres/sqlc"
	"github.com/IkBenJur/streaming-service/internal/storage"
	"github.com/IkBenJur/streaming-service/internal/videoProcessing"
	"github.com/gin-gonic/gin"
)

type Application struct {
	Port          string
	Queries       repo.Querier
	Transcoder    videoProcessing.Transcoder
	StorageClient storage.StorageClient
	LocalStorage  bool
}

func (app *Application) Mount() http.Handler {
	router := gin.Default()

	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "OK",
		})
	})

	storageHandler := storage.NewHandler(app.Queries, app.StorageClient, app.Transcoder)
	router.POST("/videos/create-and-get-upload-url", storageHandler.CreateVideoAndGetUploadUrl)
	router.POST("/videos/:id/process", storageHandler.SubmitVideoProcessJob)

	if app.LocalStorage {
		localStore := app.StorageClient.(*storage.LocalStorage)
		router.MaxMultipartMemory = 8 << 20
		localUploadHandler := storage.NewLocalUploadHandler(localStore)
		router.PUT("/videos/upload-raw/:filename", localUploadHandler.UploadRawFile)

		videoProcessingHandler := videoProcessing.NewHandler(localStore, app.Queries, app.Transcoder)
		router.GET("/video-stream/:id/:file", videoProcessingHandler.GetVideoStream)
	}

	return router
}

func (app *Application) Run(ctx context.Context, router http.Handler) error {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", app.Port),
		Handler: router,
	}

	go func() {
		<-ctx.Done()
		srv.Shutdown(context.Background())
	}()

	return srv.ListenAndServe()
}
