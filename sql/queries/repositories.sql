-- name: CreateRepository :one
INSERT INTO repositories (
  project_id,
  provider,
  owner,
  name,
  full_name,
  external_id,
  default_branch
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: ListRepositoriesForWorkspaceMember :many
SELECT r.*
FROM repositories r
INNER JOIN projects p ON p.id = r.project_id
INNER JOIN workspace_members wm ON wm.workspace_id = p.workspace_id
WHERE r.project_id = $1 AND wm.user_id = $2
ORDER BY r.created_at DESC
LIMIT $3 OFFSET $4;

-- name: GetRepositoryForWorkspaceMember :one
SELECT r.*
FROM repositories r
INNER JOIN projects p ON p.id = r.project_id
INNER JOIN workspace_members wm ON wm.workspace_id = p.workspace_id
WHERE r.id = $1 AND wm.user_id = $2
LIMIT 1;

-- name: GetRepositoryByProviderFullName :one
SELECT * FROM repositories
WHERE provider = $1 AND full_name = $2
LIMIT 1;
