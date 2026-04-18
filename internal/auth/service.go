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

type Service struct {
	repo                Repository
	jwtSecret           []byte
	jwtExpiry           time.Duration
	refreshTokenExpiry  time.Duration
	refreshCookieName   string
	refreshCookieSecure bool
}

type Config struct {
	JWTSecret           string
	JWTExpiry           time.Duration
	RefreshTokenExpiry  time.Duration
	RefreshCookieName   string
	RefreshCookieSecure bool
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

type AuthResult struct {
	AccessToken  string
	RefreshToken string
	User         UserProfile
	ExpiresIn    int64
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
		repo:                repo,
		jwtSecret:           []byte(config.JWTSecret),
		jwtExpiry:           config.JWTExpiry,
		refreshTokenExpiry:  config.RefreshTokenExpiry,
		refreshCookieName:   config.RefreshCookieName,
		refreshCookieSecure: config.RefreshCookieSecure,
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
