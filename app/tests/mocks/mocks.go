package mocks

import (
	"context"

	"github.com/PKSonlem/testtask-secunda-api/internal/domain"
)

// UserRepo implements domain.UserRepository.
type UserRepo struct {
	CreateFn     func(ctx context.Context, u *domain.User) (int64, error)
	GetByIDFn    func(ctx context.Context, id int64) (*domain.User, error)
	GetByEmailFn func(ctx context.Context, email string) (*domain.User, error)
}

func (m *UserRepo) Create(ctx context.Context, u *domain.User) (int64, error) {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, u)
	}
	return 0, nil
}

func (m *UserRepo) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.GetByEmailFn != nil {
		return m.GetByEmailFn(ctx, email)
	}
	return nil, nil
}

// TeamRepo implements domain.TeamRepository.
type TeamRepo struct {
	CreateFn       func(ctx context.Context, team *domain.Team) (int64, error)
	GetByIDFn      func(ctx context.Context, id int64) (*domain.Team, error)
	ListByUserIDFn func(ctx context.Context, userID int64) ([]*domain.Team, error)
	AddMemberFn    func(ctx context.Context, member *domain.TeamMember) error
	GetMemberFn    func(ctx context.Context, teamID, userID int64) (*domain.TeamMember, error)
	IsMemberFn     func(ctx context.Context, teamID, userID int64) (bool, error)
}

func (m *TeamRepo) Create(ctx context.Context, team *domain.Team) (int64, error) {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, team)
	}
	return 0, nil
}

func (m *TeamRepo) GetByID(ctx context.Context, id int64) (*domain.Team, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *TeamRepo) ListByUserID(ctx context.Context, userID int64) ([]*domain.Team, error) {
	if m.ListByUserIDFn != nil {
		return m.ListByUserIDFn(ctx, userID)
	}
	return nil, nil
}

func (m *TeamRepo) AddMember(ctx context.Context, member *domain.TeamMember) error {
	if m.AddMemberFn != nil {
		return m.AddMemberFn(ctx, member)
	}
	return nil
}

func (m *TeamRepo) GetMember(ctx context.Context, teamID, userID int64) (*domain.TeamMember, error) {
	if m.GetMemberFn != nil {
		return m.GetMemberFn(ctx, teamID, userID)
	}
	return nil, nil
}

func (m *TeamRepo) IsMember(ctx context.Context, teamID, userID int64) (bool, error) {
	if m.IsMemberFn != nil {
		return m.IsMemberFn(ctx, teamID, userID)
	}
	return false, nil
}

// TaskRepo implements domain.TaskRepository.
type TaskRepo struct {
	CreateFn     func(ctx context.Context, t *domain.Task) (int64, error)
	GetByIDFn    func(ctx context.Context, id int64) (*domain.Task, error)
	UpdateFn     func(ctx context.Context, task *domain.Task, changedBy int64) error
	ListFn       func(ctx context.Context, f domain.TaskFilter) ([]*domain.Task, int, error)
	GetHistoryFn func(ctx context.Context, taskID int64) ([]*domain.TaskHistory, error)
}

func (m *TaskRepo) Create(ctx context.Context, t *domain.Task) (int64, error) {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, t)
	}
	return 0, nil
}

func (m *TaskRepo) GetByID(ctx context.Context, id int64) (*domain.Task, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *TaskRepo) Update(ctx context.Context, task *domain.Task, changedBy int64) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, task, changedBy)
	}
	return nil
}

func (m *TaskRepo) List(ctx context.Context, f domain.TaskFilter) ([]*domain.Task, int, error) {
	if m.ListFn != nil {
		return m.ListFn(ctx, f)
	}
	return nil, 0, nil
}

func (m *TaskRepo) GetHistory(ctx context.Context, taskID int64) ([]*domain.TaskHistory, error) {
	if m.GetHistoryFn != nil {
		return m.GetHistoryFn(ctx, taskID)
	}
	return nil, nil
}

// TaskCache implements domain.TaskCache.
type TaskCache struct {
	GetTeamTasksFn func(ctx context.Context, key string) ([]*domain.Task, bool, error)
	SetTeamTasksFn func(ctx context.Context, key string, tasks []*domain.Task) error
	InvalidateFn   func(ctx context.Context, key string) error
}

func (m *TaskCache) GetTeamTasks(ctx context.Context, key string) ([]*domain.Task, bool, error) {
	if m.GetTeamTasksFn != nil {
		return m.GetTeamTasksFn(ctx, key)
	}
	return nil, false, nil
}

func (m *TaskCache) SetTeamTasks(ctx context.Context, key string, tasks []*domain.Task) error {
	if m.SetTeamTasksFn != nil {
		return m.SetTeamTasksFn(ctx, key, tasks)
	}
	return nil
}

func (m *TaskCache) Invalidate(ctx context.Context, key string) error {
	if m.InvalidateFn != nil {
		return m.InvalidateFn(ctx, key)
	}
	return nil
}

// EmailSender satisfies the unexported emailSender interface in the service package.
// Works via Go's structural typing — caller doesn't need to know the interface name.
type EmailSender struct {
	SendInviteFn func(ctx context.Context, email, teamName string) error
}

func (m *EmailSender) SendInvite(ctx context.Context, email, teamName string) error {
	if m.SendInviteFn != nil {
		return m.SendInviteFn(ctx, email, teamName)
	}
	return nil
}
