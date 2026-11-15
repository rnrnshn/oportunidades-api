package admin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rnrnshn/oportunidades-api/pkg/db/queries"
)

var (
	// ErrSelfRoleChange is returned when an admin tries to change their own role.
	ErrSelfRoleChange = errors.New("admin: cannot change own role")
	// ErrSelfSuspend is returned when an admin tries to suspend themselves.
	ErrSelfSuspend = errors.New("admin: cannot suspend self")
	// ErrAlreadySuspended / ErrNotSuspended guard suspend and unsuspend.
	ErrAlreadySuspended = errors.New("admin: already suspended")
	ErrNotSuspended     = errors.New("admin: not suspended")
)

var (
	UserRoles  = []string{"user", "mentor", "cms_partner", "admin"}
	UserScopes = []string{"team", "community"}
	UserStatus = []string{"active", "suspended", "unverified"}
	UserSorts  = []string{"name_asc", "name_desc", "created_at_asc", "last_seen_desc"}
)

type UserItem struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	Role          string `json:"role"`
	AvatarURL     string `json:"avatar_url,omitempty"`
	EmailVerified bool   `json:"email_verified"`
	CreatedAt     string `json:"created_at,omitempty"`
	Suspended     bool   `json:"suspended"`
	LastSeenAt    string `json:"last_seen_at,omitempty"`
	ContentCount  int64  `json:"content_count"`
}

type UserDetail struct {
	UserItem
	SuspendedAt      string         `json:"suspended_at,omitempty"`
	SuspensionReason string         `json:"suspension_reason,omitempty"`
	SuspendedBy      string         `json:"suspended_by,omitempty"`
	Content          ContentCounts  `json:"content"`
	Activity         []ActivityItem `json:"activity"`
}

type ContentCounts struct {
	Articles      int64 `json:"articles"`
	Opportunities int64 `json:"opportunities"`
	Universities  int64 `json:"universities"`
}

type ActivityItem struct {
	ID        string         `json:"id"`
	Action    string         `json:"action"`
	ActorName string         `json:"actor_name,omitempty"`
	Data      map[string]any `json:"data"`
	CreatedAt string         `json:"created_at"`
}

type UsersResult struct {
	Data []UserItem `json:"data"`
	Meta UsersMeta  `json:"meta"`
}

// UsersMeta adds per-tab and per-role totals (unfiltered) to the usual pagination meta.
type UsersMeta struct {
	PaginationMeta
	TabCounts  TabCounts        `json:"tab_counts"`
	RoleCounts map[string]int64 `json:"role_counts"`
}

type TabCounts struct {
	Team      int64 `json:"team"`
	Community int64 `json:"community"`
	Invites   int64 `json:"invites"`
}

type UserResult struct {
	Data UserItem `json:"data"`
}

type UserDetailResult struct {
	Data UserDetail `json:"data"`
}

type UserListFilters struct {
	Query  string
	Scope  string
	Role   string
	Status string
	Sort   string
}

type UpdateUserRoleInput struct {
	UserID  string
	Role    string
	ActorID string
}

type SuspendUserInput struct {
	UserID  string
	Reason  string
	ActorID string
}

func (s *Service) ListUsers(ctx context.Context, params PaginationParams, filters UserListFilters) (*UsersResult, error) {
	page, perPage := normalizePagination(params)
	scope, role, status, query := textToPg(filters.Scope), textToPg(filters.Role), textToPg(filters.Status), textToPg(filters.Query)

	total, err := s.repo.CountUsers(ctx, queries.AdminCountUsersParams{Scope: scope, Role: role, Status: status, Query: query})
	if err != nil {
		return nil, fmt.Errorf("admin: count users: %w", err)
	}
	rows, err := s.repo.ListUsers(ctx, queries.AdminListUsersParams{
		Scope:  scope,
		Role:   role,
		Status: status,
		Query:  query,
		Sort:   filters.Sort,
		Limit:  int32(perPage),
		Offset: int32((page - 1) * perPage),
	})
	if err != nil {
		return nil, fmt.Errorf("admin: list users: %w", err)
	}
	tabs, err := s.repo.UserTabCounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("admin: user tab counts: %w", err)
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
		item.Suspended = row.SuspendedAt.Valid
		item.LastSeenAt = timestamptzValue(row.LastSeenAt)
		item.ContentCount = row.ContentCount
		data = append(data, item)
	}
	counts := make(map[string]int64, len(UserRoles))
	for _, role := range UserRoles {
		counts[role] = 0
	}
	for _, row := range byRole {
		counts[row.Role] = row.Total
	}
	return &UsersResult{Data: data, Meta: UsersMeta{
		PaginationMeta: buildMeta(total, page, perPage),
		TabCounts:      TabCounts{Team: tabs.Team, Community: tabs.Community, Invites: tabs.Invites},
		RoleCounts:     counts,
	}}, nil
}

func (s *Service) GetUser(ctx context.Context, userID string) (*UserDetailResult, error) {
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
	item.Suspended = row.SuspendedAt.Valid
	item.LastSeenAt = timestamptzValue(row.LastSeenAt)
	item.ContentCount = row.ArticleCount + row.OpportunityCount + row.UniversityCount

	events, err := s.repo.ListAuditEvents(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("admin: list audit events: %w", err)
	}
	activity := make([]ActivityItem, 0, len(events))
	for _, event := range events {
		eventID, err := uuidFromPg(event.ID)
		if err != nil {
			continue
		}
		data := map[string]any{}
		_ = json.Unmarshal(event.Data, &data)
		activity = append(activity, ActivityItem{ID: eventID.String(), Action: event.Action, ActorName: textValue(event.ActorName), Data: data, CreatedAt: timestamptzValue(event.CreatedAt)})
	}

	return &UserDetailResult{Data: UserDetail{
		UserItem:         item,
		SuspendedAt:      timestamptzValue(row.SuspendedAt),
		SuspensionReason: textValue(row.SuspensionReason),
		SuspendedBy:      textValue(row.SuspendedByName),
		Content:          ContentCounts{Articles: row.ArticleCount, Opportunities: row.OpportunityCount, Universities: row.UniversityCount},
		Activity:         activity,
	}}, nil
}

// UpdateUserRole changes a user's role. Admins cannot change their own role and the
// last active admin cannot be demoted. The user's sessions end; their next request
// already uses the new role because the auth middleware reads it from the database.
func (s *Service) UpdateUserRole(ctx context.Context, input UpdateUserRoleInput) (*UserResult, error) {
	id, actor, err := s.targetAndActor(input.UserID, input.ActorID)
	if err != nil {
		return nil, err
	}
	if id == actor {
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
	row, err := s.repo.UpdateUserRole(ctx, queries.AdminUpdateUserRoleParams{ID: id, Role: input.Role}, AuditEntry{
		ActorID: actor,
		Action:  "role_changed",
		Target:  id,
		Data:    map[string]any{"from": current.Role, "to": input.Role},
	})
	if err != nil {
		return nil, err
	}
	item, err := mapUser(row.ID, row.Email, row.Name, row.Role, row.AvatarUrl, row.EmailVerifiedAt, row.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &UserResult{Data: item}, nil
}

// SuspendUser blocks an account: login and refresh are refused, sessions end and the
// auth middleware rejects it on the next request. Not allowed on yourself or the last admin.
func (s *Service) SuspendUser(ctx context.Context, input SuspendUserInput) error {
	id, actor, err := s.targetAndActor(input.UserID, input.ActorID)
	if err != nil {
		return err
	}
	if id == actor {
		return ErrSelfSuspend
	}
	current, err := s.repo.GetUser(ctx, id)
	if err != nil {
		return err
	}
	if current.SuspendedAt.Valid {
		return ErrAlreadySuspended
	}
	reason := strings.TrimSpace(input.Reason)
	return s.repo.SuspendUser(ctx, queries.AdminSuspendUserParams{ID: id, ActorID: actor, Reason: reason}, AuditEntry{
		ActorID: actor,
		Action:  "suspended",
		Target:  id,
		Data:    map[string]any{"reason": reason},
	})
}

func (s *Service) UnsuspendUser(ctx context.Context, userID, actorID string) error {
	id, actor, err := s.targetAndActor(userID, actorID)
	if err != nil {
		return err
	}
	if _, err := s.repo.GetUser(ctx, id); err != nil {
		return err
	}
	return s.repo.UnsuspendUser(ctx, id, AuditEntry{ActorID: actor, Action: "unsuspended", Target: id})
}

func (s *Service) targetAndActor(userID, actorID string) (pgtype.UUID, pgtype.UUID, error) {
	id, err := parseID(userID)
	if err != nil {
		return pgtype.UUID{}, pgtype.UUID{}, ErrNotFound
	}
	actor, err := parseID(actorID)
	if err != nil {
		return pgtype.UUID{}, pgtype.UUID{}, ErrNotFound
	}
	return id, actor, nil
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
