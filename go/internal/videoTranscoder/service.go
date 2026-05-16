package videotranscoder

import (
	"context"
	"log/slog"

	repo "github.com/IkBenJur/streaming-service/internal/postgres/sqlc"
	"github.com/IkBenJur/streaming-service/internal/videoProcessing"
	"github.com/jackc/pgx/v5/pgtype"
)

type VideoTranscoder struct {
	service   videoProcessing.VideoProcessingService
	semaphore chan struct{}
}

func NewTranscoder(service videoProcessing.VideoProcessingService, maxNrOfWorkers int) *VideoTranscoder {
	return &VideoTranscoder{
		service:   service,
		semaphore: make(chan struct{}, maxNrOfWorkers),
	}
}

func (t *VideoTranscoder) Submit(id pgtype.UUID) {
	go func() {
		t.semaphore <- struct{}{}
		defer func() { <-t.semaphore }()

		c := context.Background()

		logger := slog.With("video_id", id)
		logger.Info("transcode start")

		// Update status to processing
		t.service.UpdateVideoStatus(c, repo.UpdateVideoStatusParams{
			ID:     id,
			Status: repo.VideoStatuses.Processing,
		})

		// FFMPEG

		// Handle failure
		//		delete files and update status

		t.service.UpdateVideoStatus(c, repo.UpdateVideoStatusParams{
			ID:     id,
			Status: repo.VideoStatuses.Finished,
		})

		logger.Info("transcode finished")
	}()
}
