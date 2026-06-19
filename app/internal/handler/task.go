package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/PKSonlem/testtask-secunda-api/internal/domain"
	"github.com/PKSonlem/testtask-secunda-api/internal/middleware"
	"github.com/labstack/echo/v4"
)

type taskService interface {
	Create(ctx context.Context, userID int64, task *domain.Task) (*domain.Task, error)
	Update(ctx context.Context, userID, taskID int64, update *domain.Task) (*domain.Task, error)
	List(ctx context.Context, filter domain.TaskFilter) ([]*domain.Task, int, error)
	GetHistory(ctx context.Context, userID, taskID int64) ([]*domain.TaskHistory, error)
}

type TaskHandler struct {
	svc taskService
}

func NewTaskHandler(svc taskService) *TaskHandler {
	return &TaskHandler{svc: svc}
}

func (h *TaskHandler) Create(c echo.Context) error {
	var task domain.Task
	if err := c.Bind(&task); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if task.Status != "" {
		switch task.Status {
		case domain.StatusTodo, domain.StatusInProgress, domain.StatusDone:
		default:
			return echo.NewHTTPError(http.StatusBadRequest, "invalid status")
		}
	}

	created, err := h.svc.Create(c.Request().Context(), middleware.UserIDFromCtx(c), &task)
	if err != nil {
		return taskErr(err)
	}

	return c.JSON(http.StatusCreated, created)
}

func (h *TaskHandler) List(c echo.Context) error {
	filter := domain.TaskFilter{Page: 1, PageSize: 20}
	if v := c.QueryParam("team_id"); v != "" {
		id, _ := strconv.ParseInt(v, 10, 64)
		filter.TeamID = &id
	}
	if v := c.QueryParam("status"); v != "" {
		s := domain.TaskStatus(v)
		switch s {
		case domain.StatusTodo, domain.StatusInProgress, domain.StatusDone:
			filter.Status = &s
		default:
			return echo.NewHTTPError(http.StatusBadRequest, "invalid status")
		}
	}
	if v := c.QueryParam("assignee_id"); v != "" {
		id, _ := strconv.ParseInt(v, 10, 64)
		filter.AssigneeID = &id
	}
	if v := c.QueryParam("page"); v != "" {
		filter.Page, _ = strconv.Atoi(v)
	}

	tasks, total, err := h.svc.List(c.Request().Context(), filter)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, echo.Map{"data": tasks, "total": total})
}

func (h *TaskHandler) Update(c echo.Context) error {
	taskID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid task id")
	}

	var update domain.Task
	if err := c.Bind(&update); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	updated, err := h.svc.Update(c.Request().Context(), middleware.UserIDFromCtx(c), taskID, &update)
	if err != nil {
		return taskErr(err)
	}

	return c.JSON(http.StatusOK, updated)
}

func (h *TaskHandler) GetHistory(c echo.Context) error {
	taskID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid task id")
	}

	history, err := h.svc.GetHistory(c.Request().Context(), middleware.UserIDFromCtx(c), taskID)
	if err != nil {
		return taskErr(err)
	}

	return c.JSON(http.StatusOK, history)
}

func taskErr(err error) error {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return echo.NewHTTPError(http.StatusNotFound, "not found")
	case errors.Is(err, domain.ErrForbidden):
		return echo.NewHTTPError(http.StatusForbidden, "forbidden")
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
}
