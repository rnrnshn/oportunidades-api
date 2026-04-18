-- name: UpdateUserProfile :one
UPDATE users
SET
  name = $2,
  avatar_url = $3
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;
