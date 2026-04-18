-- name: GetOpportunityBySlug :one
SELECT *
FROM opportunities
WHERE slug = $1
  AND deleted_at IS NULL;

-- name: GetOpportunityByID :one
SELECT *
FROM opportunities
WHERE id = $1
  AND deleted_at IS NULL;

-- name: CreateOpportunity :one
INSERT INTO opportunities (
  slug,
  title,
  type,
  entity_name,
  description,
  requirements,
  deadline,
  apply_url,
  country,
  language,
  area,
  is_active,
  published_by,
  verified
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5,
  $6,
  $7,
  $8,
  $9,
  $10,
  $11,
  $12,
  $13,
  $14
)
RETURNING *;

-- name: VerifyOpportunity :one
UPDATE opportunities
SET
  verified = TRUE,
  is_active = TRUE
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;

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
