package admin

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rnrnshn/oportunidades-api/pkg/db/queries"
	appemail "github.com/rnrnshn/oportunidades-api/pkg/email"
	"golang.org/x/crypto/bcrypt"
)

const inviteTTL = 7 * 24 * time.Hour

var (
	ErrEmailInUse    = errors.New("admin: email already has an account")
	ErrInvitePending = errors.New("admin: invite already pending")
	ErrInviteInvalid = errors.New("admin: invite invalid")
	ErrInviteExpired = errors.New("admin: invite expired")
	ErrInviteInput   = errors.New("admin: invalid invite input")
	InviteRoles      = []string{"admin", "cms_partner"}
	inviteRoleLabels = map[string]string{"admin": "administrador", "cms_partner": "parceiro de conteúdo"}
)

// Option configures optional collaborators of the admin service.
type Option func(*Service)

// WithInvites enables invite emails; links point to dashboardURL/convite/<token>.
func WithInvites(sender appemail.Sender, dashboardURL string) Option {
	return func(s *Service) {
		s.emailSender = sender
		s.dashboardURL = strings.TrimRight(dashboardURL, "/")
	}
}

type InviteItem struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	InvitedBy string `json:"invited_by,omitempty"`
	ExpiresAt string `json:"expires_at"`
	CreatedAt string `json:"created_at"`
	Expired   bool   `json:"expired"`
}

// InviteLinkResult is returned when a link is issued (create/resend). The link is
// shown once to the admin so it can be shared if the email doesn't arrive.
type InviteLinkResult struct {
	Data struct {
		Invite    InviteItem `json:"invite"`
		InviteURL string     `json:"invite_url"`
		EmailSent bool       `json:"email_sent"`
	} `json:"data"`
}

type InvitesResult struct {
	Data []InviteItem `json:"data"`
}

type PublicInviteResult struct {
	Data struct {
		Email     string `json:"email"`
		Name      string `json:"name"`
		Role      string `json:"role"`
		ExpiresAt string `json:"expires_at"`
	} `json:"data"`
}

type AcceptInviteResult struct {
	Data struct {
		Email string `json:"email"`
	} `json:"data"`
}

type CreateInviteInput struct {
	Email   string
	Name    string
	Role    string
	ActorID string
}

func (s *Service) CreateInvite(ctx context.Context, input CreateInviteInput) (*InviteLinkResult, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	name := strings.TrimSpace(input.Name)
	if _, err := mail.ParseAddress(email); err != nil || name == "" || !contains(InviteRoles, input.Role) {
		return nil, ErrInviteInput
	}
	actor, err := parseID(input.ActorID)
	if err != nil {
		return nil, ErrNotFound
	}
	// One account per email: existing accounts get a role change, not an invite.
	if exists, err := s.repo.UserExistsByEmail(ctx, email); err != nil {
		return nil, fmt.Errorf("admin: check user email: %w", err)
	} else if exists {
		return nil, ErrEmailInUse
	}
	if pending, err := s.repo.PendingInviteExists(ctx, email); err != nil {
		return nil, fmt.Errorf("admin: check pending invite: %w", err)
	} else if pending {
		return nil, ErrInvitePending
	}

	token, hash, err := newInviteToken()
	if err != nil {
		return nil, err
	}
	invite, err := s.repo.CreateInvite(ctx, queries.CreateUserInviteParams{
		Email:     email,
		Name:      name,
		Role:      input.Role,
		TokenHash: hash,
		InvitedBy: actor,
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(inviteTTL), Valid: true},
	}, AuditEntry{ActorID: actor, Action: "invite_sent", Data: map[string]any{"email": email, "role": input.Role}})
	if err != nil {
		return nil, fmt.Errorf("admin: create invite: %w", err)
	}
	return s.inviteLink(ctx, invite, token, actor)
}

func (s *Service) ListInvites(ctx context.Context) (*InvitesResult, error) {
	rows, err := s.repo.ListPendingInvites(ctx)
	if err != nil {
		return nil, fmt.Errorf("admin: list invites: %w", err)
	}
	data := make([]InviteItem, 0, len(rows))
	for _, row := range rows {
		id, err := uuidFromPg(row.ID)
		if err != nil {
			continue
		}
		data = append(data, InviteItem{
			ID:        id.String(),
			Email:     row.Email,
			Name:      row.Name,
			Role:      row.Role,
			InvitedBy: row.InvitedByName,
			ExpiresAt: timestamptzValue(row.ExpiresAt),
			CreatedAt: timestamptzValue(row.CreatedAt),
			Expired:   row.ExpiresAt.Valid && time.Now().After(row.ExpiresAt.Time),
		})
	}
	return &InvitesResult{Data: data}, nil
}

// ResendInvite issues a new link (the old one stops working) and restarts the 7 days.
func (s *Service) ResendInvite(ctx context.Context, inviteID, actorID string) (*InviteLinkResult, error) {
	id, actor, err := s.targetAndActor(inviteID, actorID)
	if err != nil {
		return nil, err
	}
	existing, err := s.repo.GetInviteByID(ctx, id)
	if err != nil {
		return nil, err
	}
	token, hash, err := newInviteToken()
	if err != nil {
		return nil, err
	}
	invite, err := s.repo.RotateInvite(ctx, queries.RotateUserInviteTokenParams{ID: id, TokenHash: hash, ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(inviteTTL), Valid: true}},
		AuditEntry{ActorID: actor, Action: "invite_resent", Data: map[string]any{"email": existing.Email}})
	if err != nil {
		return nil, err
	}
	return s.inviteLink(ctx, invite, token, actor)
}

func (s *Service) RevokeInvite(ctx context.Context, inviteID, actorID string) error {
	id, actor, err := s.targetAndActor(inviteID, actorID)
	if err != nil {
		return err
	}
	existing, err := s.repo.GetInviteByID(ctx, id)
	if err != nil {
		return err
	}
	_, err = s.repo.RevokeInvite(ctx, id, AuditEntry{ActorID: actor, Action: "invite_revoked", Data: map[string]any{"email": existing.Email}})
	return err
}

// PublicInvite returns what the accept page shows. No auth: the token is the credential.
func (s *Service) PublicInvite(ctx context.Context, token string) (*PublicInviteResult, error) {
	invite, err := s.usableInvite(ctx, token)
	if err != nil {
		return nil, err
	}
	result := &PublicInviteResult{}
	result.Data.Email = invite.Email
	result.Data.Name = invite.Name
	result.Data.Role = invite.Role
	result.Data.ExpiresAt = timestamptzValue(invite.ExpiresAt)
	return result, nil
}

// AcceptInvite creates the account (email already confirmed: the link proves it).
func (s *Service) AcceptInvite(ctx context.Context, token, name, password string) (*AcceptInviteResult, error) {
	if len(password) < 8 || strings.TrimSpace(name) == "" {
		return nil, ErrInviteInput
	}
	invite, err := s.usableInvite(ctx, token)
	if err != nil {
		return nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("admin: hash password: %w", err)
	}
	if _, err := s.repo.AcceptInvite(ctx, invite.ID, queries.AdminCreateInvitedUserParams{
		Email:        invite.Email,
		PasswordHash: string(hash),
		Role:         invite.Role,
		Name:         strings.TrimSpace(name),
	}); err != nil {
		return nil, err
	}
	result := &AcceptInviteResult{}
	result.Data.Email = invite.Email
	return result, nil
}

func (s *Service) usableInvite(ctx context.Context, token string) (queries.UserInvite, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return queries.UserInvite{}, ErrInviteInvalid
	}
	invite, err := s.repo.GetInviteByTokenHash(ctx, hashInviteToken(token))
	if errors.Is(err, ErrNotFound) || (err == nil && (invite.AcceptedAt.Valid || invite.RevokedAt.Valid)) {
		return queries.UserInvite{}, ErrInviteInvalid
	}
	if err != nil {
		return queries.UserInvite{}, fmt.Errorf("admin: get invite: %w", err)
	}
	if !invite.ExpiresAt.Valid || time.Now().After(invite.ExpiresAt.Time) {
		return queries.UserInvite{}, ErrInviteExpired
	}
	return invite, nil
}

func (s *Service) inviteLink(ctx context.Context, invite queries.UserInvite, token string, actor pgtype.UUID) (*InviteLinkResult, error) {
	id, err := uuidFromPg(invite.ID)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s/convite/%s", s.dashboardURL, token)
	inviter := "A equipa"
	if row, err := s.repo.GetUser(ctx, actor); err == nil {
		inviter = row.Name
	}
	emailSent := false
	if s.emailSender != nil {
		// Best effort: the invite exists either way and the admin gets the link to share.
		emailSent = s.emailSender.SendInviteEmail(ctx, appemail.InviteMessage{
			To:            invite.Email,
			RecipientName: invite.Name,
			InviterName:   inviter,
			RoleLabel:     inviteRoleLabels[invite.Role],
			InviteURL:     url,
		}) == nil
	}
	result := &InviteLinkResult{}
	result.Data.Invite = InviteItem{
		ID:        id.String(),
		Email:     invite.Email,
		Name:      invite.Name,
		Role:      invite.Role,
		InvitedBy: inviter,
		ExpiresAt: timestamptzValue(invite.ExpiresAt),
		CreatedAt: timestamptzValue(invite.CreatedAt),
	}
	result.Data.InviteURL = url
	result.Data.EmailSent = emailSent
	return result, nil
}

func newInviteToken() (token string, hash string, err error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", "", fmt.Errorf("admin: invite token: %w", err)
	}
	token = base64.RawURLEncoding.EncodeToString(raw)
	return token, hashInviteToken(token), nil
}

func hashInviteToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func contains(values []string, value string) bool {
	for _, item := range values {
		if item == value {
			return true
		}
	}
	return false
}
