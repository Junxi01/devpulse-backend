-- name: CreateProject :one
INSERT INTO projects (
  owner_id,
  name,
  description,
  github_owner,
  github_repo
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetProjectByID :one
SELECT * FROM projects
WHERE id = $1
LIMIT 1;

-- name: ListProjectsByOwner :many
SELECT * FROM projects
WHERE owner_id = $1
ORDER BY created_at DESC
LIMIT $2
OFFSET $3;

-- name: UpdateProject :one
UPDATE projects
SET
  name = $2,
  description = $3,
  github_owner = $4,
  github_repo = $5,
  updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteProject :exec
DELETE FROM projects
WHERE id = $1;

