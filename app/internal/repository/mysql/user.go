package mysql

import (
	"context"
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/PKSonlem/testtask-secunda-api/internal/domain"
)

type UserRepository struct{ db *sql.DB }

func NewUserRepository(db *sql.DB) *UserRepository { return &UserRepository{db: db} }

func (r *UserRepository) Create(ctx context.Context, u *domain.User) (int64, error) {
	res, err := builder.
		Insert("users").
		Columns("email", "password_hash", "name").
		Values(u.Email, u.PasswordHash, u.Name).
		RunWith(r.db).ExecContext(ctx)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	u := &domain.User{}
	err := builder.
		Select("id", "email", "password_hash", "name", "created_at").
		From("users").
		Where(sq.Eq{"id": id}).
		RunWith(r.db).QueryRowContext(ctx).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return u, err
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	u := &domain.User{}
	err := builder.
		Select("id", "email", "password_hash", "name", "created_at").
		From("users").
		Where(sq.Eq{"email": email}).
		RunWith(r.db).QueryRowContext(ctx).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return u, err
}
