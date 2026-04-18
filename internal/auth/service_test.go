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
	createAuthActionTokenFn        func(context.Context, queries.CreateAuthActionTokenParams) (queries.AuthActionToken, error)
	getAuthActionTokenByHashFn     func(context.Context, queries.GetAuthActionTokenByHashParams) (queries.AuthActionToken, error)
	consumeAuthActionTokenFn       func(context.Context, pgtype.UUID) error
	createRefreshTokenFn           func(context.Context, queries.CreateRefreshTokenParams) (queries.RefreshToken, error)
	getRefreshTokenFn              func(context.Context, string) (queries.RefreshToken, error)
	revokeRefreshTokenFn           func(context.Context, pgtype.UUID) error
	revokeAllRefreshTokensByUserFn func(context.Context, pgtype.UUID) error
	updateUserPasswordByIDFn       func(context.Context, queries.UpdateUserPasswordByIDParams) (queries.User, error)
	markUserEmailVerifiedFn        func(context.Context, pgtype.UUID) (queries.User, error)
	deactivateUserFn               func(context.Context, pgtype.UUID) (queries.User, error)
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

func (m *mockRepository) CreateAuthActionToken(ctx context.Context, params queries.CreateAuthActionTokenParams) (queries.AuthActionToken, error) {
	return m.createAuthActionTokenFn(ctx, params)
}

func (m *mockRepository) GetAuthActionTokenByHash(ctx context.Context, params queries.GetAuthActionTokenByHashParams) (queries.AuthActionToken, error) {
	return m.getAuthActionTokenByHashFn(ctx, params)
}

func (m *mockRepository) ConsumeAuthActionToken(ctx context.Context, id pgtype.UUID) error {
	return m.consumeAuthActionTokenFn(ctx, id)
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

func (m *mockRepository) UpdateUserPasswordByID(ctx context.Context, params queries.UpdateUserPasswordByIDParams) (queries.User, error) {
	return m.updateUserPasswordByIDFn(ctx, params)
}

func (m *mockRepository) MarkUserEmailVerified(ctx context.Context, id pgtype.UUID) (queries.User, error) {
	return m.markUserEmailVerifiedFn(ctx, id)
}

func (m *mockRepository) DeactivateUser(ctx context.Context, id pgtype.UUID) (queries.User, error) {
	return m.deactivateUserFn(ctx, id)
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

func TestServiceForgotPasswordReturnsDebugTokenForExistingUser(t *testing.T) {
	userID := uuid.New()
	service := NewService(&mockRepository{
		getUserByEmailFn: func(context.Context, string) (queries.User, error) {
			return queries.User{ID: uuidToPg(userID), Email: "user@example.com", Name: "User"}, nil
		},
		createAuthActionTokenFn: func(context.Context, queries.CreateAuthActionTokenParams) (queries.AuthActionToken, error) {
			return queries.AuthActionToken{}, nil
		},
	}, Config{JWTSecret: "secret", JWTExpiry: 15 * time.Minute, RefreshTokenExpiry: 30 * 24 * time.Hour, AuthActionTokenExpiry: 24 * time.Hour, RefreshCookieName: "refresh_token", ExposeDebugTokens: true})
	result, err := service.ForgotPassword(context.Background(), ForgotPasswordInput{Email: "user@example.com"})
	if err != nil {
		t.Fatalf("forgot password returned error: %v", err)
	}
	if result.DebugToken == "" {
		t.Fatal("expected debug token")
	}
}

func TestServiceResetPassword(t *testing.T) {
	userID := uuid.New()
	actionTokenID := uuid.New()
	consumed := false
	revokedAll := false
	service := NewService(&mockRepository{
		getAuthActionTokenByHashFn: func(context.Context, queries.GetAuthActionTokenByHashParams) (queries.AuthActionToken, error) {
			return queries.AuthActionToken{ID: uuidToPg(actionTokenID), UserID: uuidToPg(userID), Purpose: purposePasswordReset, ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true}}, nil
		},
		updateUserPasswordByIDFn: func(context.Context, queries.UpdateUserPasswordByIDParams) (queries.User, error) {
			return queries.User{ID: uuidToPg(userID)}, nil
		},
		consumeAuthActionTokenFn:       func(context.Context, pgtype.UUID) error { consumed = true; return nil },
		revokeAllRefreshTokensByUserFn: func(context.Context, pgtype.UUID) error { revokedAll = true; return nil },
	}, Config{JWTSecret: "secret", JWTExpiry: 15 * time.Minute, RefreshTokenExpiry: 30 * 24 * time.Hour, AuthActionTokenExpiry: 24 * time.Hour, RefreshCookieName: "refresh_token"})
	result, err := service.ResetPassword(context.Background(), ResetPasswordInput{Token: "reset-token", NewPassword: "newpassword"})
	if err != nil {
		t.Fatalf("reset password returned error: %v", err)
	}
	if result.Message == "" || !consumed || !revokedAll {
		t.Fatalf("expected success and side effects, got %+v consumed=%v revokedAll=%v", result, consumed, revokedAll)
	}
}

func TestServiceVerifyEmail(t *testing.T) {
	userID := uuid.New()
	actionTokenID := uuid.New()
	verified := false
	service := NewService(&mockRepository{
		getAuthActionTokenByHashFn: func(context.Context, queries.GetAuthActionTokenByHashParams) (queries.AuthActionToken, error) {
			return queries.AuthActionToken{ID: uuidToPg(actionTokenID), UserID: uuidToPg(userID), Purpose: purposeEmailVerification, ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true}}, nil
		},
		markUserEmailVerifiedFn: func(context.Context, pgtype.UUID) (queries.User, error) {
			verified = true
			return queries.User{ID: uuidToPg(userID)}, nil
		},
		consumeAuthActionTokenFn: func(context.Context, pgtype.UUID) error { return nil },
	}, Config{JWTSecret: "secret", JWTExpiry: 15 * time.Minute, RefreshTokenExpiry: 30 * 24 * time.Hour, AuthActionTokenExpiry: 24 * time.Hour, RefreshCookieName: "refresh_token"})
	result, err := service.VerifyEmail(context.Background(), VerifyEmailInput{Token: "verify-token"})
	if err != nil || !verified || result.Message == "" {
		t.Fatalf("verify email failed: result=%+v err=%v verified=%v", result, err, verified)
	}
}

func TestServiceDeactivateAccount(t *testing.T) {
	userID := uuid.New()
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("generate hash: %v", err)
	}
	deactivated := false
	revokedAll := false
	service := NewService(&mockRepository{
		getUserByIDFn: func(context.Context, pgtype.UUID) (queries.User, error) {
			return queries.User{ID: uuidToPg(userID), PasswordHash: string(passwordHash)}, nil
		},
		deactivateUserFn: func(context.Context, pgtype.UUID) (queries.User, error) {
			deactivated = true
			return queries.User{ID: uuidToPg(userID)}, nil
		},
		revokeAllRefreshTokensByUserFn: func(context.Context, pgtype.UUID) error { revokedAll = true; return nil },
	}, Config{JWTSecret: "secret", JWTExpiry: 15 * time.Minute, RefreshTokenExpiry: 30 * 24 * time.Hour, AuthActionTokenExpiry: 24 * time.Hour, RefreshCookieName: "refresh_token"})
	result, err := service.DeactivateAccount(context.Background(), userID.String(), "password")
	if err != nil || !deactivated || !revokedAll || result.Message == "" {
		t.Fatalf("deactivate failed: result=%+v err=%v deactivated=%v revoked=%v", result, err, deactivated, revokedAll)
	}
}

func uuidToPg(value uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(value), Valid: true}
}
