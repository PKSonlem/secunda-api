package handlertest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/PKSonlem/testtask-secunda-api/internal/domain"
	"github.com/PKSonlem/testtask-secunda-api/internal/handler"
	"github.com/PKSonlem/testtask-secunda-api/tests/mocks"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskHandler_Create(t *testing.T) {
	e := newEcho()

	t.Run("success", func(t *testing.T) {
		svc := &mocks.TaskSvc{
			CreateFn: func(_ context.Context, _ int64, task *domain.Task) (*domain.Task, error) {
				task.ID = 10
				return task, nil
			},
		}
		req := httptest.NewRequest(http.MethodPost, "/tasks",
			strings.NewReader(`{"title":"Fix tests","team_id":1}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		call(e, handler.NewTaskHandler(svc).Create, withUserID(e, req, rec, 1))

		assert.Equal(t, http.StatusCreated, rec.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, float64(10), resp["id"])
	})

	t.Run("invalid_status", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/tasks",
			strings.NewReader(`{"title":"T","status":"unknown"}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		call(e, handler.NewTaskHandler(&mocks.TaskSvc{}).Create, withUserID(e, req, rec, 1))

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("forbidden", func(t *testing.T) {
		svc := &mocks.TaskSvc{
			CreateFn: func(_ context.Context, _ int64, _ *domain.Task) (*domain.Task, error) {
				return nil, domain.ErrForbidden
			},
		}
		req := httptest.NewRequest(http.MethodPost, "/tasks",
			strings.NewReader(`{"title":"T","team_id":1}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		call(e, handler.NewTaskHandler(svc).Create, withUserID(e, req, rec, 1))

		assert.Equal(t, http.StatusForbidden, rec.Code)
	})
}

func TestTaskHandler_List(t *testing.T) {
	e := newEcho()

	t.Run("success_with_filters", func(t *testing.T) {
		svc := &mocks.TaskSvc{
			ListFn: func(_ context.Context, f domain.TaskFilter) ([]*domain.Task, int, error) {
				assert.NotNil(t, f.TeamID)
				assert.Equal(t, int64(1), *f.TeamID)
				return []*domain.Task{{ID: 1}, {ID: 2}}, 2, nil
			},
		}
		req := httptest.NewRequest(http.MethodGet, "/tasks?team_id=1", nil)
		rec := httptest.NewRecorder()

		call(e, handler.NewTaskHandler(svc).List, withUserID(e, req, rec, 1))

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, float64(2), resp["total"])
	})

	t.Run("invalid_status_filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/tasks?status=invalid", nil)
		rec := httptest.NewRecorder()

		call(e, handler.NewTaskHandler(&mocks.TaskSvc{}).List, withUserID(e, req, rec, 1))

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestTaskHandler_Update(t *testing.T) {
	e := newEcho()

	t.Run("success", func(t *testing.T) {
		updated := &domain.Task{ID: 3, Title: "Updated", Status: domain.StatusDone}
		svc := &mocks.TaskSvc{
			UpdateFn: func(_ context.Context, _, _ int64, _ *domain.Task) (*domain.Task, error) {
				return updated, nil
			},
		}
		req := httptest.NewRequest(http.MethodPut, "/tasks/3",
			strings.NewReader(`{"title":"Updated","status":"done"}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := withUserID(e, req, rec, 1)
		c.SetParamNames("id")
		c.SetParamValues("3")

		call(e, handler.NewTaskHandler(svc).Update, c)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("task_not_found", func(t *testing.T) {
		svc := &mocks.TaskSvc{
			UpdateFn: func(_ context.Context, _, _ int64, _ *domain.Task) (*domain.Task, error) {
				return nil, domain.ErrNotFound
			},
		}
		req := httptest.NewRequest(http.MethodPut, "/tasks/999",
			strings.NewReader(`{"title":"x"}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := withUserID(e, req, rec, 1)
		c.SetParamNames("id")
		c.SetParamValues("999")

		call(e, handler.NewTaskHandler(svc).Update, c)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("invalid_task_id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/tasks/abc",
			strings.NewReader(`{}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := withUserID(e, req, rec, 1)
		c.SetParamNames("id")
		c.SetParamValues("abc")

		call(e, handler.NewTaskHandler(&mocks.TaskSvc{}).Update, c)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestTaskHandler_GetHistory(t *testing.T) {
	e := newEcho()

	t.Run("success", func(t *testing.T) {
		history := []*domain.TaskHistory{
			{ID: 1, FieldName: "status", OldValue: "todo", NewValue: "done"},
		}
		svc := &mocks.TaskSvc{
			GetHistoryFn: func(_ context.Context, _, _ int64) ([]*domain.TaskHistory, error) {
				return history, nil
			},
		}
		req := httptest.NewRequest(http.MethodGet, "/tasks/1/history", nil)
		rec := httptest.NewRecorder()
		c := withUserID(e, req, rec, 1)
		c.SetParamNames("id")
		c.SetParamValues("1")

		call(e, handler.NewTaskHandler(svc).GetHistory, c)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp []interface{}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Len(t, resp, 1)
	})

	t.Run("forbidden", func(t *testing.T) {
		svc := &mocks.TaskSvc{
			GetHistoryFn: func(_ context.Context, _, _ int64) ([]*domain.TaskHistory, error) {
				return nil, domain.ErrForbidden
			},
		}
		req := httptest.NewRequest(http.MethodGet, "/tasks/1/history", nil)
		rec := httptest.NewRecorder()
		c := withUserID(e, req, rec, 1)
		c.SetParamNames("id")
		c.SetParamValues("1")

		call(e, handler.NewTaskHandler(svc).GetHistory, c)

		assert.Equal(t, http.StatusForbidden, rec.Code)
	})
}
