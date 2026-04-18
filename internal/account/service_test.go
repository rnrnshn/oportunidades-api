package account

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rnrnshn/oportunidades-api/pkg/db/queries"
	"golang.org/x/crypto/bcrypt"
)

type mockRepository struct {
	getUserByIDFn        func(context.Context, pgtype.UUID) (queries.User, error)
	updateUserProfileFn  func(context.Context, queries.UpdateUserProfileParams) (queries.User, error)
	updateUserPasswordFn func(context.Context, queries.UpdateUserPasswordParams) (queries.User, error)
}

func (m *mockRepository) GetUserByID(ctx context.Context, id pgtype.UUID) (queries.User, error) {
	return m.getUserByIDFn(ctx, id)
}
func (m *mockRepository) UpdateUserProfile(ctx context.Context, params queries.UpdateUserProfileParams) (queries.User, error) {
	return m.updateUserProfileFn(ctx, params)
}
func (m *mockRepository) UpdateUserPassword(ctx context.Context, params queries.UpdateUserPasswordParams) (queries.User, error) {
	return m.updateUserPasswordFn(ctx, params)
}

func TestGetProfile(t *testing.T) {
	userID := uuid.New()
	service := NewService(&mockRepository{getUserByIDFn: func(context.Context, pgtype.UUID) (queries.User, error) {
		return queries.User{ID: uuidToPg(userID), Email: "user@example.com", Role: "user", Name: "User"}, nil
	}})
	result, err := service.GetProfile(context.Background(), userID.String())
	if err != nil {
		t.Fatalf("get profile returned error: %v", err)
	}
	if result.Data.Email != "user@example.com" {
		t.Fatalf("unexpected email: %s", result.Data.Email)
	}
}

func TestUpdateProfileRequiresName(t *testing.T) {
	service := NewService(&mockRepository{})
	_, err := service.UpdateProfile(context.Background(), UpdateProfileInput{UserID: uuid.NewString(), Name: ""})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestGetProfileNotFound(t *testing.T) {
	service := NewService(&mockRepository{getUserByIDFn: func(context.Context, pgtype.UUID) (queries.User, error) { return queries.User{}, ErrNotFound }})
	_, err := service.GetProfile(context.Background(), uuid.NewString())
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestChangePassword(t *testing.T) {
	userID := uuid.New()
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("generate password hash: %v", err)
	}
	service := NewService(&mockRepository{
		getUserByIDFn: func(context.Context, pgtype.UUID) (queries.User, error) {
			return queries.User{ID: uuidToPg(userID), PasswordHash: string(passwordHash)}, nil
		},
		updateUserPasswordFn: func(context.Context, queries.UpdateUserPasswordParams) (queries.User, error) {
			return queries.User{ID: uuidToPg(userID)}, nil
		},
	})
	result, err := service.ChangePassword(context.Background(), ChangePasswordInput{UserID: userID.String(), CurrentPassword: "password", NewPassword: "newpassword"})
	if err != nil {
		t.Fatalf("change password returned error: %v", err)
	}
	if result.Data.Message == "" {
		t.Fatal("expected success message")
	}
}

func TestChangePasswordRejectsWrongCurrentPassword(t *testing.T) {
	userID := uuid.New()
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("generate password hash: %v", err)
	}
	service := NewService(&mockRepository{
		getUserByIDFn: func(context.Context, pgtype.UUID) (queries.User, error) {
			return queries.User{ID: uuidToPg(userID), PasswordHash: string(passwordHash)}, nil
		},
	})
	_, err = service.ChangePassword(context.Background(), ChangePasswordInput{UserID: userID.String(), CurrentPassword: "wrong", NewPassword: "newpassword"})
	if err == nil {
		t.Fatal("expected error")
	}
}
