-- name: CreateUser :one
INSERT INTO users (
  email,
  password_hash,
  name
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1
LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1
LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY created_at DESC
LIMIT $1
OFFSET $2;

