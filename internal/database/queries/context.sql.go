// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: context.sql

package queries

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createContext = `-- name: CreateContext :one
INSERT INTO context (
  id, name, description, created_at
) VALUES (
  $1, $2, $3, $4
)
RETURNING id, name, description, created_at
`

type CreateContextParams struct {
	ID          pgtype.UUID
	Name        string
	Description pgtype.Text
	CreatedAt   pgtype.Timestamp
}

func (q *Queries) CreateContext(ctx context.Context, arg CreateContextParams) (Context, error) {
	row := q.db.QueryRow(ctx, createContext,
		arg.ID,
		arg.Name,
		arg.Description,
		arg.CreatedAt,
	)
	var i Context
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.CreatedAt,
	)
	return i, err
}

const deleteContext = `-- name: DeleteContext :exec
DELETE FROM context
WHERE id = $1
`

func (q *Queries) DeleteContext(ctx context.Context, id pgtype.UUID) error {
	_, err := q.db.Exec(ctx, deleteContext, id)
	return err
}

const getContext = `-- name: GetContext :one
SELECT id, name, description, created_at FROM context
WHERE id = $1
`

func (q *Queries) GetContext(ctx context.Context, id pgtype.UUID) (Context, error) {
	row := q.db.QueryRow(ctx, getContext, id)
	var i Context
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.CreatedAt,
	)
	return i, err
}

const getContextIDByName = `-- name: GetContextIDByName :one
SELECT id FROM context
WHERE name = $1
`

func (q *Queries) GetContextIDByName(ctx context.Context, name string) (pgtype.UUID, error) {
	row := q.db.QueryRow(ctx, getContextIDByName, name)
	var id pgtype.UUID
	err := row.Scan(&id)
	return id, err
}

const getContextNameByID = `-- name: GetContextNameByID :one
SELECT name FROM context
WHERE id = $1
`

func (q *Queries) GetContextNameByID(ctx context.Context, id pgtype.UUID) (string, error) {
	row := q.db.QueryRow(ctx, getContextNameByID, id)
	var name string
	err := row.Scan(&name)
	return name, err
}

const listContexts = `-- name: ListContexts :many
SELECT id, name, description, created_at FROM context
`

func (q *Queries) ListContexts(ctx context.Context) ([]Context, error) {
	rows, err := q.db.Query(ctx, listContexts)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Context
	for rows.Next() {
		var i Context
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
