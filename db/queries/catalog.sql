-- name: GetUniversityBySlug :one
SELECT *
FROM universities
WHERE slug = $1
  AND deleted_at IS NULL;

-- name: ListUniversities :many
SELECT *
FROM universities
WHERE deleted_at IS NULL
ORDER BY name ASC
LIMIT $1 OFFSET $2;

-- name: CountUniversities :one
SELECT COUNT(*)
FROM universities
WHERE deleted_at IS NULL;

-- name: GetCourseBySlug :one
SELECT *
FROM courses
WHERE slug = $1
  AND deleted_at IS NULL;

-- name: ListCourses :many
SELECT *
FROM courses
WHERE deleted_at IS NULL
ORDER BY name ASC
LIMIT $1 OFFSET $2;

-- name: CountCourses :one
SELECT COUNT(*)
FROM courses
WHERE deleted_at IS NULL;
