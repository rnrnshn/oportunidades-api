-- Team = people with panel access; community = everyone else.
-- last_seen_at: newest session (every login/refresh creates one).
-- name: AdminListUsers :many
SELECT
  u.id, u.email, u.role, u.name, u.avatar_url, u.email_verified_at, u.created_at, u.suspended_at,
  (SELECT MAX(rt.created_at) FROM refresh_tokens rt WHERE rt.user_id = u.id)::timestamptz AS last_seen_at,
  (
    (SELECT COUNT(*) FROM articles a WHERE a.author_id = u.id AND a.deleted_at IS NULL)
    + (SELECT COUNT(*) FROM opportunities o WHERE o.published_by = u.id AND o.deleted_at IS NULL)
    + (SELECT COUNT(*) FROM universities un WHERE un.created_by = u.id AND un.deleted_at IS NULL)
  )::bigint AS content_count
FROM users u
WHERE u.deleted_at IS NULL
  AND (
    sqlc.narg('scope')::text IS NULL
    OR (sqlc.narg('scope')::text = 'team' AND u.role IN ('admin', 'cms_partner'))
    OR (sqlc.narg('scope')::text = 'community' AND u.role IN ('user', 'mentor'))
  )
  AND (sqlc.narg('role')::text IS NULL OR u.role = sqlc.narg('role')::text)
  AND (
    sqlc.narg('status')::text IS NULL
    OR (sqlc.narg('status')::text = 'active' AND u.suspended_at IS NULL)
    OR (sqlc.narg('status')::text = 'suspended' AND u.suspended_at IS NOT NULL)
    OR (sqlc.narg('status')::text = 'unverified' AND u.email_verified_at IS NULL)
  )
  AND (
    sqlc.narg('query')::text IS NULL
    OR u.name ILIKE '%' || sqlc.narg('query')::text || '%'
    OR u.email ILIKE '%' || sqlc.narg('query')::text || '%'
  )
ORDER BY
  CASE WHEN sqlc.arg('sort')::text = 'name_asc' THEN u.name END ASC,
  CASE WHEN sqlc.arg('sort')::text = 'name_desc' THEN u.name END DESC,
  CASE WHEN sqlc.arg('sort')::text = 'created_at_asc' THEN u.created_at END ASC,
  CASE WHEN sqlc.arg('sort')::text = 'last_seen_desc' THEN (SELECT MAX(rt.created_at) FROM refresh_tokens rt WHERE rt.user_id = u.id) END DESC NULLS LAST,
  u.created_at DESC,
  u.id
LIMIT sqlc.arg('limit')::int OFFSET sqlc.arg('offset')::int;

-- name: AdminCountUsers :one
SELECT COUNT(*)
FROM users u
WHERE u.deleted_at IS NULL
  AND (
    sqlc.narg('scope')::text IS NULL
    OR (sqlc.narg('scope')::text = 'team' AND u.role IN ('admin', 'cms_partner'))
    OR (sqlc.narg('scope')::text = 'community' AND u.role IN ('user', 'mentor'))
  )
  AND (sqlc.narg('role')::text IS NULL OR u.role = sqlc.narg('role')::text)
  AND (
    sqlc.narg('status')::text IS NULL
    OR (sqlc.narg('status')::text = 'active' AND u.suspended_at IS NULL)
    OR (sqlc.narg('status')::text = 'suspended' AND u.suspended_at IS NOT NULL)
    OR (sqlc.narg('status')::text = 'unverified' AND u.email_verified_at IS NULL)
  )
  AND (
    sqlc.narg('query')::text IS NULL
    OR u.name ILIKE '%' || sqlc.narg('query')::text || '%'
    OR u.email ILIKE '%' || sqlc.narg('query')::text || '%'
  );

-- name: AdminUserTabCounts :one
SELECT
  (SELECT COUNT(*) FROM users WHERE deleted_at IS NULL AND role IN ('admin', 'cms_partner'))::bigint AS team,
  (SELECT COUNT(*) FROM users WHERE deleted_at IS NULL AND role IN ('user', 'mentor'))::bigint AS community,
  (SELECT COUNT(*) FROM user_invites WHERE accepted_at IS NULL AND revoked_at IS NULL AND expires_at > NOW())::bigint AS invites;

-- name: AdminCountUsersByRole :many
SELECT role, COUNT(*)::bigint AS total
FROM users
WHERE deleted_at IS NULL
GROUP BY role;

-- name: AdminGetUser :one
SELECT
  u.id, u.email, u.role, u.name, u.avatar_url, u.email_verified_at, u.created_at,
  u.suspended_at, u.suspension_reason,
  sb.name AS suspended_by_name,
  (SELECT MAX(rt.created_at) FROM refresh_tokens rt WHERE rt.user_id = u.id)::timestamptz AS last_seen_at,
  (SELECT COUNT(*) FROM articles a WHERE a.author_id = u.id AND a.deleted_at IS NULL)::bigint AS article_count,
  (SELECT COUNT(*) FROM opportunities o WHERE o.published_by = u.id AND o.deleted_at IS NULL)::bigint AS opportunity_count,
  (SELECT COUNT(*) FROM universities un WHERE un.created_by = u.id AND un.deleted_at IS NULL)::bigint AS university_count
FROM users u
LEFT JOIN users sb ON sb.id = u.suspended_by
WHERE u.id = $1
  AND u.deleted_at IS NULL;

-- Demoting an admin only succeeds while another active admin remains, checked in the same statement.
-- name: AdminUpdateUserRole :one
UPDATE users AS u
SET role = sqlc.arg('role')::text
WHERE u.id = sqlc.arg('id')
  AND u.deleted_at IS NULL
  AND (
    sqlc.arg('role')::text = 'admin'
    OR u.role <> 'admin'
    OR (SELECT COUNT(*) FROM users AS a WHERE a.role = 'admin' AND a.deleted_at IS NULL AND a.suspended_at IS NULL AND a.id <> u.id) > 0
  )
RETURNING u.id, u.email, u.role, u.name, u.avatar_url, u.email_verified_at, u.created_at;

-- Suspending an admin only succeeds while another active admin remains.
-- name: AdminSuspendUser :one
UPDATE users AS u
SET suspended_at = NOW(), suspended_by = sqlc.arg('actor_id'), suspension_reason = sqlc.arg('reason')::text
WHERE u.id = sqlc.arg('id')
  AND u.deleted_at IS NULL
  AND u.suspended_at IS NULL
  AND (
    u.role <> 'admin'
    OR (SELECT COUNT(*) FROM users AS a WHERE a.role = 'admin' AND a.deleted_at IS NULL AND a.suspended_at IS NULL AND a.id <> u.id) > 0
  )
RETURNING u.id;

-- name: AdminUnsuspendUser :one
UPDATE users
SET suspended_at = NULL, suspended_by = NULL, suspension_reason = NULL
WHERE id = $1
  AND deleted_at IS NULL
  AND suspended_at IS NOT NULL
RETURNING id;

-- name: AdminUserExistsByEmail :one
SELECT EXISTS(SELECT 1 FROM users WHERE lower(email) = lower($1) AND deleted_at IS NULL);

-- name: AdminCreateInvitedUser :one
INSERT INTO users (email, password_hash, role, name, email_verified_at)
VALUES ($1, $2, $3, $4, NOW())
RETURNING id;
