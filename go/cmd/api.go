package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	repo "github.com/IkBenJur/streaming-service/internal/postgres/sqlc"
	"github.com/IkBenJur/streaming-service/internal/storage"
	"github.com/IkBenJur/streaming-service/internal/videos"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

type Application struct {
	Port          string
	Queries       repo.Querier
	Transcoder    storage.Transcoder
	StorageClient storage.StorageClient
	LocalStorage  bool
}

func (app *Application) restartLooseVideoProcessingJobs(ctx context.Context) {
	slog.Info("Restarting loose jobs")
	videos, err := app.Queries.ListVideosByStatus(ctx, repo.VideoStatuses.Processing)
	if err != nil {
		slog.Error("Error restart, list videos", "error", err)
		return
	}

	slog.Info("Submitting new jobs", "count", len(videos))
	for _, video := range videos {
		// Update status to processing
		app.Queries.UpdateVideoStatus(ctx, repo.UpdateVideoStatusParams{
			ID:     video.ID,
			Status: repo.VideoStatuses.Processing,
		})
		app.Transcoder.Submit(video.ID)
		// Video submited reset the progress
		app.Queries.UpdateVideoProgress(ctx, repo.UpdateVideoProgressParams{
			ID:       video.ID,
			Progress: pgtype.Int4{Int32: 0, Valid: true},
		})
	}
	slog.Info("Done restart")
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
	router.GET("/videos/:id/stream/:file/signed-url", storageHandler.GetSegmentSignedUrl)

	videoHandler := videos.NewHandler(app.Queries)
	router.GET("/videos", videoHandler.ListVideos)
	videoGroup := router.Group("/videos/:id", videos.RequireVideo(app.Queries))
	videoGroup.GET("", videoHandler.FindById)
	videoGroup.GET("/is-status-finished", videoHandler.VideoStatusIsFinished)

	if app.LocalStorage {
		localStore := app.StorageClient.(*storage.LocalStorage)
		router.MaxMultipartMemory = 8 << 20
		localUploadHandler := storage.NewLocalUploadHandler(localStore)
		router.PUT("/videos/upload-raw/:filename", localUploadHandler.UploadRawFile)
		router.GET("/videos/hls/:id/:file", localUploadHandler.ServeHlsFile)
	}

	return router
}

func (app *Application) Run(ctx context.Context, router http.Handler) error {
	app.restartLooseVideoProcessingJobs(ctx)

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
