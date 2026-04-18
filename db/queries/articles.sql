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

-- name: CreateArticle :one
INSERT INTO articles (
  slug,
  title,
  excerpt,
  content,
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

-- name: UpdateArticle :one
UPDATE articles
SET
  title = $2,
  excerpt = $3,
  content = $4,
  cover_image_url = $5,
  type = $6,
  source_name = $7,
  source_url = $8,
  seo_title = $9,
  seo_description = $10,
  is_featured = $11
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
