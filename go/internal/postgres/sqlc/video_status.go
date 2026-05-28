package repo

import (
	"github.com/IkBenJur/streaming-service/internal/utils"
	"github.com/jackc/pgx/v5/pgtype"
)

type videoStatusIDs struct {
	Pending    pgtype.UUID
	Processing pgtype.UUID
	Finished   pgtype.UUID
	Failed     pgtype.UUID
}

var VideoStatuses = videoStatusIDs{
	Pending:    utils.MustUUID("00000000-0000-0000-0000-000000000001"),
	Processing: utils.MustUUID("00000000-0000-0000-0000-000000000002"),
	Finished:   utils.MustUUID("00000000-0000-0000-0000-000000000003"),
	Failed:     utils.MustUUID("00000000-0000-0000-0000-000000000004"),
}
