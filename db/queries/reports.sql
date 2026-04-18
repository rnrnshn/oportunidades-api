-- name: CreateReport :one
INSERT INTO reports (
  reporter_id,
  entity_type,
  entity_id,
  reason,
  status,
  reviewed_by,
  moderation_notes,
  resolved_at
) VALUES (
  $1,
  $2,
  $3,
  $4,
  'pending',
  NULL,
  NULL,
  NULL
)
RETURNING *;

-- name: ReportUniversityExists :one
SELECT EXISTS(
  SELECT 1 FROM universities
  WHERE id = $1
    AND deleted_at IS NULL
);

-- name: ReportCourseExists :one
SELECT EXISTS(
  SELECT 1 FROM courses
  WHERE id = $1
    AND deleted_at IS NULL
);

-- name: ReportOpportunityExists :one
SELECT EXISTS(
  SELECT 1 FROM opportunities
  WHERE id = $1
    AND deleted_at IS NULL
);

-- name: GetReportByID :one
SELECT *
FROM reports
WHERE id = $1
  AND deleted_at IS NULL;

-- name: ListReports :many
SELECT *
FROM reports
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountReports :one
SELECT COUNT(*)
FROM reports
WHERE deleted_at IS NULL;

-- name: UpdateReportStatus :one
UPDATE reports
SET
  status = $2,
  reviewed_by = $3,
  moderation_notes = $4,
  resolved_at = CASE WHEN $2 IN ('resolved', 'dismissed') THEN NOW() ELSE NULL END
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;
