package videotranscoder

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/exec"
	"strconv"

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

		video, err := t.service.FindVideoById(c, id)
		if err != nil {
			logger.Error(fmt.Sprintf("transcode failed %s", err.Error()))
			return
		}

		filename := fmt.Sprintf("%x.%s", video.ID.Bytes, video.FileExtension)
		durationInMs, err := determineTranscodeDurationInMs(c, t.service.RawFilePath(filename))
		if err != nil {
			logger.Error(fmt.Sprintf("transcode failed to determine duration %s", err.Error()))
			t.service.UpdateVideoStatus(c, repo.UpdateVideoStatusParams{
				ID:     id,
				Status: repo.VideoStatuses.Failed,
			})
			return
		}

		logger.Info(fmt.Sprintf("transcode duration: %d ms", durationInMs))

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

func determineTranscodeDurationInMs(c context.Context, inputPath string) (int, error) {
	cmd := exec.CommandContext(c,
		"ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_entries", "stream=codec_type,duration",
		inputPath,
	)

	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	var result struct {
		Streams []struct {
			Codec_type string `json:"codec_type"`
			Duration   string `json:"duration"`
		} `json:"streams"`
	}

	if err = json.Unmarshal(out, &result); err != nil {
		return 0, err
	}

	for _, s := range result.Streams {
		if s.Codec_type != "video" {
			continue
		}

		seconds, err := strconv.ParseFloat(s.Duration, 64)
		if err != nil {
			return 0, err
		}

		return int(seconds * 1000), nil
	}

	return 0, nil
}
