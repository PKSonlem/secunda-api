package service

import (
	"context"
	"fmt"

	"github.com/PKSonlem/testtask-secunda-api/internal/domain"
)

type TaskService struct {
	tasks domain.TaskRepository
	teams domain.TeamRepository
	cache domain.TaskCache
}

func NewTaskService(tasks domain.TaskRepository, teams domain.TeamRepository, cache domain.TaskCache) *TaskService {
	return &TaskService{tasks: tasks, teams: teams, cache: cache}
}

func (s *TaskService) Create(ctx context.Context, userID int64, task *domain.Task) (*domain.Task, error) {
	ok, err := s.teams.IsMember(ctx, task.TeamID, userID)
	if err != nil || !ok {
		return nil, domain.ErrForbidden
	}

	task.CreatedBy = userID
	if task.Status == "" {
		task.Status = domain.StatusTodo
	}

	id, err := s.tasks.Create(ctx, task)
	if err != nil {
		return nil, err
	}

	task.ID = id
	_ = s.cache.Invalidate(ctx, teamCacheKey(task.TeamID))

	return task, nil
}

func (s *TaskService) Update(ctx context.Context, userID, taskID int64, update *domain.Task) (*domain.Task, error) {
	existing, err := s.tasks.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	ok, err := s.teams.IsMember(ctx, existing.TeamID, userID)
	if err != nil || !ok {
		return nil, domain.ErrForbidden
	}

	update.ID = taskID
	update.TeamID = existing.TeamID
	if err := s.tasks.Update(ctx, update, userID); err != nil {
		return nil, err
	}

	_ = s.cache.Invalidate(ctx, teamCacheKey(existing.TeamID))

	return update, nil
}

func (s *TaskService) List(ctx context.Context, filter domain.TaskFilter) ([]*domain.Task, int, error) {
	if filter.TeamID != nil {
		key := teamCacheKey(*filter.TeamID)
		if tasks, hit, err := s.cache.GetTeamTasks(ctx, key); hit && err == nil {
			return tasks, len(tasks), nil
		}
	}

	tasks, total, err := s.tasks.List(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	if filter.TeamID != nil {
		_ = s.cache.SetTeamTasks(ctx, teamCacheKey(*filter.TeamID), tasks)
	}

	return tasks, total, nil
}

func (s *TaskService) GetHistory(ctx context.Context, userID, taskID int64) ([]*domain.TaskHistory, error) {
	task, err := s.tasks.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	ok, err := s.teams.IsMember(ctx, task.TeamID, userID)
	if err != nil || !ok {
		return nil, domain.ErrForbidden
	}

	return s.tasks.GetHistory(ctx, taskID)
}

func teamCacheKey(teamID int64) string {
	return fmt.Sprintf("tasks:team:%d", teamID)
}
