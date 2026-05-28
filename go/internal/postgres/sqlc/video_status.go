package repo

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type videoStatusIDs struct {
	Pending    pgtype.UUID
	Processing pgtype.UUID
	Finished   pgtype.UUID
	Failed     pgtype.UUID
}

var VideoStatuses = videoStatusIDs{
	Pending:    mustUUID("00000000-0000-0000-0000-000000000001"),
	Processing: mustUUID("00000000-0000-0000-0000-000000000002"),
	Finished:   mustUUID("00000000-0000-0000-0000-000000000003"),
	Failed:     mustUUID("00000000-0000-0000-0000-000000000004"),
}

func mustUUID(s string) pgtype.UUID {
	return pgtype.UUID{Bytes: uuid.MustParse(s), Valid: true}
}
