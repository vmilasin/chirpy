// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package database

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RefreshToken struct {
	ID           int32        `json:"id"`
	UserID       uuid.UUID    `json:"user_id"`
	RefreshToken string       `json:"refresh_token"`
	CreatedAt    sql.NullTime `json:"created_at"`
	ExpiresAt    time.Time    `json:"expires_at"`
	RevokedAt    sql.NullTime `json:"revoked_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	PasswordHash []byte    `json:"password_hash"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
}
