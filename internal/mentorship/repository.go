package mentorship

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rnrnshn/oportunidades-api/pkg/db/queries"
)

var ErrNotFound = errors.New("mentorship: not found")

type Repository interface {
	ListMentors(ctx context.Context, params queries.ListMentorsParams) ([]queries.ListMentorsRow, error)
	CountMentors(ctx context.Context) (int64, error)
	GetMentorByID(ctx context.Context, userID pgtype.UUID) (queries.GetMentorByIDRow, error)
	CreateMentorshipSession(ctx context.Context, params queries.CreateMentorshipSessionParams) (queries.MentorshipSession, error)
}
