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

func (r *PostgresRepository) CreateAuthActionToken(ctx context.Context, params queries.CreateAuthActionTokenParams) (queries.AuthActionToken, error) {
	token, err := r.queries.CreateAuthActionToken(ctx, params)
	if err != nil {
		return queries.AuthActionToken{}, err
	}
	return token, nil
}

func (r *PostgresRepository) GetAuthActionTokenByHash(ctx context.Context, params queries.GetAuthActionTokenByHashParams) (queries.AuthActionToken, error) {
	token, err := r.queries.GetAuthActionTokenByHash(ctx, params)
	if errors.Is(err, pgx.ErrNoRows) {
		return queries.AuthActionToken{}, ErrNotFound
	}
	return token, err
}

func (r *PostgresRepository) ConsumeAuthActionToken(ctx context.Context, id pgtype.UUID) error {
	return r.queries.ConsumeAuthActionToken(ctx, id)
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

func (r *PostgresRepository) RevokeAllRefreshTokensByUser(ctx context.Context, userID pgtype.UUID) error {
	return r.queries.RevokeAllRefreshTokensByUser(ctx, userID)
}

func (r *PostgresRepository) UpdateUserPasswordByID(ctx context.Context, params queries.UpdateUserPasswordByIDParams) (queries.User, error) {
	user, err := r.queries.UpdateUserPasswordByID(ctx, params)
	if errors.Is(err, pgx.ErrNoRows) {
		return queries.User{}, ErrNotFound
	}
	return user, err
}

func (r *PostgresRepository) MarkUserEmailVerified(ctx context.Context, id pgtype.UUID) (queries.User, error) {
	user, err := r.queries.MarkUserEmailVerified(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return queries.User{}, ErrNotFound
	}
	return user, err
}

func (r *PostgresRepository) DeactivateUser(ctx context.Context, id pgtype.UUID) (queries.User, error) {
	user, err := r.queries.DeactivateUser(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return queries.User{}, ErrNotFound
	}
	return user, err
}
