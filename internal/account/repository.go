package account

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rnrnshn/oportunidades-api/pkg/db/queries"
)

var ErrNotFound = errors.New("account: not found")

type Repository interface {
	GetUserByID(ctx context.Context, id pgtype.UUID) (queries.User, error)
	UpdateUserProfile(ctx context.Context, params queries.UpdateUserProfileParams) (queries.User, error)
	UpdateUserPassword(ctx context.Context, params queries.UpdateUserPasswordParams) (queries.User, error)
}
