package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"github.com/IkBenJur/streaming-service/internal/env"
	"github.com/IkBenJur/streaming-service/internal/postgres"
	repo "github.com/IkBenJur/streaming-service/internal/postgres/sqlc"
	"github.com/IkBenJur/streaming-service/internal/videoProcessing"
	videotranscoder "github.com/IkBenJur/streaming-service/internal/videoTranscoder"
	"github.com/jackc/pgx/v5/pgxpool"
)

func run(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	dsn := env.GetEnv("GOOSE_DBSTRING", "host=localhost user=postgres password=postgres dbname=video-stream sslmode=disable")

	if err := postgres.RunMigrations(ctx, dsn); err != nil {
		slog.Error("Failed to run migrations", "error", err)
		return err
	}

	conn, err := pgxpool.New(ctx, dsn)
	if err != nil {
		slog.Error("Failed DB connections", "error", err)
		return err
	}
	defer conn.Close()

	slog.Info("Connected to DB")

	queries := repo.New(conn)

	if err := repo.LoadVideoStatuses(ctx, queries); err != nil {
		slog.Error("Failed to load video statuses", "error", err)
		return err
	}

	transcoder := videotranscoder.NewTranscoder(
		videoProcessing.NewLocalStorage("./files", queries),
		env.GetEnvInt("TRANCODE_JOB_NUMBER_OF_WORKERS", 2),
	)

	api := Application{
		Port:       env.GetEnv("PORT", "8080"),
		Queries:    queries,
		Transcoder: transcoder,
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
		os.Exit(1)
	}
}
