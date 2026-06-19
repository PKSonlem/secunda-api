package service

import (
	"context"
	"fmt"
	"time"

	"github.com/PKSonlem/testtask-secunda-api/internal/config"
	"github.com/PKSonlem/testtask-secunda-api/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	users  domain.UserRepository
	secret []byte
	ttl    time.Duration
}

func NewAuthService(users domain.UserRepository, cfg config.JWTConfig) *AuthService {
	return &AuthService{
		users:  users,
		secret: []byte(cfg.Secret),
		ttl:    cfg.AccessTokenTTL,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password, name string) (*domain.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &domain.User{Email: email, PasswordHash: string(hash), Name: name}
	id, err := s.users.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	user.ID = id

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return "", domain.ErrUnauthorized
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", domain.ErrUnauthorized
	}

	return s.signToken(user.ID)
}

func (s *AuthService) ValidateToken(tokenStr string) (int64, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return s.secret, nil
	})

	if err != nil || !token.Valid {
		return 0, domain.ErrUnauthorized
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, domain.ErrUnauthorized
	}

	sub, ok := claims["sub"].(float64)
	if !ok {
		return 0, domain.ErrUnauthorized
	}

	return int64(sub), nil
}

func (s *AuthService) signToken(userID int64) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(s.ttl).Unix(),
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
}
