package domain

import (
	"context"
	"time"
)

type TaskStatus string

const (
	StatusTodo       TaskStatus = "todo"
	StatusInProgress TaskStatus = "in_progress"
	StatusDone       TaskStatus = "done"
)


type Task struct {
	ID          int64      `json:"id"          db:"id"`
	Title       string     `json:"title"       db:"title"`
	Description string     `json:"description" db:"description"`
	Status      TaskStatus `json:"status"      db:"status"`
	TeamID      int64      `json:"team_id"     db:"team_id"`
	AssigneeID  *int64     `json:"assignee_id" db:"assignee_id"`
	CreatedBy   int64      `json:"created_by"  db:"created_by"`
	CreatedAt   time.Time  `json:"created_at"  db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"  db:"updated_at"`
}

type TaskHistory struct {
	ID        int64     `json:"id"         db:"id"`
	TaskID    int64     `json:"task_id"    db:"task_id"`
	ChangedBy int64     `json:"changed_by" db:"changed_by"`
	FieldName string    `json:"field_name" db:"field_name"`
	OldValue  string    `json:"old_value"  db:"old_value"`
	NewValue  string    `json:"new_value"  db:"new_value"`
	ChangedAt time.Time `json:"changed_at" db:"changed_at"`
}

type TaskComment struct {
	ID        int64     `json:"id"         db:"id"`
	TaskID    int64     `json:"task_id"    db:"task_id"`
	UserID    int64     `json:"user_id"    db:"user_id"`
	Body      string    `json:"body"       db:"body"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type TaskFilter struct {
	TeamID     *int64
	Status     *TaskStatus
	AssigneeID *int64
	Page       int
	PageSize   int
}

type TaskRepository interface {
	Create(ctx context.Context, task *Task) (int64, error)
	GetByID(ctx context.Context, id int64) (*Task, error)
	Update(ctx context.Context, task *Task, changedBy int64) error
	List(ctx context.Context, filter TaskFilter) ([]*Task, int, error)
	GetHistory(ctx context.Context, taskID int64) ([]*TaskHistory, error)
}

type TaskCache interface {
	GetTeamTasks(ctx context.Context, key string) ([]*Task, bool, error)
	SetTeamTasks(ctx context.Context, key string, tasks []*Task) error
	Invalidate(ctx context.Context, key string) error
}
