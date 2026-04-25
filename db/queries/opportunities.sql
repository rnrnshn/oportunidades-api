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
  external_url_label,
  country,
  location,
  is_remote,
  language,
  area,
  hero_image_url,
  provider_logo_url,
  amount_min,
  amount_max,
  amount_currency,
  coverage,
  eligibility,
  application_process,
  degree_level,
  program_area,
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
  $14,
  $15,
  $16,
  $17,
  $18,
  $19,
  $20,
  $21,
  $22,
  $23,
  $24,
  $25,
  $26,
  $27
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
  external_url_label = $9,
  country = $10,
  location = $11,
  is_remote = $12,
  language = $13,
  area = $14,
  hero_image_url = $15,
  provider_logo_url = $16,
  amount_min = $17,
  amount_max = $18,
  amount_currency = $19,
  coverage = $20,
  eligibility = $21,
  application_process = $22,
  degree_level = $23,
  program_area = $24
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
