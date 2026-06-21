package mocks

import (
	"context"

	"github.com/PKSonlem/testtask-secunda-api/internal/domain"
)

// AuthSvc mocks the authService interface used by AuthHandler.
type AuthSvc struct {
	RegisterFn      func(ctx context.Context, email, password, name string) (*domain.User, error)
	LoginFn         func(ctx context.Context, email, password string) (string, error)
	ValidateTokenFn func(token string) (int64, error)
}

func (m *AuthSvc) Register(ctx context.Context, email, password, name string) (*domain.User, error) {
	if m.RegisterFn != nil {
		return m.RegisterFn(ctx, email, password, name)
	}
	return nil, nil
}

func (m *AuthSvc) Login(ctx context.Context, email, password string) (string, error) {
	if m.LoginFn != nil {
		return m.LoginFn(ctx, email, password)
	}
	return "", nil
}

func (m *AuthSvc) ValidateToken(token string) (int64, error) {
	if m.ValidateTokenFn != nil {
		return m.ValidateTokenFn(token)
	}
	return 0, nil
}

// TeamSvc mocks the teamService interface used by TeamHandler.
type TeamSvc struct {
	CreateFn func(ctx context.Context, userID int64, name string) (*domain.Team, error)
	ListFn   func(ctx context.Context, userID int64) ([]*domain.Team, error)
	InviteFn func(ctx context.Context, callerID, teamID int64, inviteeEmail string) error
}

func (m *TeamSvc) Create(ctx context.Context, userID int64, name string) (*domain.Team, error) {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, userID, name)
	}
	return nil, nil
}

func (m *TeamSvc) List(ctx context.Context, userID int64) ([]*domain.Team, error) {
	if m.ListFn != nil {
		return m.ListFn(ctx, userID)
	}
	return nil, nil
}

func (m *TeamSvc) Invite(ctx context.Context, callerID, teamID int64, inviteeEmail string) error {
	if m.InviteFn != nil {
		return m.InviteFn(ctx, callerID, teamID, inviteeEmail)
	}
	return nil
}

// TaskSvc mocks the taskService interface used by TaskHandler.
type TaskSvc struct {
	CreateFn     func(ctx context.Context, userID int64, task *domain.Task) (*domain.Task, error)
	UpdateFn     func(ctx context.Context, userID, taskID int64, update *domain.Task) (*domain.Task, error)
	ListFn       func(ctx context.Context, filter domain.TaskFilter) ([]*domain.Task, int, error)
	GetHistoryFn func(ctx context.Context, userID, taskID int64) ([]*domain.TaskHistory, error)
}

func (m *TaskSvc) Create(ctx context.Context, userID int64, task *domain.Task) (*domain.Task, error) {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, userID, task)
	}
	return nil, nil
}

func (m *TaskSvc) Update(ctx context.Context, userID, taskID int64, update *domain.Task) (*domain.Task, error) {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, userID, taskID, update)
	}
	return nil, nil
}

func (m *TaskSvc) List(ctx context.Context, filter domain.TaskFilter) ([]*domain.Task, int, error) {
	if m.ListFn != nil {
		return m.ListFn(ctx, filter)
	}
	return nil, 0, nil
}

func (m *TaskSvc) GetHistory(ctx context.Context, userID, taskID int64) ([]*domain.TaskHistory, error) {
	if m.GetHistoryFn != nil {
		return m.GetHistoryFn(ctx, userID, taskID)
	}
	return nil, nil
}
