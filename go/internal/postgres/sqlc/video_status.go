package repo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

const (
	videoStatusPending    = "pending"
	videoStatusProcessing = "processing"
	videoStatusFinished   = "finished"
	videoStatusFailed     = "failed"
)

type videoStatusIDs struct {
	Pending    pgtype.UUID
	Processing pgtype.UUID
	Finished   pgtype.UUID
	Failed     pgtype.UUID
}

var VideoStatuses videoStatusIDs

func LoadVideoStatuses(ctx context.Context, q Querier) error {
	pending, err := q.FindStatusIdByName(ctx, videoStatusPending)
	if err != nil {
		return fmt.Errorf("load pending status: %w", err)
	}

	processing, err := q.FindStatusIdByName(ctx, videoStatusProcessing)
	if err != nil {
		return fmt.Errorf("load processing status: %w", err)
	}

	finished, err := q.FindStatusIdByName(ctx, videoStatusFinished)
	if err != nil {
		return fmt.Errorf("load finished status: %w", err)
	}

	failed, err := q.FindStatusIdByName(ctx, videoStatusFailed)
	if err != nil {
		return fmt.Errorf("load failed status: %w", err)
	}

	VideoStatuses = videoStatusIDs{
		Pending:    pending,
		Processing: processing,
		Finished:   finished,
		Failed:     failed,
	}

	return nil
}
