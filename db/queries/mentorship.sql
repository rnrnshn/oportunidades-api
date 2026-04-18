-- name: ListMentors :many
SELECT
  mp.id,
  mp.user_id,
  mp.headline,
  mp.bio,
  mp.expertise,
  mp.availability,
  mp.is_active,
  mp.created_at,
  mp.updated_at,
  mp.deleted_at,
  u.id AS user_id,
  u.email,
  u.password_hash,
  u.role,
  u.name,
  u.avatar_url,
  u.created_at,
  u.updated_at,
  u.deleted_at
FROM mentor_profiles mp
JOIN users u ON u.id = mp.user_id
WHERE mp.deleted_at IS NULL
  AND u.deleted_at IS NULL
  AND u.role = 'mentor'
  AND mp.is_active = TRUE
ORDER BY u.name ASC
LIMIT $1 OFFSET $2;

-- name: CountMentors :one
SELECT COUNT(*)
FROM mentor_profiles mp
JOIN users u ON u.id = mp.user_id
WHERE mp.deleted_at IS NULL
  AND u.deleted_at IS NULL
  AND u.role = 'mentor'
  AND mp.is_active = TRUE;

-- name: GetMentorByID :one
SELECT
  mp.id,
  mp.user_id,
  mp.headline,
  mp.bio,
  mp.expertise,
  mp.availability,
  mp.is_active,
  mp.created_at,
  mp.updated_at,
  mp.deleted_at,
  u.id AS user_id,
  u.email,
  u.password_hash,
  u.role,
  u.name,
  u.avatar_url,
  u.created_at,
  u.updated_at,
  u.deleted_at
FROM mentor_profiles mp
JOIN users u ON u.id = mp.user_id
WHERE mp.user_id = $1
  AND mp.deleted_at IS NULL
  AND u.deleted_at IS NULL
  AND u.role = 'mentor'
  AND mp.is_active = TRUE;

-- name: CreateMentorshipSession :one
INSERT INTO mentorship_sessions (
  mentor_id,
  requester_id,
  message,
  status,
  scheduled_at
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5
)
RETURNING *;

-- name: GetMentorshipSessionByID :one
SELECT *
FROM mentorship_sessions
WHERE id = $1
  AND deleted_at IS NULL;

-- name: ListMentorshipSessionsForUser :many
SELECT *
FROM mentorship_sessions
WHERE deleted_at IS NULL
  AND (mentor_id = $1 OR requester_id = $1)
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountMentorshipSessionsForUser :one
SELECT COUNT(*)
FROM mentorship_sessions
WHERE deleted_at IS NULL
  AND (mentor_id = $1 OR requester_id = $1);

-- name: UpdateMentorshipSessionStatus :one
UPDATE mentorship_sessions
SET
  status = $2,
  scheduled_at = $3
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;
