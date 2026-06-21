//go:build integration

package integrationtest

import (
	"context"
	"fmt"
	"testing"

	"github.com/PKSonlem/secunda-api/internal/domain"
	mysqlrepo "github.com/PKSonlem/secunda-api/internal/repository/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTaskRepo(t *testing.T) (repo *mysqlrepo.TaskRepository, userID, teamID int64) {
	t.Helper()
	cleanDB(t)
	userID = insertUser(t, "dave@example.com", "Dave")
	teamID = insertTeam(t, "Dev Team", userID)
	txMgr := mysqlrepo.NewTxManager(testDB)
	return mysqlrepo.NewTaskRepository(testDB, txMgr), userID, teamID
}

func TestTaskRepo_Create(t *testing.T) {
	repo, userID, teamID := setupTaskRepo(t)
	ctx := context.Background()

	id, err := repo.Create(ctx, &domain.Task{
		Title:       "Implement login",
		Description: "JWT-based login flow",
		Status:      domain.StatusTodo,
		TeamID:      teamID,
		CreatedBy:   userID,
	})

	require.NoError(t, err)
	assert.Greater(t, id, int64(0))
}

func TestTaskRepo_GetByID(t *testing.T) {
	repo, userID, teamID := setupTaskRepo(t)
	ctx := context.Background()

	id, err := repo.Create(ctx, &domain.Task{
		Title:     "Fix bug #42",
		Status:    domain.StatusInProgress,
		TeamID:    teamID,
		CreatedBy: userID,
	})
	require.NoError(t, err)

	t.Run("found", func(t *testing.T) {
		got, err := repo.GetByID(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, id, got.ID)
		assert.Equal(t, "Fix bug #42", got.Title)
		assert.Equal(t, domain.StatusInProgress, got.Status)
	})

	t.Run("not_found", func(t *testing.T) {
		_, err := repo.GetByID(ctx, 99999)
		assert.ErrorIs(t, err, domain.ErrNotFound)
	})
}

func TestTaskRepo_Update(t *testing.T) {
	repo, userID, teamID := setupTaskRepo(t)
	ctx := context.Background()

	id, err := repo.Create(ctx, &domain.Task{
		Title:     "Original title",
		Status:    domain.StatusTodo,
		TeamID:    teamID,
		CreatedBy: userID,
	})
	require.NoError(t, err)

	err = repo.Update(ctx, &domain.Task{
		ID:          id,
		Title:       "Updated title",
		Description: "Some description",
		Status:      domain.StatusDone,
		TeamID:      teamID,
	}, userID)
	require.NoError(t, err)

	got, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, "Updated title", got.Title)
	assert.Equal(t, domain.StatusDone, got.Status)

	history, err := repo.GetHistory(ctx, id)
	require.NoError(t, err)
	require.NotEmpty(t, history)

	var statusChange *domain.TaskHistory
	for _, h := range history {
		if h.FieldName == "status" {
			statusChange = h
			break
		}
	}
	require.NotNil(t, statusChange, "status change should be in history")
	assert.Equal(t, "todo", statusChange.OldValue)
	assert.Equal(t, "done", statusChange.NewValue)
}

func TestTaskRepo_List(t *testing.T) {
	repo, userID, teamID := setupTaskRepo(t)
	ctx := context.Background()

	statuses := []domain.TaskStatus{domain.StatusTodo, domain.StatusTodo, domain.StatusDone}
	for i, s := range statuses {
		_, err := repo.Create(ctx, &domain.Task{
			Title:     fmt.Sprintf("Task %d", i+1),
			Status:    s,
			TeamID:    teamID,
			CreatedBy: userID,
		})
		require.NoError(t, err)
	}

	t.Run("all_tasks_in_team", func(t *testing.T) {
		tasks, total, err := repo.List(ctx, domain.TaskFilter{TeamID: &teamID, Page: 1, PageSize: 10})
		require.NoError(t, err)
		assert.Equal(t, 3, total)
		assert.Len(t, tasks, 3)
	})

	t.Run("filter_by_status", func(t *testing.T) {
		s := domain.StatusTodo
		tasks, total, err := repo.List(ctx, domain.TaskFilter{TeamID: &teamID, Status: &s, Page: 1, PageSize: 10})
		require.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, tasks, 2)
	})

	t.Run("pagination", func(t *testing.T) {
		tasks, total, err := repo.List(ctx, domain.TaskFilter{TeamID: &teamID, Page: 1, PageSize: 2})
		require.NoError(t, err)
		assert.Equal(t, 3, total)
		assert.Len(t, tasks, 2)
	})
}
