-- name: CreateUserInvite :one
INSERT INTO user_invites (email, name, role, token_hash, invited_by, expires_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListPendingUserInvites :many
SELECT i.id, i.email, i.name, i.role, i.expires_at, i.created_at, inviter.name AS invited_by_name
FROM user_invites i
JOIN users inviter ON inviter.id = i.invited_by
WHERE i.accepted_at IS NULL
  AND i.revoked_at IS NULL
ORDER BY i.created_at DESC;

-- name: GetUserInviteByID :one
SELECT * FROM user_invites WHERE id = $1;

-- name: GetUserInviteByTokenHash :one
SELECT * FROM user_invites WHERE token_hash = $1;

-- name: PendingUserInviteExists :one
SELECT EXISTS(
  SELECT 1 FROM user_invites
  WHERE lower(email) = lower($1) AND accepted_at IS NULL AND revoked_at IS NULL AND expires_at > NOW()
);

-- name: RotateUserInviteToken :one
UPDATE user_invites
SET token_hash = $2, expires_at = $3
WHERE id = $1 AND accepted_at IS NULL AND revoked_at IS NULL
RETURNING *;

-- name: RevokeUserInvite :one
UPDATE user_invites
SET revoked_at = NOW()
WHERE id = $1 AND accepted_at IS NULL AND revoked_at IS NULL
RETURNING *;

-- Single use: only succeeds once, and only while the invite is still valid.
-- name: AcceptUserInvite :one
UPDATE user_invites
SET accepted_at = NOW()
WHERE id = $1 AND accepted_at IS NULL AND revoked_at IS NULL AND expires_at > NOW()
RETURNING *;
