package account

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rnrnshn/oportunidades-api/pkg/db/queries"
	"golang.org/x/crypto/bcrypt"
)

type Service struct{ repo Repository }

type ProfileResult struct {
	Data Profile `json:"data"`
}

type Profile struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

type UpdateProfileInput struct {
	UserID    string
	Name      string
	AvatarURL string
}

type ChangePasswordInput struct {
	UserID          string
	CurrentPassword string
	NewPassword     string
}

type MessageResult struct {
	Data struct {
		Message string `json:"message"`
	} `json:"data"`
}

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) GetProfile(ctx context.Context, userID string) (*ProfileResult, error) {
	parsedID, err := uuid.Parse(strings.TrimSpace(userID))
	if err != nil {
		return nil, ErrNotFound
	}
	user, err := s.repo.GetUserByID(ctx, uuidToPg(parsedID))
	if err != nil {
		return nil, err
	}
	profile, err := mapProfile(user)
	if err != nil {
		return nil, err
	}
	return &ProfileResult{Data: profile}, nil
}

func (s *Service) UpdateProfile(ctx context.Context, input UpdateProfileInput) (*ProfileResult, error) {
	parsedID, err := uuid.Parse(strings.TrimSpace(input.UserID))
	if err != nil {
		return nil, ErrNotFound
	}
	if strings.TrimSpace(input.Name) == "" {
		return nil, fmt.Errorf("account: name is required")
	}
	user, err := s.repo.UpdateUserProfile(ctx, queries.UpdateUserProfileParams{
		ID:        uuidToPg(parsedID),
		Name:      strings.TrimSpace(input.Name),
		AvatarUrl: textToPg(strings.TrimSpace(input.AvatarURL)),
	})
	if err != nil {
		return nil, err
	}
	profile, err := mapProfile(user)
	if err != nil {
		return nil, err
	}
	return &ProfileResult{Data: profile}, nil
}

func (s *Service) ChangePassword(ctx context.Context, input ChangePasswordInput) (*MessageResult, error) {
	parsedID, err := uuid.Parse(strings.TrimSpace(input.UserID))
	if err != nil {
		return nil, ErrNotFound
	}
	if strings.TrimSpace(input.CurrentPassword) == "" || strings.TrimSpace(input.NewPassword) == "" {
		return nil, fmt.Errorf("account: current and new passwords are required")
	}
	if len(input.NewPassword) < 8 {
		return nil, fmt.Errorf("account: new password must be at least 8 characters")
	}
	user, err := s.repo.GetUserByID(ctx, uuidToPg(parsedID))
	if err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.CurrentPassword)); err != nil {
		return nil, fmt.Errorf("account: current password is invalid")
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("account: hash password: %w", err)
	}
	if _, err := s.repo.UpdateUserPassword(ctx, queries.UpdateUserPasswordParams{
		ID:           uuidToPg(parsedID),
		PasswordHash: string(passwordHash),
	}); err != nil {
		return nil, err
	}
	result := &MessageResult{}
	result.Data.Message = "Password actualizada com sucesso."
	return result, nil
}

func mapProfile(user queries.User) (Profile, error) {
	id, err := uuidFromPg(user.ID)
	if err != nil {
		return Profile{}, fmt.Errorf("account: user id: %w", err)
	}
	profile := Profile{ID: id.String(), Email: user.Email, Role: user.Role, Name: user.Name}
	if user.AvatarUrl.Valid {
		profile.AvatarURL = user.AvatarUrl.String
	}
	return profile, nil
}

func uuidToPg(value uuid.UUID) pgtype.UUID { return pgtype.UUID{Bytes: [16]byte(value), Valid: true} }

func uuidFromPg(value pgtype.UUID) (uuid.UUID, error) {
	if !value.Valid {
		return uuid.Nil, fmt.Errorf("invalid uuid")
	}
	return uuid.UUID(value.Bytes), nil
}

func textToPg(value string) pgtype.Text {
	if value == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: value, Valid: true}
}
