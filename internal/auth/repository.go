package auth

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rnrnshn/oportunidades-api/pkg/db/queries"
)

var ErrNotFound = errors.New("auth: not found")

type Repository interface {
	CreateUser(ctx context.Context, params queries.CreateUserParams) (queries.User, error)
	GetUserByEmail(ctx context.Context, email string) (queries.User, error)
	GetUserByID(ctx context.Context, id pgtype.UUID) (queries.User, error)
	CreateAuthActionToken(ctx context.Context, params queries.CreateAuthActionTokenParams) (queries.AuthActionToken, error)
	GetAuthActionTokenByHash(ctx context.Context, params queries.GetAuthActionTokenByHashParams) (queries.AuthActionToken, error)
	ConsumeAuthActionToken(ctx context.Context, id pgtype.UUID) error
	CreateRefreshToken(ctx context.Context, params queries.CreateRefreshTokenParams) (queries.RefreshToken, error)
	GetRefreshTokenByHash(ctx context.Context, tokenHash string) (queries.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, id pgtype.UUID) error
	RevokeAllRefreshTokensByUser(ctx context.Context, userID pgtype.UUID) error
	UpdateUserPasswordByID(ctx context.Context, params queries.UpdateUserPasswordByIDParams) (queries.User, error)
	MarkUserEmailVerified(ctx context.Context, id pgtype.UUID) (queries.User, error)
	DeactivateUser(ctx context.Context, id pgtype.UUID) (queries.User, error)
}
