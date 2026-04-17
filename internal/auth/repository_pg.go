package auth

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

func (r *PostgresRepository) CreateUser(ctx context.Context, params queries.CreateUserParams) (queries.User, error) {
	user, err := r.queries.CreateUser(ctx, params)
	if err != nil {
		return queries.User{}, err
	}

	return user, nil
}

func (r *PostgresRepository) GetUserByEmail(ctx context.Context, email string) (queries.User, error) {
	user, err := r.queries.GetUserByEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return queries.User{}, ErrNotFound
	}

	return user, err
}

func (r *PostgresRepository) GetUserByID(ctx context.Context, id pgtype.UUID) (queries.User, error) {
	user, err := r.queries.GetUserByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return queries.User{}, ErrNotFound
	}

	return user, err
}

func (r *PostgresRepository) CreateRefreshToken(ctx context.Context, params queries.CreateRefreshTokenParams) (queries.RefreshToken, error) {
	token, err := r.queries.CreateRefreshToken(ctx, params)
	if err != nil {
		return queries.RefreshToken{}, err
	}

	return token, nil
}

func (r *PostgresRepository) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (queries.RefreshToken, error) {
	token, err := r.queries.GetRefreshTokenByHash(ctx, tokenHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return queries.RefreshToken{}, ErrNotFound
	}

	return token, err
}

func (r *PostgresRepository) RevokeRefreshToken(ctx context.Context, id pgtype.UUID) error {
	return r.queries.RevokeRefreshToken(ctx, id)
}
