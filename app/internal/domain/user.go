package domain

import (
	"context"
	"time"
)

type User struct {
	ID           int64     `json:"id"         db:"id"`
	Email        string    `json:"email"      db:"email"`
	PasswordHash string    `json:"-"          db:"password_hash"`
	Name         string    `json:"name"       db:"name"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type UserRepository interface {
	Create(ctx context.Context, user *User) (int64, error)
	GetByID(ctx context.Context, id int64) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
}
