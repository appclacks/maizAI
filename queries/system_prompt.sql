-- name: CreateSystemPrompt :exec
INSERT INTO system_prompt (
  id, name, description, content, created_at
) VALUES (
  $1, $2, $3, $4, $5
);

-- name: GetSystemPrompt :one
SELECT id, name, description, content, created_at
FROM system_prompt
WHERE id = $1;

-- name: GetSystemPromptByName :one
SELECT id, name, description, content, created_at
FROM system_prompt
WHERE name = $1;

-- name: ListSystemPrompts :many
SELECT id, name, description, content, created_at
FROM system_prompt;

-- name: UpdateSystemPrompt :exec
UPDATE system_prompt
SET content = $2
WHERE id = $1;

-- name: DeleteSystemPrompt :exec
DELETE FROM system_prompt
WHERE id = $1;
