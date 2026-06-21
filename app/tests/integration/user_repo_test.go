//go:build integration

package integrationtest

import (
	"context"
	"testing"

	"github.com/PKSonlem/secunda-api/internal/domain"
	mysqlrepo "github.com/PKSonlem/secunda-api/internal/repository/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepo_Create(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	repo := mysqlrepo.NewUserRepository(testDB)

	t.Run("success", func(t *testing.T) {
		id, err := repo.Create(ctx, &domain.User{
			Email:        "alice@example.com",
			PasswordHash: "$2a$10$somehash",
			Name:         "Alice",
		})
		require.NoError(t, err)
		assert.Greater(t, id, int64(0))
	})

	t.Run("duplicate_email", func(t *testing.T) {
		_, err := repo.Create(ctx, &domain.User{Email: "alice@example.com", PasswordHash: "h2", Name: "Alice2"})
		require.Error(t, err, "duplicate email should fail")
	})
}

func TestUserRepo_GetByID(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	repo := mysqlrepo.NewUserRepository(testDB)

	id := insertUser(t, "bob@example.com", "Bob")

	t.Run("found", func(t *testing.T) {
		u, err := repo.GetByID(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, id, u.ID)
		assert.Equal(t, "bob@example.com", u.Email)
		assert.Equal(t, "Bob", u.Name)
	})

	t.Run("not_found", func(t *testing.T) {
		_, err := repo.GetByID(ctx, 99999)
		assert.ErrorIs(t, err, domain.ErrNotFound)
	})
}

func TestUserRepo_GetByEmail(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	repo := mysqlrepo.NewUserRepository(testDB)

	insertUser(t, "carol@example.com", "Carol")

	t.Run("found", func(t *testing.T) {
		u, err := repo.GetByEmail(ctx, "carol@example.com")
		require.NoError(t, err)
		assert.Equal(t, "carol@example.com", u.Email)
	})

	t.Run("not_found", func(t *testing.T) {
		_, err := repo.GetByEmail(ctx, "nobody@example.com")
		assert.ErrorIs(t, err, domain.ErrNotFound)
	})
}
