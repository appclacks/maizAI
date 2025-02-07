-- name: CreateContexSource :exec
INSERT INTO context_source (
  context_id, source_context_id
) VALUES (
  $1, $2
);

-- name: GetContextSourcesForContext :many
SELECT source_context_id FROM context_source
WHERE context_id=$1
ORDER BY ordering;

-- name: CleanContextSourcesForContext :exec
DELETE FROM context_source
WHERE context_id=$1 OR source_context_id=$1;

-- name: DeleteContextSource :exec
DELETE FROM context_source
WHERE context_id=$1 AND source_context_id=$2;
