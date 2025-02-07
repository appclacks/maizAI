-- name: GetContext :one
SELECT id, name, description, created_at FROM context
WHERE id = $1;

-- name: GetContextIDByName :one
SELECT id FROM context
WHERE name = $1;

-- name: GetContextNameByID :one
SELECT name FROM context
WHERE id = $1;

-- name: ListContexts :many
SELECT id, name, description, created_at FROM context;

-- name: CreateContext :one
INSERT INTO context (
  id, name, description, created_at
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: DeleteContext :exec
DELETE FROM context
WHERE id = $1;
