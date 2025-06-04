-- name: GetContextMessages :many
SELECT id, type, tool_call_id, tool_call_name, tool_call_input, role, content, created_at FROM context_message
WHERE context_id = $1
ORDER BY ordering;

-- name: CreateContextMessage :one
INSERT INTO context_message (
  id, type, tool_call_id, tool_call_name, tool_call_input, role, content, created_at, context_id
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: DeleteContextMessagesForContext :exec
DELETE FROM context_message
WHERE context_id = $1;

-- name: UpdateContextMessage :exec
UPDATE context_message
SET content = $2, role=$3
WHERE id = $1;

-- name: DeleteContextMessage :exec
DELETE FROM context_message
WHERE id = $1;
