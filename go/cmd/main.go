package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"github.com/jackc/pgx/v5"
)

func run(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// TODO Use env varaibles
	conn, err := pgx.Connect(ctx, "host=localhost user=postgres password=postgres dbname=video-stream sslmode=disable")
	if err != nil {
		slog.Error("Failed DB connections", "error", err)
		return err
	}
	defer conn.Close(ctx)

	slog.Info("Connected to DB")

	api := Application{
		Port: "8080",
	}

	slog.Info("Starting server")
	if err := api.Run(ctx, api.Mount()); err != nil && err != http.ErrServerClosed {
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
