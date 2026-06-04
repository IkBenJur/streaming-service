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
	"path/filepath"
	"strconv"
	"strings"

	repo "github.com/IkBenJur/streaming-service/internal/postgres/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type StorageClient interface {
	GetRawKey(string) string
	GetHlsKey(string) string
	GetRawFilePath(string) string
	GetHlsFilePath(string) string
	DeleteRawLocalFile(ctx context.Context, localPath string) error
	UploadFile(ctx context.Context, localPath string, key string) error
	DownloadFile(ctx context.Context, key string, destPath string) error
	DeleteHlsLocalFolder(filePath string) error
}

type VideoTranscoder struct {
	ctx context.Context
	repo.Querier
	storageClient StorageClient
	semaphore     chan struct{}
}

func NewTranscoder(ctx context.Context, querier repo.Querier, storageClient StorageClient, maxNrOfWorkers int) *VideoTranscoder {
	return &VideoTranscoder{
		ctx:           ctx,
		Querier:       querier,
		storageClient: storageClient,
		semaphore:     make(chan struct{}, maxNrOfWorkers),
	}
}

func (t *VideoTranscoder) Submit(id pgtype.UUID) {
	go func() {
		select {
		case t.semaphore <- struct{}{}:
		case <-t.ctx.Done():
			return
		}
		defer func() { <-t.semaphore }()

		logger := slog.With("video_id", id)
		logger.Info("transcode job start")

		video, err := t.Querier.FindVideoById(t.ctx, id)
		if err != nil {
			// Most likely failed because does not exists. Does not hurt to try and set to failed
			t.Querier.UpdateVideoStatus(context.Background(), repo.UpdateVideoStatusParams{
				ID:     id,
				Status: repo.VideoStatuses.Failed,
			})
			logger.Error("transcode failed", "error", err)
			return
		}

		logger.Info("Getting file start")
		key := t.storageClient.GetRawKey(fmt.Sprintf("%x.%s", id.Bytes, video.FileExtension))
		filename := fmt.Sprintf("%x.%s", id.Bytes, video.FileExtension)
		rawFilepath := t.storageClient.GetRawFilePath(filename)
		err = t.storageClient.DownloadFile(t.ctx, key, rawFilepath)
		if err != nil {
			logger.Error("transcode failed", "error", err)
			t.Querier.UpdateVideoStatus(context.Background(), repo.UpdateVideoStatusParams{
				ID:     id,
				Status: repo.VideoStatuses.Failed,
			})
			return
		}
		logger.Info("Getting file end")

		logger.Info("transcode start")
		if err = t.transcodeVideo(t.ctx, video); err != nil {
			logger.Error("failed to transcode", "error", err)
			err = t.Querier.UpdateVideoStatus(context.Background(), repo.UpdateVideoStatusParams{
				ID:     id,
				Status: repo.VideoStatuses.Failed,
			})
			if err != nil {
				logger.Error("failed to update video status", "error", err)
			}
			return
		}
		logger.Info("transcode finished")

		logger.Info("delete local file raw start")
		if err = t.storageClient.DeleteRawLocalFile(t.ctx, rawFilepath); err != nil {
			logger.Error("failed to remove raw file", "error", err)
		}
		logger.Info("delete local file raw end")

		logger.Info("HLS files upload start")
		if err = t.uploadHLSFiles(t.ctx, fmt.Sprintf("%x", video.ID.Bytes)); err != nil {
			logger.Error("failed to upload HLS", "error", err)
			err = t.Querier.UpdateVideoStatus(context.Background(), repo.UpdateVideoStatusParams{
				ID:     id,
				Status: repo.VideoStatuses.Failed,
			})
			if err != nil {
				logger.Error("failed to update video status %s", "error", err)
			}
			return
		}
		logger.Info("HLS files upload end")

		if err = t.Querier.UpdateVideoStatus(context.Background(), repo.UpdateVideoStatusParams{
			ID:     id,
			Status: repo.VideoStatuses.Finished,
		}); err != nil {
			logger.Error("failed to update video status %s", "error", err)
		}

		logger.Info("transcode job end")
	}()
}

func (t *VideoTranscoder) uploadHLSFiles(ctx context.Context, videoID string) error {
	outputPath := t.storageClient.GetHlsFilePath(videoID)

	entries, err := os.ReadDir(outputPath)
	if err != nil {
		return fmt.Errorf("read HLS output dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		localPath := filepath.Join(outputPath, entry.Name())
		key := fmt.Sprintf("hls/%s/%s", videoID, entry.Name())
		if err = t.storageClient.UploadFile(ctx, localPath, key); err != nil {
			return fmt.Errorf("upload %s: %w", entry.Name(), err)
		}
	}

	return t.storageClient.DeleteHlsLocalFolder(outputPath)
}

func (t *VideoTranscoder) transcodeVideo(c context.Context, video repo.Video) error {
	filepath := t.storageClient.GetRawFilePath(fmt.Sprintf("%x.%s", video.ID.Bytes, video.FileExtension))
	durationInUs, err := determineTranscodeDurationInUs(c, filepath)
	if err != nil {
		return fmt.Errorf("transcode failed to determine duration %s", err.Error())
	}

	outputPath := t.storageClient.GetHlsFilePath(fmt.Sprintf("%x", video.ID.Bytes))

	cmd := exec.CommandContext(c,
		"ffmpeg",
		"-i", filepath,
		"-progress", "pipe:1",
		"-loglevel", "warning",
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
			// Update may fail but that shouldn't stop trancode job
			_ = t.Querier.UpdateVideoProgress(c,
				repo.UpdateVideoProgressParams{
					ID:       video.ID,
					Progress: pgtype.Int4{Int32: int32(progress), Valid: true},
				})
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
