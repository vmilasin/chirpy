// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: users.sql

package database

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const createUser = `-- name: CreateUser :one
INSERT INTO users (email, password_hash)
VALUES ($1, $2)
RETURNING id, email, created_at, updated_at
`

type CreateUserParams struct {
	Email        string `json:"email"`
	PasswordHash []byte `json:"password_hash"`
}

type CreateUserRow struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (CreateUserRow, error) {
	row := q.db.QueryRowContext(ctx, createUser, arg.Email, arg.PasswordHash)
	var i CreateUserRow
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
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
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash []byte    `json:"password_hash"`
}

func (q *Queries) GetUserByID(ctx context.Context, id uuid.UUID) (GetUserByIDRow, error) {
	row := q.db.QueryRowContext(ctx, getUserByID, id)
	var i GetUserByIDRow
	err := row.Scan(&i.ID, &i.Email, &i.PasswordHash)
	return i, err
}

const updateUser = `-- name: UpdateUser :one
UPDATE users
SET
    email = COALESCE(NULLIF($1, ''), email),  -- If email is provided, update it, otherwise keep the existing value
    password_hash = COALESCE($2::BYTEA, password_hash)  -- If password is provided, update it, otherwise keep the existing value
WHERE id = $3
RETURNING id, email
`

type UpdateUserParams struct {
	Column1 interface{} `json:"column_1"`
	Column2 []byte      `json:"column_2"`
	ID      uuid.UUID   `json:"id"`
}

type UpdateUserRow struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
}

func (q *Queries) UpdateUser(ctx context.Context, arg UpdateUserParams) (UpdateUserRow, error) {
	row := q.db.QueryRowContext(ctx, updateUser, arg.Column1, arg.Column2, arg.ID)
	var i UpdateUserRow
	err := row.Scan(&i.ID, &i.Email)
	return i, err
}
