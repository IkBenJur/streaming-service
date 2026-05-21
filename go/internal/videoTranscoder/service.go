package videotranscoder

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strconv"
	"strings"

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
		ctx := context.Background()
		t.semaphore <- struct{}{}
		defer func() { <-t.semaphore }()

		logger := slog.With("video_id", id)
		logger.Info("transcode job start")

		// Update status to processing
		t.service.UpdateVideoStatus(ctx, repo.UpdateVideoStatusParams{
			ID:     id,
			Status: repo.VideoStatuses.Processing,
		})

		video, err := t.service.FindVideoById(ctx, id)
		if err != nil {
			logger.Error(fmt.Sprintf("transcode failed %s", err.Error()))
			return
		}

		logger.Info("transcode start")
		if err = t.transcodeVideo(ctx, video); err != nil {
			logger.Error(err.Error())
			err = t.service.UpdateVideoStatus(ctx, repo.UpdateVideoStatusParams{
				ID:     id,
				Status: repo.VideoStatuses.Failed,
			})
			if err != nil {
				logger.Error(fmt.Sprintf("failed to update video status %s", err.Error()))
			}
			return
		}
		logger.Info("transcode finished")

		rawPath := t.service.RawFilePath(fmt.Sprintf("%x.%s", video.ID.Bytes, video.FileExtension))
		if err = os.Remove(rawPath); err != nil {
			logger.Error(fmt.Sprintf("failed to remove raw file %s", err.Error()))
		}

		if err = t.service.UpdateVideoStatus(ctx, repo.UpdateVideoStatusParams{
			ID:     id,
			Status: repo.VideoStatuses.Finished,
		}); err != nil {
			logger.Error(fmt.Sprintf("failed to update video status %s", err.Error()))
		}

		logger.Info("transcode job end")
	}()
}

func (t *VideoTranscoder) transcodeVideo(c context.Context, video repo.Video) error {
	filepath := t.service.RawFilePath(fmt.Sprintf("%x.%s", video.ID.Bytes, video.FileExtension))
	durationInUs, err := determineTranscodeDurationInUs(c, filepath)
	if err != nil {
		return fmt.Errorf("transcode failed to determine duration %s", err.Error())
	}

	outputPath, err := t.service.HLSOutputPath(fmt.Sprintf("%x", video.ID.Bytes))
	if err != nil {
		return fmt.Errorf("transcode failed to create output path %s", err.Error())
	}

	cmd := exec.CommandContext(c,
		"ffmpeg",
		"-i", filepath,
		"-progress", "pipe:1",
		"-loglevel", "error",
		"-c:v", "libx264",
		"-c:a", "aac",
		"-f", "hls",
		"-hls_time", "6",
		"-hls_playlist_type", "vod",
		"-hls_segment_type", "fmp4",
		"-hls_segment_filename", outputPath+"/segment%03d.m4s", outputPath+"/playlist.m3u8",
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("transcode failed to get stdout %s", err.Error())
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("transcode failed to start command  %s", err.Error())
	}

	scanner := bufio.NewScanner(stdout)
scanLoop:
	for scanner.Scan() {
		line := scanner.Text()
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		switch key {
		case "out_time_ms":
			ms, _ := strconv.ParseInt(value, 10, 64)
			progress := ms * 100 / int64(durationInUs)
			err = t.service.UpdateVideoProgress(c,
				repo.UpdateVideoProgressParams{
					ID:       video.ID,
					Progress: pgtype.Int4{Int32: int32(progress), Valid: true},
				})
			if err != nil {
				return fmt.Errorf("transcode failed at video progess update %s", err.Error())
			}
		case "progress":
			if value == "end" {
				break scanLoop
			}
		}
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("transcode failed ffmpeg %s", stderr.String())
	}

	return nil
}

func determineTranscodeDurationInUs(c context.Context, inputPath string) (int, error) {
	cmd := exec.CommandContext(c,
		"ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_entries", "stream=codec_type,duration:format=duration",
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
		Format struct {
			Duration string `json:"duration"`
		} `json:"format"`
	}

	if err = json.Unmarshal(out, &result); err != nil {
		return 0, err
	}

	for _, s := range result.Streams {
		if s.Codec_type != "video" {
			continue
		}

		if s.Duration != "" {
			seconds, err := strconv.ParseFloat(s.Duration, 64)
			if err != nil {
				return 0, err
			}
			return int(seconds * 1000 * 1000), nil
		}
		break
	}

	if result.Format.Duration != "" {
		seconds, err := strconv.ParseFloat(result.Format.Duration, 64)
		if err != nil {
			return 0, err
		}
		return int(seconds * 1000 * 1000), nil
	}

	return 0, fmt.Errorf("was not able to find video duration")
}
