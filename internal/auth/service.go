package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"

	"github.com/rnrnshn/oportunidades-api/pkg/db/queries"
)

var ErrInvalidCredentials = errors.New("auth: invalid credentials")

const (
	purposePasswordReset     = "password_reset"
	purposeEmailVerification = "email_verification"
)

type Service struct {
	repo                  Repository
	jwtSecret             []byte
	jwtExpiry             time.Duration
	refreshTokenExpiry    time.Duration
	authActionTokenExpiry time.Duration
	refreshCookieName     string
	refreshCookieSecure   bool
	exposeDebugTokens     bool
}

type Config struct {
	JWTSecret             string
	JWTExpiry             time.Duration
	RefreshTokenExpiry    time.Duration
	AuthActionTokenExpiry time.Duration
	RefreshCookieName     string
	RefreshCookieSecure   bool
	ExposeDebugTokens     bool
}

type RegisterInput struct {
	Email    string
	Password string
	Name     string
}

type LoginInput struct {
	Email    string
	Password string
}

type ForgotPasswordInput struct{ Email string }
type ResetPasswordInput struct {
	Token       string
	NewPassword string
}
type SendVerificationInput struct{ UserID string }
type VerifyEmailInput struct{ Token string }

type AuthResult struct {
	AccessToken  string
	RefreshToken string
	User         UserProfile
	ExpiresIn    int64
}

type ActionResult struct {
	Message    string
	DebugToken string
}

type UserProfile struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

type Claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func NewService(repo Repository, config Config) *Service {
	return &Service{
		repo:                  repo,
		jwtSecret:             []byte(config.JWTSecret),
		jwtExpiry:             config.JWTExpiry,
		refreshTokenExpiry:    config.RefreshTokenExpiry,
		authActionTokenExpiry: config.AuthActionTokenExpiry,
		refreshCookieName:     config.RefreshCookieName,
		refreshCookieSecure:   config.RefreshCookieSecure,
		exposeDebugTokens:     config.ExposeDebugTokens,
	}
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (*AuthResult, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("auth: hash password: %w", err)
	}

	user, err := s.repo.CreateUser(ctx, queries.CreateUserParams{
		Email:        input.Email,
		PasswordHash: string(passwordHash),
		Role:         "user",
		Name:         input.Name,
		AvatarUrl:    pgtype.Text{},
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, fmt.Errorf("auth: email already exists: %w", err)
		}

		return nil, fmt.Errorf("auth: create user: %w", err)
	}

	return s.issueTokens(ctx, user)
}

func (s *Service) Login(ctx context.Context, input LoginInput) (*AuthResult, error) {
	user, err := s.repo.GetUserByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrInvalidCredentials
		}

		return nil, fmt.Errorf("auth: get user by email: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.issueTokens(ctx, user)
}

func (s *Service) Refresh(ctx context.Context, rawRefreshToken string) (*AuthResult, error) {
	tokenHash := hashToken(rawRefreshToken)
	storedToken, err := s.repo.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrInvalidCredentials
		}

		return nil, fmt.Errorf("auth: get refresh token: %w", err)
	}

	if !storedToken.ExpiresAt.Valid || time.Now().After(storedToken.ExpiresAt.Time) {
		return nil, ErrInvalidCredentials
	}

	user, err := s.repo.GetUserByID(ctx, storedToken.UserID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrInvalidCredentials
		}

		return nil, fmt.Errorf("auth: get user by id: %w", err)
	}

	if err := s.repo.RevokeRefreshToken(ctx, storedToken.ID); err != nil {
		return nil, fmt.Errorf("auth: revoke refresh token: %w", err)
	}

	return s.issueTokens(ctx, user)
}

func (s *Service) Logout(ctx context.Context, rawRefreshToken string) error {
	if rawRefreshToken == "" {
		return nil
	}

	tokenHash := hashToken(rawRefreshToken)
	storedToken, err := s.repo.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil
		}

		return fmt.Errorf("auth: get refresh token for logout: %w", err)
	}

	if err := s.repo.RevokeRefreshToken(ctx, storedToken.ID); err != nil {
		return fmt.Errorf("auth: revoke refresh token for logout: %w", err)
	}

	return nil
}

func (s *Service) LogoutAll(ctx context.Context, userID string) error {
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("auth: invalid user id for logout all: %w", err)
	}
	if err := s.repo.RevokeAllRefreshTokensByUser(ctx, pgtype.UUID{Bytes: [16]byte(parsedUserID), Valid: true}); err != nil {
		return fmt.Errorf("auth: revoke all refresh tokens: %w", err)
	}
	return nil
}

func (s *Service) ForgotPassword(ctx context.Context, input ForgotPasswordInput) (*ActionResult, error) {
	user, err := s.repo.GetUserByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return &ActionResult{Message: "Se existir uma conta com este email, enviámos instruções para redefinir a password."}, nil
		}
		return nil, fmt.Errorf("auth: get user for forgot password: %w", err)
	}
	rawToken, err := s.createActionToken(ctx, user.ID, purposePasswordReset)
	if err != nil {
		return nil, err
	}
	result := &ActionResult{Message: "Se existir uma conta com este email, enviámos instruções para redefinir a password."}
	if s.exposeDebugTokens {
		result.DebugToken = rawToken
	}
	return result, nil
}

func (s *Service) ResetPassword(ctx context.Context, input ResetPasswordInput) (*ActionResult, error) {
	if len(input.NewPassword) < 8 {
		return nil, fmt.Errorf("auth: new password must be at least 8 characters")
	}
	actionToken, err := s.getValidActionToken(ctx, input.Token, purposePasswordReset)
	if err != nil {
		return nil, err
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("auth: hash reset password: %w", err)
	}
	if _, err := s.repo.UpdateUserPasswordByID(ctx, queries.UpdateUserPasswordByIDParams{ID: actionToken.UserID, PasswordHash: string(passwordHash)}); err != nil {
		return nil, fmt.Errorf("auth: update password by token: %w", err)
	}
	if err := s.repo.ConsumeAuthActionToken(ctx, actionToken.ID); err != nil {
		return nil, fmt.Errorf("auth: consume reset token: %w", err)
	}
	if err := s.repo.RevokeAllRefreshTokensByUser(ctx, actionToken.UserID); err != nil {
		return nil, fmt.Errorf("auth: revoke refresh tokens after reset: %w", err)
	}
	return &ActionResult{Message: "Password redefinida com sucesso."}, nil
}

func (s *Service) SendVerification(ctx context.Context, input SendVerificationInput) (*ActionResult, error) {
	parsedUserID, err := uuid.Parse(input.UserID)
	if err != nil {
		return nil, fmt.Errorf("auth: invalid user id for verification: %w", err)
	}
	user, err := s.repo.GetUserByID(ctx, pgtype.UUID{Bytes: [16]byte(parsedUserID), Valid: true})
	if err != nil {
		return nil, fmt.Errorf("auth: get user for verification: %w", err)
	}
	if user.EmailVerifiedAt.Valid {
		return &ActionResult{Message: "Email já verificado."}, nil
	}
	rawToken, err := s.createActionToken(ctx, user.ID, purposeEmailVerification)
	if err != nil {
		return nil, err
	}
	result := &ActionResult{Message: "Instruções de verificação preparadas com sucesso."}
	if s.exposeDebugTokens {
		result.DebugToken = rawToken
	}
	return result, nil
}

func (s *Service) VerifyEmail(ctx context.Context, input VerifyEmailInput) (*ActionResult, error) {
	actionToken, err := s.getValidActionToken(ctx, input.Token, purposeEmailVerification)
	if err != nil {
		return nil, err
	}
	if _, err := s.repo.MarkUserEmailVerified(ctx, actionToken.UserID); err != nil {
		return nil, fmt.Errorf("auth: mark email verified: %w", err)
	}
	if err := s.repo.ConsumeAuthActionToken(ctx, actionToken.ID); err != nil {
		return nil, fmt.Errorf("auth: consume verification token: %w", err)
	}
	return &ActionResult{Message: "Email verificado com sucesso."}, nil
}

func (s *Service) DeactivateAccount(ctx context.Context, userID string, currentPassword string) (*ActionResult, error) {
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("auth: invalid user id for deactivate: %w", err)
	}
	user, err := s.repo.GetUserByID(ctx, pgtype.UUID{Bytes: [16]byte(parsedUserID), Valid: true})
	if err != nil {
		return nil, fmt.Errorf("auth: get user for deactivate: %w", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)); err != nil {
		return nil, fmt.Errorf("auth: current password is invalid")
	}
	if _, err := s.repo.DeactivateUser(ctx, user.ID); err != nil {
		return nil, fmt.Errorf("auth: deactivate user: %w", err)
	}
	if err := s.repo.RevokeAllRefreshTokensByUser(ctx, user.ID); err != nil {
		return nil, fmt.Errorf("auth: revoke refresh tokens after deactivate: %w", err)
	}
	return &ActionResult{Message: "Conta desactivada com sucesso."}, nil
}

func (s *Service) ParseAccessToken(rawToken string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(rawToken, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("auth: unexpected signing method")
		}

		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidCredentials
	}

	return claims, nil
}

func (s *Service) RefreshCookieName() string {
	return s.refreshCookieName
}

func (s *Service) RefreshCookieMaxAge() int {
	return int(s.refreshTokenExpiry.Seconds())
}

func (s *Service) RefreshCookieSecure() bool {
	return s.refreshCookieSecure
}

func (s *Service) issueTokens(ctx context.Context, user queries.User) (*AuthResult, error) {
	userID, err := uuidFromPg(user.ID)
	if err != nil {
		return nil, fmt.Errorf("auth: user id: %w", err)
	}

	accessToken, err := s.createAccessToken(userID.String(), user.Role)
	if err != nil {
		return nil, fmt.Errorf("auth: create access token: %w", err)
	}

	refreshToken, refreshHash, err := generateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("auth: create refresh token: %w", err)
	}

	_, err = s.repo.CreateRefreshToken(ctx, queries.CreateRefreshTokenParams{
		UserID:    user.ID,
		TokenHash: refreshHash,
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(s.refreshTokenExpiry), Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("auth: store refresh token: %w", err)
	}

	return &AuthResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         profileFromUser(user, userID.String()),
		ExpiresIn:    int64(s.jwtExpiry.Seconds()),
	}, nil
}

func (s *Service) createAccessToken(userID string, role string) (string, error) {
	now := time.Now()
	claims := Claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.jwtExpiry)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func generateRefreshToken() (string, string, error) {
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", "", err
	}

	rawToken := base64.RawURLEncoding.EncodeToString(randomBytes)
	return rawToken, hashToken(rawToken), nil
}

func hashToken(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return fmt.Sprintf("%x", sum[:])
}

func (s *Service) createActionToken(ctx context.Context, userID pgtype.UUID, purpose string) (string, error) {
	rawToken, tokenHash, err := generateRefreshToken()
	if err != nil {
		return "", fmt.Errorf("auth: create action token: %w", err)
	}
	_, err = s.repo.CreateAuthActionToken(ctx, queries.CreateAuthActionTokenParams{
		UserID:    userID,
		Purpose:   purpose,
		TokenHash: tokenHash,
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(s.authActionTokenExpiry), Valid: true},
	})
	if err != nil {
		return "", fmt.Errorf("auth: store action token: %w", err)
	}
	return rawToken, nil
}

func (s *Service) getValidActionToken(ctx context.Context, rawToken string, purpose string) (queries.AuthActionToken, error) {
	tokenHash := hashToken(rawToken)
	actionToken, err := s.repo.GetAuthActionTokenByHash(ctx, queries.GetAuthActionTokenByHashParams{TokenHash: tokenHash, Purpose: purpose})
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return queries.AuthActionToken{}, ErrInvalidCredentials
		}
		return queries.AuthActionToken{}, fmt.Errorf("auth: get action token: %w", err)
	}
	if !actionToken.ExpiresAt.Valid || time.Now().After(actionToken.ExpiresAt.Time) {
		return queries.AuthActionToken{}, ErrInvalidCredentials
	}
	return actionToken, nil
}

func profileFromUser(user queries.User, userID string) UserProfile {
	profile := UserProfile{
		ID:    userID,
		Email: user.Email,
		Role:  user.Role,
		Name:  user.Name,
	}

	if user.AvatarUrl.Valid {
		profile.AvatarURL = user.AvatarUrl.String
	}

	return profile
}

func uuidFromPg(value pgtype.UUID) (uuid.UUID, error) {
	if !value.Valid {
		return uuid.Nil, fmt.Errorf("invalid uuid")
	}

	return uuid.UUID(value.Bytes), nil
}
