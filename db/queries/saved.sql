-- name: IsPublishedOpportunity :one
SELECT EXISTS (
  SELECT 1
  FROM opportunities
  WHERE id = $1
    AND deleted_at IS NULL
    AND is_active
    AND verified
);

-- name: SaveOpportunity :exec
INSERT INTO saved_opportunities (user_id, opportunity_id)
VALUES ($1, $2)
ON CONFLICT (user_id, opportunity_id) DO NOTHING;

-- name: UnsaveOpportunity :exec
DELETE FROM saved_opportunities
WHERE user_id = $1
  AND opportunity_id = $2;

-- name: ListSavedOpportunities :many
SELECT sqlc.embed(o)
FROM saved_opportunities s
JOIN opportunities o ON o.id = s.opportunity_id
WHERE s.user_id = $1
  AND o.deleted_at IS NULL
  AND o.is_active
  AND o.verified
ORDER BY s.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountSavedOpportunities :one
SELECT COUNT(*)
FROM saved_opportunities s
JOIN opportunities o ON o.id = s.opportunity_id
WHERE s.user_id = $1
  AND o.deleted_at IS NULL
  AND o.is_active
  AND o.verified;

-- name: ListSavedOpportunityIDs :many
SELECT s.opportunity_id
FROM saved_opportunities s
JOIN opportunities o ON o.id = s.opportunity_id
WHERE s.user_id = $1
  AND o.deleted_at IS NULL
  AND o.is_active
  AND o.verified;
