-- name: GetOpportunityBySlug :one
SELECT *
FROM opportunities
WHERE slug = $1
  AND deleted_at IS NULL;

-- name: ListOpportunities :many
SELECT *
FROM opportunities
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountOpportunities :one
SELECT COUNT(*)
FROM opportunities
WHERE deleted_at IS NULL;
