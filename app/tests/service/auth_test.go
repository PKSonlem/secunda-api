package servicetest

import (
	"context"
	"testing"
	"time"

	"github.com/PKSonlem/testtask-secunda-api/internal/config"
	"github.com/PKSonlem/testtask-secunda-api/internal/domain"
	"github.com/PKSonlem/testtask-secunda-api/internal/service"
	"github.com/PKSonlem/testtask-secunda-api/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func jwtCfg() config.JWTConfig {
	return config.JWTConfig{Secret: "test-secret-key", AccessTokenTTL: time.Hour}
}

func TestRegister(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := &mocks.UserRepo{
			CreateFn: func(_ context.Context, u *domain.User) (int64, error) {
				assert.Equal(t, "test@example.com", u.Email)
				assert.Equal(t, "Test User", u.Name)
				assert.NotEmpty(t, u.PasswordHash)
				assert.NotEqual(t, "password123", u.PasswordHash)
				return 42, nil
			},
		}

		svc := service.NewAuthService(repo, jwtCfg())
		user, err := svc.Register(ctx, "test@example.com", "password123", "Test User")

		require.NoError(t, err)
		assert.Equal(t, int64(42), user.ID)
		assert.Equal(t, "test@example.com", user.Email)
	})

	t.Run("repo_error", func(t *testing.T) {
		repo := &mocks.UserRepo{
			CreateFn: func(_ context.Context, _ *domain.User) (int64, error) {
				return 0, domain.ErrAlreadyExists
			},
		}

		svc := service.NewAuthService(repo, jwtCfg())
		_, err := svc.Register(ctx, "test@example.com", "password123", "Test User")

		assert.ErrorIs(t, err, domain.ErrAlreadyExists)
	})
}

func TestLogin(t *testing.T) {
	ctx := context.Background()

	// register first to get a real bcrypt hash
	var storedUser *domain.User
	regRepo := &mocks.UserRepo{
		CreateFn: func(_ context.Context, u *domain.User) (int64, error) {
			storedUser = &domain.User{ID: 1, Email: u.Email, PasswordHash: u.PasswordHash, Name: u.Name}
			return 1, nil
		},
	}
	_, err := service.NewAuthService(regRepo, jwtCfg()).Register(ctx, "login@example.com", "mypassword", "Login User")
	require.NoError(t, err)
	require.NotNil(t, storedUser)

	t.Run("success", func(t *testing.T) {
		repo := &mocks.UserRepo{
			GetByEmailFn: func(_ context.Context, _ string) (*domain.User, error) {
				return storedUser, nil
			},
		}
		token, err := service.NewAuthService(repo, jwtCfg()).Login(ctx, "login@example.com", "mypassword")

		require.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("user_not_found", func(t *testing.T) {
		repo := &mocks.UserRepo{
			GetByEmailFn: func(_ context.Context, _ string) (*domain.User, error) {
				return nil, domain.ErrNotFound
			},
		}
		_, err := service.NewAuthService(repo, jwtCfg()).Login(ctx, "nobody@example.com", "pass")

		assert.ErrorIs(t, err, domain.ErrUnauthorized)
	})

	t.Run("wrong_password", func(t *testing.T) {
		repo := &mocks.UserRepo{
			GetByEmailFn: func(_ context.Context, _ string) (*domain.User, error) {
				return storedUser, nil
			},
		}
		_, err := service.NewAuthService(repo, jwtCfg()).Login(ctx, "login@example.com", "wrongpassword")

		assert.ErrorIs(t, err, domain.ErrUnauthorized)
	})
}

func TestValidateToken(t *testing.T) {
	ctx := context.Background()

	var storedUser *domain.User
	repo := &mocks.UserRepo{
		CreateFn: func(_ context.Context, u *domain.User) (int64, error) {
			storedUser = &domain.User{ID: 7, Email: u.Email, PasswordHash: u.PasswordHash}
			return 7, nil
		},
		GetByEmailFn: func(_ context.Context, _ string) (*domain.User, error) {
			return storedUser, nil
		},
	}

	svc := service.NewAuthService(repo, jwtCfg())
	_, err := svc.Register(ctx, "validate@example.com", "pass", "V")
	require.NoError(t, err)

	token, err := svc.Login(ctx, "validate@example.com", "pass")
	require.NoError(t, err)

	t.Run("valid_token", func(t *testing.T) {
		userID, err := svc.ValidateToken(token)
		require.NoError(t, err)
		assert.Equal(t, int64(7), userID)
	})

	t.Run("invalid_token", func(t *testing.T) {
		_, err := svc.ValidateToken("not.a.token")
		assert.ErrorIs(t, err, domain.ErrUnauthorized)
	})

	t.Run("wrong_secret", func(t *testing.T) {
		otherSvc := service.NewAuthService(repo, config.JWTConfig{Secret: "different", AccessTokenTTL: time.Hour})
		_, err := otherSvc.ValidateToken(token)
		assert.ErrorIs(t, err, domain.ErrUnauthorized)
	})

	t.Run("expired_token", func(t *testing.T) {
		expiredSvc := service.NewAuthService(repo, config.JWTConfig{
			Secret:         "test-secret-key",
			AccessTokenTTL: -time.Hour,
		})
		expiredToken, err := expiredSvc.Login(ctx, "validate@example.com", "pass")
		require.NoError(t, err)

		_, err = svc.ValidateToken(expiredToken)
		assert.ErrorIs(t, err, domain.ErrUnauthorized)
	})
}
