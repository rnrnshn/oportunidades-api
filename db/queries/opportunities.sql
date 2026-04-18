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

-- name: ListCMSOpportunities :many
SELECT *
FROM opportunities
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountCMSOpportunities :one
SELECT COUNT(*)
FROM opportunities
WHERE deleted_at IS NULL;

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

-- name: RejectOpportunity :one
UPDATE opportunities
SET
  verified = FALSE,
  is_active = FALSE
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;

-- name: DeactivateOpportunity :one
UPDATE opportunities
SET
  is_active = FALSE
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;

-- name: UpdateOpportunity :one
UPDATE opportunities
SET
  title = $2,
  type = $3,
  entity_name = $4,
  description = $5,
  requirements = $6,
  deadline = $7,
  apply_url = $8,
  country = $9,
  language = $10,
  area = $11
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
