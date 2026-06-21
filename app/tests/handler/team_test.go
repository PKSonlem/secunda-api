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

func withUserID(e *echo.Echo, req *http.Request, rec *httptest.ResponseRecorder, userID int64) echo.Context {
	c := e.NewContext(req, rec)
	c.Set("userID", userID)
	return c
}

func TestTeamHandler_Create(t *testing.T) {
	e := newEcho()

	t.Run("success", func(t *testing.T) {
		svc := &mocks.TeamSvc{
			CreateFn: func(_ context.Context, userID int64, name string) (*domain.Team, error) {
				return &domain.Team{ID: 5, Name: name, CreatedBy: userID}, nil
			},
		}
		req := httptest.NewRequest(http.MethodPost, "/teams",
			strings.NewReader(`{"name":"My Team"}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		call(e, handler.NewTeamHandler(svc).Create, withUserID(e, req, rec, 1))

		assert.Equal(t, http.StatusCreated, rec.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, "My Team", resp["name"])
	})

	t.Run("service_error", func(t *testing.T) {
		svc := &mocks.TeamSvc{
			CreateFn: func(_ context.Context, _ int64, _ string) (*domain.Team, error) {
				return nil, assert.AnError
			},
		}
		req := httptest.NewRequest(http.MethodPost, "/teams",
			strings.NewReader(`{"name":"fail"}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		call(e, handler.NewTeamHandler(svc).Create, withUserID(e, req, rec, 1))

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestTeamHandler_List(t *testing.T) {
	e := newEcho()

	svc := &mocks.TeamSvc{
		ListFn: func(_ context.Context, userID int64) ([]*domain.Team, error) {
			assert.Equal(t, int64(2), userID)
			return []*domain.Team{{ID: 1, Name: "Alpha"}, {ID: 2, Name: "Beta"}}, nil
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/teams", nil)
	rec := httptest.NewRecorder()

	call(e, handler.NewTeamHandler(svc).List, withUserID(e, req, rec, 2))

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp []interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp, 2)
}

func TestTeamHandler_Invite(t *testing.T) {
	e := newEcho()

	t.Run("success", func(t *testing.T) {
		svc := &mocks.TeamSvc{
			InviteFn: func(_ context.Context, _, _ int64, _ string) error { return nil },
		}
		req := httptest.NewRequest(http.MethodPost, "/teams/3/invite",
			strings.NewReader(`{"email":"bob@test.com"}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := withUserID(e, req, rec, 1)
		c.SetParamNames("id")
		c.SetParamValues("3")

		call(e, handler.NewTeamHandler(svc).Invite, c)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("forbidden", func(t *testing.T) {
		svc := &mocks.TeamSvc{
			InviteFn: func(_ context.Context, _, _ int64, _ string) error { return domain.ErrForbidden },
		}
		req := httptest.NewRequest(http.MethodPost, "/teams/3/invite",
			strings.NewReader(`{"email":"bob@test.com"}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := withUserID(e, req, rec, 1)
		c.SetParamNames("id")
		c.SetParamValues("3")

		call(e, handler.NewTeamHandler(svc).Invite, c)

		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("user_not_found", func(t *testing.T) {
		svc := &mocks.TeamSvc{
			InviteFn: func(_ context.Context, _, _ int64, _ string) error { return domain.ErrNotFound },
		}
		req := httptest.NewRequest(http.MethodPost, "/teams/3/invite",
			strings.NewReader(`{"email":"ghost@test.com"}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := withUserID(e, req, rec, 1)
		c.SetParamNames("id")
		c.SetParamValues("3")

		call(e, handler.NewTeamHandler(svc).Invite, c)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("invalid_team_id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/teams/abc/invite",
			strings.NewReader(`{"email":"x@x.com"}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := withUserID(e, req, rec, 1)
		c.SetParamNames("id")
		c.SetParamValues("abc")

		call(e, handler.NewTeamHandler(&mocks.TeamSvc{}).Invite, c)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}
