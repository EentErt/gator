// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: generate.sql

package database

import (
	"context"
)

const generate = `-- name: Generate :exec
DELETE FROM users
`

func (q *Queries) Generate(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, generate)
	return err
}
