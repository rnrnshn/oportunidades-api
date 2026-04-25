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
  city,
  country,
  description,
  logo_url,
  campus_image_url,
  website,
  email,
  phone,
  founded_year,
  address,
  map_url,
  academic_calendar,
  student_count,
  admissions_deadline,
  tags,
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
  $22
)
RETURNING *;

-- name: UpdateUniversity :one
UPDATE universities
SET
  name = $2,
  type = $3,
  province = $4,
  city = $5,
  country = $6,
  description = $7,
  logo_url = $8,
  campus_image_url = $9,
  website = $10,
  email = $11,
  phone = $12,
  founded_year = $13,
  address = $14,
  map_url = $15,
  academic_calendar = $16,
  student_count = $17,
  admissions_deadline = $18,
  tags = $19
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;

-- name: ListCoursesByUniversityID :many
SELECT *
FROM courses
WHERE university_id = $1
  AND deleted_at IS NULL
ORDER BY level ASC, name ASC;

-- name: ListUniversityFeesByUniversityID :many
SELECT *
FROM university_fees
WHERE university_id = $1
  AND deleted_at IS NULL
ORDER BY sort_order ASC, created_at ASC;

-- name: SoftDeleteUniversityFeesByUniversityID :exec
UPDATE university_fees
SET deleted_at = NOW()
WHERE university_id = $1
  AND deleted_at IS NULL;

-- name: CreateUniversityFee :one
INSERT INTO university_fees (
  university_id,
  label,
  value,
  sort_order
) VALUES (
  $1,
  $2,
  $3,
  $4
)
RETURNING *;

-- name: ListUniversityScholarshipsByUniversityID :many
SELECT *
FROM university_scholarships
WHERE university_id = $1
  AND deleted_at IS NULL
ORDER BY sort_order ASC, created_at ASC;

-- name: SoftDeleteUniversityScholarshipsByUniversityID :exec
UPDATE university_scholarships
SET deleted_at = NOW()
WHERE university_id = $1
  AND deleted_at IS NULL;

-- name: CreateUniversityScholarship :one
INSERT INTO university_scholarships (
  university_id,
  name,
  amount,
  status,
  sort_order
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5
)
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
