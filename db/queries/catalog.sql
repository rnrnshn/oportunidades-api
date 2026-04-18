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

-- name: GetUniversityByID :one
SELECT *
FROM universities
WHERE id = $1
  AND deleted_at IS NULL;

-- name: ListCMSUniversities :many
SELECT *
FROM universities
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountCMSUniversities :one
SELECT COUNT(*)
FROM universities
WHERE deleted_at IS NULL;

-- name: CreateUniversity :one
INSERT INTO universities (
  slug,
  name,
  type,
  province,
  description,
  logo_url,
  website,
  email,
  phone,
  verified,
  verified_at,
  created_by
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
  $12
)
RETURNING *;

-- name: UpdateUniversity :one
UPDATE universities
SET
  name = $2,
  type = $3,
  province = $4,
  description = $5,
  logo_url = $6,
  website = $7,
  email = $8,
  phone = $9
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;

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

-- name: GetCourseByID :one
SELECT *
FROM courses
WHERE id = $1
  AND deleted_at IS NULL;

-- name: ListCMSCourses :many
SELECT *
FROM courses
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountCMSCourses :one
SELECT COUNT(*)
FROM courses
WHERE deleted_at IS NULL;

-- name: CreateCourse :one
INSERT INTO courses (
  slug,
  university_id,
  name,
  area,
  level,
  regime,
  duration_years,
  annual_fee,
  entry_requirements
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5,
  $6,
  $7,
  $8,
  $9
)
RETURNING *;

-- name: UpdateCourse :one
UPDATE courses
SET
  university_id = $2,
  name = $3,
  area = $4,
  level = $5,
  regime = $6,
  duration_years = $7,
  annual_fee = $8,
  entry_requirements = $9
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;
