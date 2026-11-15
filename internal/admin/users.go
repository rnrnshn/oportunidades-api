package admin

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rnrnshn/oportunidades-api/pkg/db/queries"
)

// ErrSelfRoleChange is returned when an admin tries to change their own role.
var ErrSelfRoleChange = errors.New("admin: cannot change own role")

var UserRoles = []string{"user", "mentor", "cms_partner", "admin"}

type UserItem struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	Role          string `json:"role"`
	AvatarURL     string `json:"avatar_url,omitempty"`
	EmailVerified bool   `json:"email_verified"`
	CreatedAt     string `json:"created_at,omitempty"`
}

type UsersResult struct {
	Data []UserItem `json:"data"`
	Meta UsersMeta  `json:"meta"`
}

// UsersMeta adds per-role totals (all users, unfiltered) to the usual pagination meta.
type UsersMeta struct {
	PaginationMeta
	RoleCounts map[string]int64 `json:"role_counts"`
}

type UserResult struct {
	Data UserItem `json:"data"`
}

type UserListFilters struct {
	Query string
	Role  string
	Sort  string
}

type UpdateUserRoleInput struct {
	UserID  string
	Role    string
	ActorID string
}

func (s *Service) ListUsers(ctx context.Context, params PaginationParams, filters UserListFilters) (*UsersResult, error) {
	page, perPage := normalizePagination(params)
	role := textToPg(filters.Role)
	query := textToPg(filters.Query)

	total, err := s.repo.CountUsers(ctx, queries.AdminCountUsersParams{Role: role, Query: query})
	if err != nil {
		return nil, fmt.Errorf("admin: count users: %w", err)
	}
	rows, err := s.repo.ListUsers(ctx, queries.AdminListUsersParams{
		Role:   role,
		Query:  query,
		Sort:   filters.Sort,
		Limit:  int32(perPage),
		Offset: int32((page - 1) * perPage),
	})
	if err != nil {
		return nil, fmt.Errorf("admin: list users: %w", err)
	}
	byRole, err := s.repo.CountUsersByRole(ctx)
	if err != nil {
		return nil, fmt.Errorf("admin: count users by role: %w", err)
	}

	data := make([]UserItem, 0, len(rows))
	for _, row := range rows {
		item, err := mapUser(row.ID, row.Email, row.Name, row.Role, row.AvatarUrl, row.EmailVerifiedAt, row.CreatedAt)
		if err != nil {
			return nil, err
		}
		data = append(data, item)
	}
	counts := make(map[string]int64, len(UserRoles))
	for _, role := range UserRoles {
		counts[role] = 0
	}
	for _, row := range byRole {
		counts[row.Role] = row.Total
	}
	return &UsersResult{Data: data, Meta: UsersMeta{PaginationMeta: buildMeta(total, page, perPage), RoleCounts: counts}}, nil
}

func (s *Service) GetUser(ctx context.Context, userID string) (*UserResult, error) {
	id, err := parseID(userID)
	if err != nil {
		return nil, ErrNotFound
	}
	row, err := s.repo.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}
	item, err := mapUser(row.ID, row.Email, row.Name, row.Role, row.AvatarUrl, row.EmailVerifiedAt, row.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &UserResult{Data: item}, nil
}

// UpdateUserRole changes a user's role. Admins cannot change their own role,
// and the last admin cannot be demoted. The user's sessions are ended so the
// new role applies at their next login (an access token already issued keeps
// the old role until it expires).
func (s *Service) UpdateUserRole(ctx context.Context, input UpdateUserRoleInput) (*UserResult, error) {
	id, err := parseID(input.UserID)
	if err != nil {
		return nil, ErrNotFound
	}
	if strings.EqualFold(strings.TrimSpace(input.UserID), strings.TrimSpace(input.ActorID)) {
		return nil, ErrSelfRoleChange
	}
	current, err := s.repo.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}
	if current.Role == input.Role {
		item, err := mapUser(current.ID, current.Email, current.Name, current.Role, current.AvatarUrl, current.EmailVerifiedAt, current.CreatedAt)
		if err != nil {
			return nil, err
		}
		return &UserResult{Data: item}, nil
	}
	row, err := s.repo.UpdateUserRole(ctx, queries.AdminUpdateUserRoleParams{ID: id, Role: input.Role})
	if err != nil {
		return nil, err
	}
	item, err := mapUser(row.ID, row.Email, row.Name, row.Role, row.AvatarUrl, row.EmailVerifiedAt, row.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &UserResult{Data: item}, nil
}

func mapUser(id pgtype.UUID, email, name, role string, avatar pgtype.Text, verifiedAt, createdAt pgtype.Timestamptz) (UserItem, error) {
	parsed, err := uuidFromPg(id)
	if err != nil {
		return UserItem{}, fmt.Errorf("admin: user id: %w", err)
	}
	return UserItem{
		ID:            parsed.String(),
		Email:         email,
		Name:          name,
		Role:          role,
		AvatarURL:     textValue(avatar),
		EmailVerified: verifiedAt.Valid,
		CreatedAt:     timestamptzValue(createdAt),
	}, nil
}
