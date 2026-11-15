-- name: CreateAuditEvent :exec
INSERT INTO audit_events (actor_id, action, target_user_id, data)
VALUES ($1, $2, $3, $4);

-- name: ListAuditEventsForUser :many
SELECT e.id, e.action, e.data, e.created_at, e.actor_id, actor.name AS actor_name
FROM audit_events e
LEFT JOIN users actor ON actor.id = e.actor_id
WHERE e.target_user_id = $1
ORDER BY e.created_at DESC
LIMIT 20;
