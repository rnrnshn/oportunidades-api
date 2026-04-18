package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rnrnshn/oportunidades-api/pkg/db/queries"
	"golang.org/x/crypto/bcrypt"
)

type mockRepository struct {
	createUserFn                   func(context.Context, queries.CreateUserParams) (queries.User, error)
	getUserByEmailFn               func(context.Context, string) (queries.User, error)
	getUserByIDFn                  func(context.Context, pgtype.UUID) (queries.User, error)
	createRefreshTokenFn           func(context.Context, queries.CreateRefreshTokenParams) (queries.RefreshToken, error)
	getRefreshTokenFn              func(context.Context, string) (queries.RefreshToken, error)
	revokeRefreshTokenFn           func(context.Context, pgtype.UUID) error
	revokeAllRefreshTokensByUserFn func(context.Context, pgtype.UUID) error
}

func (m *mockRepository) CreateUser(ctx context.Context, params queries.CreateUserParams) (queries.User, error) {
	return m.createUserFn(ctx, params)
}

func (m *mockRepository) GetUserByEmail(ctx context.Context, email string) (queries.User, error) {
	return m.getUserByEmailFn(ctx, email)
}

func (m *mockRepository) GetUserByID(ctx context.Context, id pgtype.UUID) (queries.User, error) {
	return m.getUserByIDFn(ctx, id)
}

func (m *mockRepository) CreateRefreshToken(ctx context.Context, params queries.CreateRefreshTokenParams) (queries.RefreshToken, error) {
	return m.createRefreshTokenFn(ctx, params)
}

func (m *mockRepository) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (queries.RefreshToken, error) {
	return m.getRefreshTokenFn(ctx, tokenHash)
}

func (m *mockRepository) RevokeRefreshToken(ctx context.Context, id pgtype.UUID) error {
	return m.revokeRefreshTokenFn(ctx, id)
}

func (m *mockRepository) RevokeAllRefreshTokensByUser(ctx context.Context, userID pgtype.UUID) error {
	return m.revokeAllRefreshTokensByUserFn(ctx, userID)
}

func TestServiceLogin(t *testing.T) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("generate password hash: %v", err)
	}

	userID := uuid.New()
	repo := &mockRepository{
		getUserByEmailFn: func(context.Context, string) (queries.User, error) {
			return queries.User{
				ID:           uuidToPg(userID),
				Email:        "user@example.com",
				PasswordHash: string(passwordHash),
				Role:         "user",
				Name:         "User",
			}, nil
		},
		createRefreshTokenFn: func(context.Context, queries.CreateRefreshTokenParams) (queries.RefreshToken, error) {
			return queries.RefreshToken{}, nil
		},
	}

	service := NewService(repo, Config{
		JWTSecret:          "secret",
		JWTExpiry:          15 * time.Minute,
		RefreshTokenExpiry: 30 * 24 * time.Hour,
		RefreshCookieName:  "refresh_token",
	})

	result, err := service.Login(context.Background(), LoginInput{
		Email:    "user@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	if result.User.Email != "user@example.com" {
		t.Fatalf("unexpected user email: %s", result.User.Email)
	}

	if result.AccessToken == "" {
		t.Fatal("expected access token")
	}

	if result.RefreshToken == "" {
		t.Fatal("expected refresh token")
	}
}

func TestServiceLoginInvalidPassword(t *testing.T) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("generate password hash: %v", err)
	}

	repo := &mockRepository{
		getUserByEmailFn: func(context.Context, string) (queries.User, error) {
			return queries.User{PasswordHash: string(passwordHash)}, nil
		},
	}

	service := NewService(repo, Config{
		JWTSecret:          "secret",
		JWTExpiry:          15 * time.Minute,
		RefreshTokenExpiry: 30 * 24 * time.Hour,
		RefreshCookieName:  "refresh_token",
	})

	_, err = service.Login(context.Background(), LoginInput{
		Email:    "user@example.com",
		Password: "wrong-password",
	})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials error, got: %v", err)
	}
}

func TestServiceRefresh(t *testing.T) {
	userID := uuid.New()
	refreshID := uuid.New()
	rawRefreshToken := "refresh-token"
	revoked := false

	repo := &mockRepository{
		getRefreshTokenFn: func(context.Context, string) (queries.RefreshToken, error) {
			return queries.RefreshToken{
				ID:        uuidToPg(refreshID),
				UserID:    uuidToPg(userID),
				ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true},
			}, nil
		},
		getUserByIDFn: func(context.Context, pgtype.UUID) (queries.User, error) {
			return queries.User{
				ID:    uuidToPg(userID),
				Email: "user@example.com",
				Role:  "user",
				Name:  "User",
			}, nil
		},
		revokeRefreshTokenFn: func(context.Context, pgtype.UUID) error {
			revoked = true
			return nil
		},
		createRefreshTokenFn: func(context.Context, queries.CreateRefreshTokenParams) (queries.RefreshToken, error) {
			return queries.RefreshToken{}, nil
		},
	}

	service := NewService(repo, Config{
		JWTSecret:          "secret",
		JWTExpiry:          15 * time.Minute,
		RefreshTokenExpiry: 30 * 24 * time.Hour,
		RefreshCookieName:  "refresh_token",
	})

	result, err := service.Refresh(context.Background(), rawRefreshToken)
	if err != nil {
		t.Fatalf("refresh returned error: %v", err)
	}

	if result.RefreshToken == rawRefreshToken {
		t.Fatal("expected rotated refresh token")
	}
	if !revoked {
		t.Fatal("expected previous refresh token to be revoked")
	}
}

func TestServiceRefreshRejectsExpiredToken(t *testing.T) {
	userID := uuid.New()
	repo := &mockRepository{
		getRefreshTokenFn: func(context.Context, string) (queries.RefreshToken, error) {
			return queries.RefreshToken{UserID: uuidToPg(userID), ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(-time.Minute), Valid: true}}, nil
		},
	}
	service := NewService(repo, Config{JWTSecret: "secret", JWTExpiry: 15 * time.Minute, RefreshTokenExpiry: 30 * 24 * time.Hour, RefreshCookieName: "refresh_token"})
	_, err := service.Refresh(context.Background(), "expired-token")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials for expired token, got %v", err)
	}
}

func TestServiceLogoutAll(t *testing.T) {
	userID := uuid.New()
	called := false
	repo := &mockRepository{
		revokeAllRefreshTokensByUserFn: func(_ context.Context, id pgtype.UUID) error {
			called = true
			if id != uuidToPg(userID) {
				t.Fatalf("unexpected user id: %+v", id)
			}
			return nil
		},
	}
	service := NewService(repo, Config{JWTSecret: "secret", JWTExpiry: 15 * time.Minute, RefreshTokenExpiry: 30 * 24 * time.Hour, RefreshCookieName: "refresh_token"})
	if err := service.LogoutAll(context.Background(), userID.String()); err != nil {
		t.Fatalf("logout all returned error: %v", err)
	}
	if !called {
		t.Fatal("expected revoke all refresh tokens call")
	}
}

func uuidToPg(value uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(value), Valid: true}
}
