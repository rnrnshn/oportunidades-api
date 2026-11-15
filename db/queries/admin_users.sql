-- name: AdminListUsers :many
SELECT id, email, role, name, avatar_url, email_verified_at, created_at
FROM users
WHERE deleted_at IS NULL
  AND (sqlc.narg('role')::text IS NULL OR role = sqlc.narg('role')::text)
  AND (
    sqlc.narg('query')::text IS NULL
    OR name ILIKE '%' || sqlc.narg('query')::text || '%'
    OR email ILIKE '%' || sqlc.narg('query')::text || '%'
  )
ORDER BY
  CASE WHEN sqlc.arg('sort')::text = 'name_asc' THEN name END ASC,
  CASE WHEN sqlc.arg('sort')::text = 'name_desc' THEN name END DESC,
  CASE WHEN sqlc.arg('sort')::text = 'created_at_asc' THEN created_at END ASC,
  created_at DESC,
  id
LIMIT sqlc.arg('limit')::int OFFSET sqlc.arg('offset')::int;

-- name: AdminCountUsers :one
SELECT COUNT(*)
FROM users
WHERE deleted_at IS NULL
  AND (sqlc.narg('role')::text IS NULL OR role = sqlc.narg('role')::text)
  AND (
    sqlc.narg('query')::text IS NULL
    OR name ILIKE '%' || sqlc.narg('query')::text || '%'
    OR email ILIKE '%' || sqlc.narg('query')::text || '%'
  );

-- name: AdminGetUser :one
SELECT id, email, role, name, avatar_url, email_verified_at, created_at
FROM users
WHERE id = $1
  AND deleted_at IS NULL;

-- name: AdminCountUsersByRole :many
SELECT role, COUNT(*)::bigint AS total
FROM users
WHERE deleted_at IS NULL
GROUP BY role;

-- Demoting an admin only succeeds while another admin remains, checked in the same statement.
-- name: AdminUpdateUserRole :one
UPDATE users AS u
SET role = sqlc.arg('role')::text
WHERE u.id = sqlc.arg('id')
  AND u.deleted_at IS NULL
  AND (
    sqlc.arg('role')::text = 'admin'
    OR u.role <> 'admin'
    OR (SELECT COUNT(*) FROM users AS a WHERE a.role = 'admin' AND a.deleted_at IS NULL) > 1
  )
RETURNING u.id, u.email, u.role, u.name, u.avatar_url, u.email_verified_at, u.created_at;
