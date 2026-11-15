-- name: GetApplication :one
SELECT id, event_id, organization_id, title, status, note
FROM applications
WHERE id = $1;

-- name: ListApplicationsByEvent :many
SELECT a.id, a.title, a.status, o.name AS organization_name
FROM applications AS a
JOIN organizations AS o ON o.id = a.organization_id
WHERE a.event_id = $1
ORDER BY a.id;

-- name: SubmitApplication :execrows
UPDATE applications
SET status = 'submitted'
WHERE id = $1 AND status = 'draft';
