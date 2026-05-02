-- name: CreateWorkspace :one
INSERT INTO workspaces (
  name,
  owner_id
) VALUES (
  $1, $2
)
RETURNING *;

-- name: AddWorkspaceMember :exec
INSERT INTO workspace_members (
  workspace_id,
  user_id,
  role
) VALUES (
  $1, $2, $3
)
ON CONFLICT (workspace_id, user_id) DO NOTHING;

-- name: ListWorkspacesByUser :many
SELECT w.*
FROM workspaces w
JOIN workspace_members wm ON wm.workspace_id = w.id
WHERE wm.user_id = $1
ORDER BY w.created_at DESC
LIMIT $2
OFFSET $3;

-- name: GetWorkspaceByID :one
SELECT w.*
FROM workspaces w
JOIN workspace_members wm ON wm.workspace_id = w.id
WHERE w.id = $1 AND wm.user_id = $2
LIMIT 1;

