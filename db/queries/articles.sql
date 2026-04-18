-- name: GetArticleBySlug :one
SELECT *
FROM articles
WHERE slug = $1
  AND status = 'published'
  AND deleted_at IS NULL;

-- name: GetArticleByID :one
SELECT *
FROM articles
WHERE id = $1
  AND deleted_at IS NULL;

-- name: ListCMSArticles :many
SELECT *
FROM articles
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountCMSArticles :one
SELECT COUNT(*)
FROM articles
WHERE deleted_at IS NULL;

-- name: CreateArticle :one
INSERT INTO articles (
  slug,
  title,
  excerpt,
  content,
  content_json,
  cover_image_url,
  type,
  status,
  source_name,
  source_url,
  seo_title,
  seo_description,
  is_featured,
  author_id,
  published_at
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
  ,$15
)
RETURNING *;

-- name: PublishArticle :one
UPDATE articles
SET
  status = 'published',
  published_at = NOW()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;

-- name: UnpublishArticle :one
UPDATE articles
SET
  status = 'draft',
  published_at = NULL
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;

-- name: ArchiveArticle :one
UPDATE articles
SET
  status = 'archived'
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;

-- name: UpdateArticle :one
UPDATE articles
SET
  title = $2,
  excerpt = $3,
  content = $4,
  content_json = $5,
  cover_image_url = $6,
  type = $7,
  source_name = $8,
  source_url = $9,
  seo_title = $10,
  seo_description = $11,
  is_featured = $12
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;

-- name: ListArticles :many
SELECT *
FROM articles
WHERE status = 'published'
  AND deleted_at IS NULL
ORDER BY published_at DESC, created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountArticles :one
SELECT COUNT(*)
FROM articles
WHERE status = 'published'
  AND deleted_at IS NULL;
