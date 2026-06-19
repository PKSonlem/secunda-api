package domain

import "context"

type AuthService interface {
	Register(ctx context.Context, email, password, name string) (*User, error)
	Login(ctx context.Context, email, password string) (string, error)
	ValidateToken(token string) (int64, error)
}

type TaskService interface {
	Create(ctx context.Context, userID int64, task *Task) (*Task, error)
	Update(ctx context.Context, userID, taskID int64, update *Task) (*Task, error)
	List(ctx context.Context, filter TaskFilter) ([]*Task, int, error)
	GetHistory(ctx context.Context, userID, taskID int64) ([]*TaskHistory, error)
}

type TeamService interface {
	Create(ctx context.Context, userID int64, name string) (*Team, error)
	List(ctx context.Context, userID int64) ([]*Team, error)
	Invite(ctx context.Context, callerID, teamID int64, inviteeEmail string) error
}
