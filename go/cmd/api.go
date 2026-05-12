package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/IkBenJur/streaming-service/internal/videoProcessing"
	"github.com/gin-gonic/gin"
)

type Application struct {
	Port string
	// db pgx
}

func (app *Application) Mount() http.Handler {
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
