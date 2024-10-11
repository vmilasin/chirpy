// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: users.sql

package database

import (
	"context"

	"github.com/google/uuid"
)

const createUser = `-- name: CreateUser :one
INSERT INTO users (email, password_hash)
VALUES ($1, $2)
RETURNING id, email
`

type CreateUserParams struct {
	Email        string
	PasswordHash []byte
}

type CreateUserRow struct {
	ID    uuid.UUID
	Email string
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (CreateUserRow, error) {
	row := q.db.QueryRowContext(ctx, createUser, arg.Email, arg.PasswordHash)
	var i CreateUserRow
	err := row.Scan(&i.ID, &i.Email)
	return i, err
}

const getPWHash = `-- name: GetPWHash :one
SELECT password_hash
FROM users
WHERE id = $1
`

func (q *Queries) GetPWHash(ctx context.Context, id uuid.UUID) ([]byte, error) {
	row := q.db.QueryRowContext(ctx, getPWHash, id)
	var password_hash []byte
	err := row.Scan(&password_hash)
	return password_hash, err
}

const getUserByEmail = `-- name: GetUserByEmail :one
SELECT id
FROM users
WHERE email = $1
`

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (uuid.UUID, error) {
	row := q.db.QueryRowContext(ctx, getUserByEmail, email)
	var id uuid.UUID
	err := row.Scan(&id)
	return id, err
}

const getUserByID = `-- name: GetUserByID :one
SELECT id, email, password_hash
FROM users
WHERE id = $1
`

type GetUserByIDRow struct {
	ID           uuid.UUID
	Email        string
	PasswordHash []byte
}

func (q *Queries) GetUserByID(ctx context.Context, id uuid.UUID) (GetUserByIDRow, error) {
	row := q.db.QueryRowContext(ctx, getUserByID, id)
	var i GetUserByIDRow
	err := row.Scan(&i.ID, &i.Email, &i.PasswordHash)
	return i, err
}
