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
	"github.com/IkBenJur/streaming-service/internal/storage"
	"github.com/IkBenJur/streaming-service/internal/videoProcessing"
	videotranscoder "github.com/IkBenJur/streaming-service/internal/videoTranscoder"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func run(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	godotenv.Load()

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
		env.GetEnvInt("TRANSCODE_JOB_NUMBER_OF_WORKERS", 2),
	)

	runLocalStorage := env.GetEnvBool("RUN_LOCAL_STORAGE", false)

	var s3Client *storage.S3Storage
	if runLocalStorage {
		slog.Warn("Running with local storage")
	} else {
		slog.Info("Running with S3")
		awsConfig, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return err
		}

		bucketName, err := env.GetEnvOrErr("AWS_S3_BUCKET_NAME")
		slog.Info(bucketName)
		if err != nil {
			return err
		}
		s3Client = storage.NewS3Storage(
			awsConfig,
			bucketName,
		)
	}

	api := Application{
		Port:       env.GetEnv("PORT", "8080"),
		Queries:    queries,
		Transcoder: transcoder,
		S3Client:   *s3Client,
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
