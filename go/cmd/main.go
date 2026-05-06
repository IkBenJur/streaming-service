package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"github.com/gin-gonic/gin"
)

func run(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	router := gin.Default()
	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "OK",
		})
	})

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
