package main

import (
	"context"
	"fmt"
	"net/http"

	repo "github.com/IkBenJur/streaming-service/internal/postgres/sqlc"
	"github.com/IkBenJur/streaming-service/internal/videoProcessing"
	"github.com/gin-gonic/gin"
)

type Application struct {
	Port       string
	Queries    repo.Querier
	Transcoder videoProcessing.Transcoder
}

func (app *Application) Mount() http.Handler {
	router := gin.Default()

	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
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

	router.MaxMultipartMemory = 8 << 20

	videoProcessingHandler := videoProcessing.NewHandler(
		videoProcessing.NewLocalStorage("./files", app.Queries),
		app.Transcoder,
	)
	router.POST("/upload-video", videoProcessingHandler.UploadVideo)
	router.GET("/video-stream/:id/:file", videoProcessingHandler.GetVideoStream)

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
