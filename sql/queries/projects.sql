-- name: CreateProject :one
INSERT INTO projects (
  workspace_id,
  name,
  description
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: ListProjectsForWorkspaceMember :many
SELECT p.*
FROM projects p
INNER JOIN workspace_members wm ON wm.workspace_id = p.workspace_id
WHERE p.workspace_id = $1 AND wm.user_id = $2
ORDER BY p.created_at DESC
LIMIT $3 OFFSET $4;

-- name: GetProjectForWorkspaceMember :one
SELECT p.*
FROM projects p
INNER JOIN workspace_members wm ON wm.workspace_id = p.workspace_id
WHERE p.id = $1 AND wm.user_id = $2
LIMIT 1;
