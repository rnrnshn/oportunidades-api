package mentorship

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rnrnshn/oportunidades-api/pkg/db/queries"
)

type PostgresRepository struct {
	queries *queries.Queries
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{queries: queries.New(pool)}
}

func (r *PostgresRepository) ListMentors(ctx context.Context, params queries.ListMentorsParams) ([]queries.ListMentorsRow, error) {
	return r.queries.ListMentors(ctx, params)
}

func (r *PostgresRepository) CountMentors(ctx context.Context) (int64, error) {
	return r.queries.CountMentors(ctx)
}

func (r *PostgresRepository) GetMentorByID(ctx context.Context, userID pgtype.UUID) (queries.GetMentorByIDRow, error) {
	row, err := r.queries.GetMentorByID(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return queries.GetMentorByIDRow{}, ErrNotFound
	}

	return row, err
}

func (r *PostgresRepository) CreateMentorshipSession(ctx context.Context, params queries.CreateMentorshipSessionParams) (queries.MentorshipSession, error) {
	return r.queries.CreateMentorshipSession(ctx, params)
}
