-- name: CreateDocument :exec
INSERT INTO document (
  id, name, description, created_at)
VALUES (
  $1, $2, $3, $4
);

-- name: GetDocument :one
SELECT id, name, description, created_at FROM document
WHERE id = $1;

-- name: ListDocuments :many
SELECT id, name, description, created_at
FROM document;

-- name: DeleteDocument :exec
DELETE FROM document
WHERE id = $1;

-- name: CreateDocumentChunk :exec
INSERT INTO document_chunk (
  id, document_id, fragment, embedding, created_at)
VALUES (
  $1, $2, $3, $4, $5
);

-- name: FindClosestChunks :many
SELECT id, document_id, fragment, embedding, created_at
FROM document_chunk
ORDER BY embedding <-> $1 LIMIT $2;

-- name: DeleteDocumentChunk :exec
DELETE FROM document_chunk
WHERE id = $1;

-- name: DeleteDocumentChunkForDocument :exec
DELETE FROM document_chunk
WHERE document_id = $1;

-- name: ListDocumentChunksForDocument :many
SELECT id, fragment, created_at, embedding
FROM document_chunk
WHERE document_id = $1;
