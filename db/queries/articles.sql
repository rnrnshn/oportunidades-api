-- name: GetArticleBySlug :one
SELECT *
FROM articles
WHERE slug = $1
  AND status = 'published'
  AND deleted_at IS NULL;

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
