package servicetest

import (
	"context"
	"testing"

	"github.com/PKSonlem/testtask-secunda-api/internal/domain"
	"github.com/PKSonlem/testtask-secunda-api/internal/service"
	"github.com/PKSonlem/testtask-secunda-api/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskCreate(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		teamRepo := &mocks.TeamRepo{
			IsMemberFn: func(_ context.Context, _, _ int64) (bool, error) { return true, nil },
		}
		var captured *domain.Task
		taskRepo := &mocks.TaskRepo{
			CreateFn: func(_ context.Context, task *domain.Task) (int64, error) {
				captured = task
				return 10, nil
			},
		}
		invalidated := false
		cache := &mocks.TaskCache{
			InvalidateFn: func(_ context.Context, _ string) error {
				invalidated = true
				return nil
			},
		}

		svc := service.NewTaskService(taskRepo, teamRepo, cache)
		task, err := svc.Create(ctx, 5, &domain.Task{Title: "Do something", TeamID: 1})

		require.NoError(t, err)
		assert.Equal(t, int64(10), task.ID)
		assert.Equal(t, domain.StatusTodo, captured.Status) // default status
		assert.Equal(t, int64(5), captured.CreatedBy)
		assert.True(t, invalidated)
	})

	t.Run("not_member", func(t *testing.T) {
		teamRepo := &mocks.TeamRepo{
			IsMemberFn: func(_ context.Context, _, _ int64) (bool, error) { return false, nil },
		}
		svc := service.NewTaskService(&mocks.TaskRepo{}, teamRepo, &mocks.TaskCache{})
		_, err := svc.Create(ctx, 5, &domain.Task{Title: "test", TeamID: 1})

		assert.ErrorIs(t, err, domain.ErrForbidden)
	})

	t.Run("default_status_todo", func(t *testing.T) {
		teamRepo := &mocks.TeamRepo{
			IsMemberFn: func(_ context.Context, _, _ int64) (bool, error) { return true, nil },
		}
		var captured *domain.Task
		taskRepo := &mocks.TaskRepo{
			CreateFn: func(_ context.Context, task *domain.Task) (int64, error) {
				captured = task
				return 1, nil
			},
		}

		svc := service.NewTaskService(taskRepo, teamRepo, &mocks.TaskCache{})
		_, err := svc.Create(ctx, 1, &domain.Task{Title: "no status", TeamID: 1})

		require.NoError(t, err)
		assert.Equal(t, domain.StatusTodo, captured.Status)
	})
}

func TestTaskUpdate(t *testing.T) {
	ctx := context.Background()
	existing := &domain.Task{ID: 3, TeamID: 2, Title: "old title", Status: domain.StatusTodo}

	t.Run("success", func(t *testing.T) {
		taskRepo := &mocks.TaskRepo{
			GetByIDFn: func(_ context.Context, _ int64) (*domain.Task, error) { return existing, nil },
			UpdateFn:  func(_ context.Context, _ *domain.Task, _ int64) error { return nil },
		}
		teamRepo := &mocks.TeamRepo{
			IsMemberFn: func(_ context.Context, _, _ int64) (bool, error) { return true, nil },
		}
		invalidated := false
		cache := &mocks.TaskCache{
			InvalidateFn: func(_ context.Context, _ string) error {
				invalidated = true
				return nil
			},
		}

		svc := service.NewTaskService(taskRepo, teamRepo, cache)
		updated, err := svc.Update(ctx, 1, 3, &domain.Task{Title: "new title", Status: domain.StatusDone})

		require.NoError(t, err)
		assert.Equal(t, int64(3), updated.ID)
		assert.True(t, invalidated)
	})

	t.Run("task_not_found", func(t *testing.T) {
		taskRepo := &mocks.TaskRepo{
			GetByIDFn: func(_ context.Context, _ int64) (*domain.Task, error) {
				return nil, domain.ErrNotFound
			},
		}
		svc := service.NewTaskService(taskRepo, &mocks.TeamRepo{}, &mocks.TaskCache{})
		_, err := svc.Update(ctx, 1, 999, &domain.Task{})

		assert.ErrorIs(t, err, domain.ErrNotFound)
	})

	t.Run("not_member", func(t *testing.T) {
		taskRepo := &mocks.TaskRepo{
			GetByIDFn: func(_ context.Context, _ int64) (*domain.Task, error) { return existing, nil },
		}
		teamRepo := &mocks.TeamRepo{
			IsMemberFn: func(_ context.Context, _, _ int64) (bool, error) { return false, nil },
		}
		svc := service.NewTaskService(taskRepo, teamRepo, &mocks.TaskCache{})
		_, err := svc.Update(ctx, 99, 3, &domain.Task{})

		assert.ErrorIs(t, err, domain.ErrForbidden)
	})
}

func TestTaskList(t *testing.T) {
	ctx := context.Background()
	teamID := int64(1)

	t.Run("cache_hit", func(t *testing.T) {
		cached := []*domain.Task{{ID: 1, Title: "from cache"}}
		dbCalled := false

		cache := &mocks.TaskCache{
			GetTeamTasksFn: func(_ context.Context, _ string) ([]*domain.Task, bool, error) {
				return cached, true, nil
			},
		}
		taskRepo := &mocks.TaskRepo{
			ListFn: func(_ context.Context, _ domain.TaskFilter) ([]*domain.Task, int, error) {
				dbCalled = true
				return nil, 0, nil
			},
		}

		svc := service.NewTaskService(taskRepo, &mocks.TeamRepo{}, cache)
		tasks, total, err := svc.List(ctx, domain.TaskFilter{TeamID: &teamID})

		require.NoError(t, err)
		assert.Equal(t, cached, tasks)
		assert.Equal(t, 1, total)
		assert.False(t, dbCalled, "DB should not be called on cache hit")
	})

	t.Run("cache_miss_falls_through_to_db", func(t *testing.T) {
		dbTasks := []*domain.Task{{ID: 2}, {ID: 3}}
		setCalled := false

		cache := &mocks.TaskCache{
			GetTeamTasksFn: func(_ context.Context, _ string) ([]*domain.Task, bool, error) {
				return nil, false, nil
			},
			SetTeamTasksFn: func(_ context.Context, _ string, _ []*domain.Task) error {
				setCalled = true
				return nil
			},
		}
		taskRepo := &mocks.TaskRepo{
			ListFn: func(_ context.Context, _ domain.TaskFilter) ([]*domain.Task, int, error) {
				return dbTasks, 2, nil
			},
		}

		svc := service.NewTaskService(taskRepo, &mocks.TeamRepo{}, cache)
		tasks, total, err := svc.List(ctx, domain.TaskFilter{TeamID: &teamID})

		require.NoError(t, err)
		assert.Equal(t, dbTasks, tasks)
		assert.Equal(t, 2, total)
		assert.True(t, setCalled)
	})

	t.Run("cache_error_falls_through", func(t *testing.T) {
		dbTasks := []*domain.Task{{ID: 5}}

		cache := &mocks.TaskCache{
			GetTeamTasksFn: func(_ context.Context, _ string) ([]*domain.Task, bool, error) {
				return nil, false, assert.AnError
			},
		}
		taskRepo := &mocks.TaskRepo{
			ListFn: func(_ context.Context, _ domain.TaskFilter) ([]*domain.Task, int, error) {
				return dbTasks, 1, nil
			},
		}

		svc := service.NewTaskService(taskRepo, &mocks.TeamRepo{}, cache)
		tasks, _, err := svc.List(ctx, domain.TaskFilter{TeamID: &teamID})

		require.NoError(t, err, "cache error should not bubble up")
		assert.Equal(t, dbTasks, tasks)
	})
}

func TestTaskGetHistory(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		history := []*domain.TaskHistory{{ID: 1, FieldName: "status", OldValue: "todo", NewValue: "done"}}
		taskRepo := &mocks.TaskRepo{
			GetByIDFn:    func(_ context.Context, _ int64) (*domain.Task, error) { return &domain.Task{ID: 1, TeamID: 1}, nil },
			GetHistoryFn: func(_ context.Context, _ int64) ([]*domain.TaskHistory, error) { return history, nil },
		}
		teamRepo := &mocks.TeamRepo{
			IsMemberFn: func(_ context.Context, _, _ int64) (bool, error) { return true, nil },
		}

		svc := service.NewTaskService(taskRepo, teamRepo, &mocks.TaskCache{})
		result, err := svc.GetHistory(ctx, 1, 1)

		require.NoError(t, err)
		assert.Equal(t, history, result)
	})

	t.Run("not_member", func(t *testing.T) {
		taskRepo := &mocks.TaskRepo{
			GetByIDFn: func(_ context.Context, _ int64) (*domain.Task, error) { return &domain.Task{ID: 1, TeamID: 1}, nil },
		}
		teamRepo := &mocks.TeamRepo{
			IsMemberFn: func(_ context.Context, _, _ int64) (bool, error) { return false, nil },
		}

		svc := service.NewTaskService(taskRepo, teamRepo, &mocks.TaskCache{})
		_, err := svc.GetHistory(ctx, 99, 1)

		assert.ErrorIs(t, err, domain.ErrForbidden)
	})
}
