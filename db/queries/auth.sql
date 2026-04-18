-- name: CreateUser :one
INSERT INTO users (
  email,
  password_hash,
  role,
  name,
  avatar_url
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5
)
RETURNING *;

-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE email = $1
  AND deleted_at IS NULL;

-- name: GetUserByID :one
SELECT *
FROM users
WHERE id = $1
  AND deleted_at IS NULL;

-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (
  user_id,
  token_hash,
  expires_at
) VALUES (
  $1,
  $2,
  $3
)
RETURNING *;

-- name: GetRefreshTokenByHash :one
SELECT *
FROM refresh_tokens
WHERE token_hash = $1
  AND deleted_at IS NULL
  AND revoked_at IS NULL;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE id = $1
  AND deleted_at IS NULL;

-- name: RevokeAllRefreshTokensByUser :exec
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE user_id = $1
  AND deleted_at IS NULL
  AND revoked_at IS NULL;

-- name: CreateAuthActionToken :one
INSERT INTO auth_action_tokens (
  user_id,
  purpose,
  token_hash,
  expires_at
) VALUES (
  $1,
  $2,
  $3,
  $4
)
RETURNING *;

-- name: GetAuthActionTokenByHash :one
SELECT *
FROM auth_action_tokens
WHERE token_hash = $1
  AND purpose = $2
  AND deleted_at IS NULL
  AND consumed_at IS NULL;

-- name: ConsumeAuthActionToken :exec
UPDATE auth_action_tokens
SET consumed_at = NOW()
WHERE id = $1
  AND deleted_at IS NULL;

-- name: UpdateUserPasswordByID :one
UPDATE users
SET password_hash = $2
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;

-- name: MarkUserEmailVerified :one
UPDATE users
SET email_verified_at = NOW()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;

-- name: DeactivateUser :one
UPDATE users
SET deleted_at = NOW()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;
