-- repository_events

-- name: CreateRepositoryEvent :one
INSERT INTO repository_events (
  repository_id,
  event_type,
  external_id,
  payload,
  occurred_at
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetRepositoryEventByID :one
SELECT * FROM repository_events
WHERE id = $1
LIMIT 1;

-- name: GetRepositoryEventForMember :one
SELECT e.*
FROM repository_events e
INNER JOIN repositories r ON r.id = e.repository_id
INNER JOIN projects p ON p.id = r.project_id
INNER JOIN workspace_members wm ON wm.workspace_id = p.workspace_id
WHERE e.id = $1 AND wm.user_id = $2
LIMIT 1;

-- name: ListRepositoryEventsForMember :many
SELECT e.*
FROM repository_events e
INNER JOIN repositories r ON r.id = e.repository_id
INNER JOIN projects p ON p.id = r.project_id
INNER JOIN workspace_members wm ON wm.workspace_id = p.workspace_id
WHERE e.repository_id = $1 AND wm.user_id = $2
ORDER BY e.occurred_at DESC
LIMIT $3 OFFSET $4;

-- pull_requests

-- name: CreatePullRequest :one
INSERT INTO pull_requests (
  repository_id,
  number,
  title,
  state,
  author,
  base_branch,
  head_branch,
  changed_files,
  additions,
  deletions,
  risk_level
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
RETURNING *;

-- name: GetPullRequestByID :one
SELECT * FROM pull_requests
WHERE id = $1
LIMIT 1;

-- name: GetPullRequestForMember :one
SELECT pr.*
FROM pull_requests pr
INNER JOIN repositories r ON r.id = pr.repository_id
INNER JOIN projects p ON p.id = r.project_id
INNER JOIN workspace_members wm ON wm.workspace_id = p.workspace_id
WHERE pr.id = $1 AND wm.user_id = $2
LIMIT 1;

-- name: ListPullRequestsForMember :many
SELECT pr.*
FROM pull_requests pr
INNER JOIN repositories r ON r.id = pr.repository_id
INNER JOIN projects p ON p.id = r.project_id
INNER JOIN workspace_members wm ON wm.workspace_id = p.workspace_id
WHERE pr.repository_id = $1 AND wm.user_id = $2
ORDER BY pr.number DESC
LIMIT $3 OFFSET $4;

-- issues

-- name: CreateIssue :one
INSERT INTO issues (
  repository_id,
  number,
  title,
  state,
  author,
  labels,
  priority,
  category
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: GetIssueByID :one
SELECT * FROM issues
WHERE id = $1
LIMIT 1;

-- name: GetIssueForMember :one
SELECT i.*
FROM issues i
INNER JOIN repositories r ON r.id = i.repository_id
INNER JOIN projects p ON p.id = r.project_id
INNER JOIN workspace_members wm ON wm.workspace_id = p.workspace_id
WHERE i.id = $1 AND wm.user_id = $2
LIMIT 1;

-- name: ListIssuesForMember :many
SELECT i.*
FROM issues i
INNER JOIN repositories r ON r.id = i.repository_id
INNER JOIN projects p ON p.id = r.project_id
INNER JOIN workspace_members wm ON wm.workspace_id = p.workspace_id
WHERE i.repository_id = $1 AND wm.user_id = $2
ORDER BY i.number DESC
LIMIT $3 OFFSET $4;

-- commits

-- name: CreateCommit :one
INSERT INTO commits (
  repository_id,
  sha,
  message,
  author,
  committed_at
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetCommitByID :one
SELECT * FROM commits
WHERE id = $1
LIMIT 1;

-- name: GetCommitForMember :one
SELECT c.*
FROM commits c
INNER JOIN repositories r ON r.id = c.repository_id
INNER JOIN projects p ON p.id = r.project_id
INNER JOIN workspace_members wm ON wm.workspace_id = p.workspace_id
WHERE c.id = $1 AND wm.user_id = $2
LIMIT 1;

-- name: ListCommitsForMember :many
SELECT c.*
FROM commits c
INNER JOIN repositories r ON r.id = c.repository_id
INNER JOIN projects p ON p.id = r.project_id
INNER JOIN workspace_members wm ON wm.workspace_id = p.workspace_id
WHERE c.repository_id = $1 AND wm.user_id = $2
ORDER BY c.committed_at DESC
LIMIT $3 OFFSET $4;
