package account

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rnrnshn/oportunidades-api/pkg/db/queries"
)

type PostgresRepository struct{ queries *queries.Queries }

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{queries: queries.New(pool)}
}

func (r *PostgresRepository) GetUserByID(ctx context.Context, id pgtype.UUID) (queries.User, error) {
	user, err := r.queries.GetUserByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return queries.User{}, ErrNotFound
	}
	return user, err
}

func (r *PostgresRepository) UpdateUserProfile(ctx context.Context, params queries.UpdateUserProfileParams) (queries.User, error) {
	user, err := r.queries.UpdateUserProfile(ctx, params)
	if errors.Is(err, pgx.ErrNoRows) {
		return queries.User{}, ErrNotFound
	}
	return user, err
}
